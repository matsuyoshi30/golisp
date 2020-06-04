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
	TypeCons
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
	fmt.Printf("Cons: %#v\n", cons)
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
	fmt.Printf("\tAtom: %#v\n", atom)
}

// Eval

func (c *Cons) Eval() (*Atom, error) {
	if c == nil {
		return nil, nil
	}

	if c.Cdr == &Nil {
		switch car := c.Car.(type) {
		case *Cons: // Cons{Cons, Nil}
			return car.Eval()
		case *Atom: // Cons{Atom, Nil}
			return car.Eval()
		default:
			return nil, errors.New("invalid type of car")
		}
	} else {
		switch car := c.Car.(type) {
		case *Cons: // Cons{Cons, Cons}
			// FIXME
			if v, err := car.Eval(); err != nil {
				return nil, err
			} else if c.Cdr == &Nil {
				return v, nil
			} else {
				return c.Cdr.(*Cons).Eval()
			}
		case *Atom: // Cons{Atom, _}
			v, err := car.Eval()
			if err != nil {
				return nil, err
			}

			if str, ok := v.Val.(string); ok {
				switch str {
				case "+", "-", "*", "/":
					return c.Cdr.(*Cons).Execute(str)
				default:
					return nil, errors.New("invalid type of car")
				}
			} else {
				if cdr, ok := c.Cdr.(*Cons); ok {
					return cdr.Eval()
				} else {
					return nil, errors.New("should be handle another way")
				}
			}
		default:
			return nil, errors.New("invalid type of car")
		}
	}

	return nil, nil
}

func (c *Cons) Execute(op string) (*Atom, error) {
	switch op {
	case "+":
		return c.evalAdd()
	case "-":
		return c.evalSub()
	case "*":
		return c.evalMul()
	case "/":
		return c.evalDiv()
	}

	return nil, errors.New("unexpected operator string")
}

func evalTerm(i interface{}) (*Atom, error) {
	switch c := i.(type) {
	case *Atom:
		return c.Eval()
	case *Cons:
		// FIXME
		return c.Eval()
	}

	return nil, errors.New("invalid type of argument")
}

func (c *Cons) evalAdd() (*Atom, error) {
	lhs, err := evalTerm(c.Car)
	if err != nil {
		return nil, err
	}

	var rhs int
	for c.Cdr != &Nil {
		temp, err := evalTerm(c.Cdr)
		if err != nil {
			return nil, err
		}

		if temp == &Nil || temp == nil {
			break
		} else if t, ok := temp.Val.(int); ok {
			rhs += t
		} else {
			return nil, errors.New("unexpected value")
		}

		c = c.Cdr.(*Cons)
	}

	return &Atom{Kind: TypeNum, Val: lhs.Val.(int) + rhs}, nil
}

func (c *Cons) evalSub() (*Atom, error) {
	lhs, err := evalTerm(c.Car)
	if err != nil {
		return nil, err
	}

	var rhs int
	for c.Cdr != &Nil {
		temp, err := evalTerm(c.Cdr)
		if err != nil {
			return nil, err
		}

		if temp == &Nil || temp == nil {
			break
		} else if t, ok := temp.Val.(int); ok {
			rhs -= t
		} else {
			return nil, errors.New("unexpected value")
		}

		c = c.Cdr.(*Cons)
	}

	return &Atom{Kind: TypeNum, Val: lhs.Val.(int) - rhs}, nil
}

func (c *Cons) evalMul() (*Atom, error) {
	lhs, err := evalTerm(c.Car)
	if err != nil {
		return nil, err
	}

	var rhs int
	for c.Cdr != &Nil {
		temp, err := evalTerm(c.Cdr)
		if err != nil {
			return nil, err
		}

		if temp == &Nil || temp == nil {
			break
		} else if t, ok := temp.Val.(int); ok {
			rhs *= t
		} else {
			return nil, errors.New("unexpected value")
		}

		c = c.Cdr.(*Cons)
	}

	return &Atom{Kind: TypeNum, Val: lhs.Val.(int) * rhs}, nil
}

func (c *Cons) evalDiv() (*Atom, error) {
	lhs, err := evalTerm(c.Car)
	if err != nil {
		return nil, err
	}

	var rhs int
	for c.Cdr != &Nil {
		temp, err := evalTerm(c.Cdr)
		if err != nil {
			return nil, err
		}

		if temp == &Nil || temp == nil {
			break
		} else if t, ok := temp.Val.(int); ok {
			if t == 0 {
				return nil, errors.New("should not divide by zero")
			} else {
				rhs /= t
			}
		} else {
			return nil, errors.New("unexpected value")
		}

		c = c.Cdr.(*Cons)
	}

	return &Atom{Kind: TypeNum, Val: lhs.Val.(int) / rhs}, nil
}

func (a *Atom) Eval() (*Atom, error) {
	if a != &Nil {
		return a, nil
	}
	return nil, nil
}

// TODO: Print

func (a *Atom) String() string {
	switch a := a.Val.(type) {
	case string:
		return fmt.Sprintf(a)
	case int:
		return fmt.Sprintf("%d", a)
	}

	return ""
}

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

		debugCons(expr)

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
