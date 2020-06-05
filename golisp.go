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
	TypeNum AtomType = iota
	TypeOp
)

type Atom struct {
	Kind AtomType
	Val  interface{}
}

func (p *Parser) Parse() (*Cons, error) {
	var cur, cons *Cons

	for {
		if p.current().Kind == TK_EOF || p.current().Kind == TK_RPAREN {
			break
		}

		if cur == nil {
			cons = &Cons{nil, nil}
			cur = cons
		} else {
			p := cur
			cur = &Cons{nil, nil}
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
			nested, err := p.Parse()
			if err != nil {
				return nil, err
			}
			cur.Car = nested
			p.next()
		}
	}

	return cons, nil
}

func debugCons(cons *Cons) {
	fmt.Printf("Cons: %#v\n", cons)
	switch car := cons.Car.(type) {
	case *Cons:
		debugCons(car)
	case *Atom:
		debugAtom(car)
	}

	if cons.Cdr != nil {
		switch cdr := cons.Cdr.(type) {
		case *Cons:
			debugCons(cdr)
		case *Atom:
			debugAtom(cdr)
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
	} else {
		switch car := c.Car.(type) {
		case *Cons:
			if v, err := car.Eval(); err != nil {
				return nil, err
			} else if c.Cdr == nil {
				return v, nil
			} else {
				return c.Cdr.(*Cons).Eval()
			}
		case *Atom:
			if v, err := car.Eval(); err != nil {
				return nil, err
			} else if c.Cdr == nil {
				return v, nil
			} else {
				if str, ok := v.Val.(string); ok {
					switch str {
					case "+", "-", "*", "/":
						return c.Cdr.(*Cons).Execute(str)
					default:
						return nil, errors.New("invalid string")
					}
				} else {
					return c.Cdr.(*Cons).Eval()
				}
			}
		default:
			return nil, errors.New("invalid type of car")
		}
	}
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

func (c *Cons) evalAdd() (*Atom, error) {
	lhs, err := evalTerm("+", c.Car)
	if err != nil {
		return nil, err
	}

	rhs, err := evalTerm("+", c.Cdr)
	if err != nil {
		return nil, err
	}

	if rv, ok := rhs.Val.(int); ok {
		return &Atom{Kind: TypeNum, Val: lhs.Val.(int) + rv}, nil
	}

	return nil, errors.New("should not reach")
}

func (c *Cons) evalSub() (*Atom, error) {
	lhs, err := evalTerm("-", c.Car)
	if err != nil {
		return nil, err
	}

	rhs, err := evalTerm("-", c.Cdr)
	if err != nil {
		return nil, err
	}

	if rv, ok := rhs.Val.(int); ok {
		return &Atom{Kind: TypeNum, Val: lhs.Val.(int) - rv}, nil
	}

	return nil, errors.New("should not reach")
}

func (c *Cons) evalMul() (*Atom, error) {
	lhs, err := evalTerm("*", c.Car)
	if err != nil {
		return nil, err
	}

	rhs, err := evalTerm("*", c.Cdr)
	if err != nil {
		return nil, err
	}

	if rv, ok := rhs.Val.(int); ok {
		return &Atom{Kind: TypeNum, Val: lhs.Val.(int) * rv}, nil
	}

	return nil, errors.New("should not reach")
}

func (c *Cons) evalDiv() (*Atom, error) {
	lhs, err := evalTerm("/", c.Car)
	if err != nil {
		return nil, err
	}

	rhs, err := evalTerm("/", c.Cdr)
	if err != nil {
		return nil, err
	}

	if rv, ok := rhs.Val.(int); ok {
		if rv == 0 {
			return nil, errors.New("should not divide by zero")
		}
		return &Atom{Kind: TypeNum, Val: lhs.Val.(int) / rv}, nil
	}

	return nil, errors.New("should not reach")
}

func evalTerm(op string, i interface{}) (*Atom, error) {
	switch c := i.(type) {
	case *Atom:
		return c.Eval()
	case *Cons:
		var lhs, rhs *Atom
		var err error
		if car, ok := c.Car.(*Cons); ok {
			lhs, err = car.Eval()
			if err != nil {
				return nil, err
			}
		} else {
			if _, ok := c.Car.(*Atom).Val.(string); ok {
				return c.Eval()
			} else {
				lhs, err = evalTerm(op, c.Car)
				if err != nil {
					return nil, err
				}
			}
		}

		if c.Cdr == nil {
			return &Atom{Kind: TypeNum, Val: lhs.Val.(int)}, nil
		}

		switch op {
		case "+":
			if cdr, ok := c.Cdr.(*Cons); ok {
				rhs, err = cdr.Eval()
				if err != nil {
					return nil, err
				}
			} else {
				rhs, err = evalTerm(op, c.Cdr)
				if err != nil {
					return nil, err
				}
			}
			return &Atom{Kind: TypeNum, Val: lhs.Val.(int) + rhs.Val.(int)}, nil
		case "-":
			if cdr, ok := c.Cdr.(*Cons); ok {
				rhs, err = cdr.Eval()
				if err != nil {
					return nil, err
				}
			} else {
				rhs, err = evalTerm(op, c.Cdr)
				if err != nil {
					return nil, err
				}
			}
			return &Atom{Kind: TypeNum, Val: lhs.Val.(int) - rhs.Val.(int)}, nil
		case "*":
			if cdr, ok := c.Cdr.(*Cons); ok {
				rhs, err = cdr.Eval()
				if err != nil {
					return nil, err
				}
			} else {
				rhs, err = evalTerm(op, c.Cdr)
				if err != nil {
					return nil, err
				}
			}
			return &Atom{Kind: TypeNum, Val: lhs.Val.(int) * rhs.Val.(int)}, nil
		case "/":
			if cdr, ok := c.Cdr.(*Cons); ok {
				rhs, err = cdr.Eval()
				if err != nil {
					return nil, err
				}
			} else {
				rhs, err = evalTerm(op, c.Cdr)
				if err != nil {
					return nil, err
				}
			}
			if rhs.Val.(int) == 0 {
				return nil, errors.New("should not divide by zero")
			}
			return &Atom{Kind: TypeNum, Val: lhs.Val.(int) / rhs.Val.(int)}, nil
		}
	}
	return nil, errors.New("invalid type of argument")
}

func (a *Atom) Eval() (*Atom, error) {
	return a, nil
}

// Print

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
