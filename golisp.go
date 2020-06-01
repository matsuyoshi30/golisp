package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Read

func Read(str string) (*Cons, error) {
	tokenizer := NewTokenizer(strings.TrimSuffix(str, " "))
	tokens, err := tokenizer.Tokenize()
	if err != nil {
		return nil, err
	}

	parser := NewParser(tokens)
	return parser.Parse()
}

/// tokenize

type TokenType int

const (
	TK_NUM TokenType = iota
	TK_OP
	TK_LPAREN
	TK_RPAREN
	TK_EOF
)

type Token struct {
	Kind TokenType
	Val  int // use if TK_NUM
	Str  string
}

type Tokenizer struct {
	input string
	pos   int
}

func NewTokenizer(str string) *Tokenizer {
	return &Tokenizer{input: str, pos: 0}
}

func (t *Tokenizer) next() {
	t.pos++
}

func (t *Tokenizer) Tokenize() ([]*Token, error) {
	tokens := make([]*Token, 0)

	for t.pos < len(t.input) {
		t.skipWhiteSpace()

		switch t.input[t.pos] {
		case '+':
			tokens = append(tokens, &Token{Kind: TK_OP, Str: "+"})
		case '-':
			tokens = append(tokens, &Token{Kind: TK_OP, Str: "-"})
		case '*':
			tokens = append(tokens, &Token{Kind: TK_OP, Str: "*"})
		case '/':
			tokens = append(tokens, &Token{Kind: TK_OP, Str: "/"})
		case '(':
			tokens = append(tokens, &Token{Kind: TK_LPAREN, Str: "("})
		case ')':
			tokens = append(tokens, &Token{Kind: TK_RPAREN, Str: ")"})
		default:
			if t.isDigit() {
				start := t.pos
				for t.pos < len(t.input) && t.isDigit() {
					t.next()
				}
				val, err := strconv.Atoi(t.input[start:t.pos])
				if err != nil {
					return nil, err
				}
				t.pos--
				tokens = append(tokens, &Token{Kind: TK_NUM, Val: val})
			}
		}
		t.next()
	}
	tokens = append(tokens, &Token{Kind: TK_EOF})

	return tokens, nil
}

func (t *Tokenizer) skipWhiteSpace() {
	for t.input[t.pos] == ' ' {
		t.next()
	}
}

func (t *Tokenizer) isDigit() bool {
	return '0' <= t.input[t.pos] && t.input[t.pos] <= '9'
}

/// parse

type Parser struct {
	tokens []*Token
	pos    int
}

func NewParser(toks []*Token) *Parser {
	return &Parser{tokens: toks}
}

func (p *Parser) Pos() int {
	return p.pos
}

func (p *Parser) current() *Token {
	return p.tokens[p.Pos()]
}

func (p *Parser) peek() *Token {
	return p.tokens[p.Pos()+1]
}

func (p *Parser) next() {
	p.pos++
}

type Cons struct {
	Car interface{} // Contents of the Address part of the Register
	Cdr interface{} // Contents of the Decrement part of the Register
}

type AtomType int

const (
	TypeNil AtomType = iota
	TypeNum
	TypeOp
)

type Atom struct {
	Kind AtomType
	Val  interface{}
}

var Nil = Atom{Kind: TypeNil, Val: nil}

// parse = <nil> | node*
func (p *Parser) Parse() (*Cons, error) {
	var cur, cons *Cons

	for {
		if p.current().Kind == TK_EOF || p.current().Kind == TK_RPAREN {
			break
		}

		if cur == nil {
			cons = &Cons{&Nil, &Nil}
			cur = cons
		} else {
			p := cur
			cur = &Cons{&Nil, &Nil}
			p.Cdr = cur
		}

		switch p.current().Kind {
		case TK_NUM:
			cur.Car = &Atom{Kind: TypeNum, Val: p.current().Val}
			p.next()
		case TK_OP:
			cur.Car = &Atom{Kind: TypeOp, Val: p.current().Str}
			p.next()
		case TK_LPAREN:
			p.next()
			if p.current().Kind == TK_RPAREN {
				cur.Car = &Nil
			} else {
				nested, err := p.Parse()
				if err != nil {
					return nil, err
				}
				cur.Car = nested
			}
			p.next()
		}
	}

	return cons, nil
}

func debugCons(cons *Cons) {
	switch cons.Car.(type) {
	case *Cons:
		debugCons(cons.Car.(*Cons))
	case *Atom:
		debugAtom(cons.Car.(*Atom))
	}

	if cons.Cdr != &Nil {
		switch cons.Cdr.(type) {
		case *Cons:
			debugCons(cons.Cdr.(*Cons))
		case *Atom:
			debugAtom(cons.Cdr.(*Atom))
		}
	}
}

func debugAtom(atom *Atom) {
	fmt.Printf("%#v\n", atom)
}

// Eval

func (c *Cons) Eval() (string, error) {
	if c == nil {
		return "", nil
	}

	// _, Nil
	if c.Cdr == &Nil {
		switch car := c.Car.(type) {
		case *Cons:
			return car.Eval()
		case *Atom:
			return car.Eval()
		default:
			return "", errors.New("invalid type of car")
		}
	} else { // _, Cons
		switch car := c.Car.(type) {
		case *Cons:
			// TODO
		case *Atom:
			v, err := car.Eval()
			if err != nil {
				return "", err
			}

			switch v {
			case "+":
				val, err := c.Cdr.(*Cons).evalAdd()
				if err != nil {
					return "", err
				}
				return strconv.Itoa(val), nil
			case "-":
				val, err := c.Cdr.(*Cons).evalSub()
				if err != nil {
					return "", err
				}
				return strconv.Itoa(val), nil
			case "*":
				val, err := c.Cdr.(*Cons).evalMul()
				if err != nil {
					return "", err
				}
				return strconv.Itoa(val), nil
			case "/":
				val, err := c.Cdr.(*Cons).evalDiv()
				if err != nil {
					return "", err
				}
				return strconv.Itoa(val), nil
			default:
				return v, nil
			}
		default:
			return "", errors.New("invalid type of car")
		}
	}

	return "", nil
}

func evalTerm(i interface{}) (int, error) {
	var val int

	switch c := i.(type) {
	case *Atom:
		v, err := c.Eval()
		if err != nil {
			return 0, err
		}
		val, err = strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
	case *Cons:
		v, err := c.Eval()
		if err != nil {
			return 0, err
		}
		val, err = strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
	}

	return val, nil
}

func (c *Cons) evalAdd() (int, error) {
	ret, err := evalTerm(c.Car)
	if err != nil {
		return 0, err
	}

	for c.Cdr != &Nil {
		add, err := evalTerm(c.Cdr)
		if err != nil {
			return 0, err
		}
		ret += add
		c = c.Cdr.(*Cons)
	}

	return ret, nil
}

func (c *Cons) evalSub() (int, error) {
	ret, err := evalTerm(c.Car)
	if err != nil {
		return 0, err
	}

	for c.Cdr != &Nil {
		sub, err := evalTerm(c.Cdr)
		if err != nil {
			return 0, err
		}
		ret -= sub
		c = c.Cdr.(*Cons)
	}

	return ret, nil
}

func (c *Cons) evalMul() (int, error) {
	ret, err := evalTerm(c.Car)
	if err != nil {
		return 0, err
	}

	for c.Cdr != &Nil {
		mul, err := evalTerm(c.Cdr)
		if err != nil {
			return 0, err
		}
		ret *= mul
		c = c.Cdr.(*Cons)
	}

	return ret, nil
}

func (c *Cons) evalDiv() (int, error) {
	ret, err := evalTerm(c.Car)
	if err != nil {
		return 0, err
	}

	for c.Cdr != &Nil {
		div, err := evalTerm(c.Cdr)
		if err != nil {
			return 0, err
		}
		if div == 0 {
			return 0, errors.New("could not divide by zero")
		}

		ret /= div
		c = c.Cdr.(*Cons)
	}

	return ret, nil
}

func (a *Atom) Eval() (string, error) {
	if a == &Nil {
		return "", nil
	}

	if a.Kind == TypeNum {
		return strconv.Itoa(a.Val.(int)), nil
	}
	if a.Kind == TypeOp {
		return a.Val.(string), nil
	}

	return "", nil
}

// TODO: Print

// Loop

func loop() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("user input> ")
		line, _, err := reader.ReadLine()
		if err != nil {
			fmt.Println(err)
			continue
		}

		expr, err := Read(string(line))
		if err != nil {
			fmt.Println(err)
			continue
		}

		// debugCons(expr)

		out, err := expr.Eval()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(out)
	}
}

func main() {
	loop()
}
