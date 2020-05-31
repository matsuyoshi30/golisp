package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Read

func Read(str string) (*Node, error) {
	tokenizer := NewTokenizer(str)
	tokens, err := tokenizer.Tokenize()
	if err != nil {
		return nil, err
	}

	parser := NewParser(tokens)
	return parser.Parse()
}

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

type Parser struct {
	tokens []*Token
	pos    int
}

func NewParser(toks []*Token) *Parser {
	return &Parser{tokens: toks}
}

func (p *Parser) current() *Token {
	return p.tokens[p.pos]
}

func (p *Parser) next() {
	p.pos++
}

type NodeKind int

const (
	ND_NUM NodeKind = iota
	ND_OP
)

type Node struct {
	Kind NodeKind
	Val  interface{}
	Next *Node
}

// parse = <nil> | node*
func (p *Parser) Parse() (*Node, error) {
	head := &Node{}
	cur := head

	for p.current().Kind != TK_EOF {
		node := p.readNode()
		if node != nil {
			cur.Next = node
			cur = cur.Next
		}
		p.next()
	}

	return head.Next, nil
}

// readNode = TK_RPALEN | TK_LPAREN | TK_NUM | TK_OP
func (p *Parser) readNode() *Node {
	switch p.current().Kind {
	case TK_RPAREN:
		return nil
	case TK_LPAREN:
		p.next()
		return p.readNode()
	case TK_NUM:
		return &Node{Kind: ND_NUM, Val: p.current().Val}
	case TK_OP:
		return &Node{Kind: ND_OP, Val: p.current().Str}
	}

	return nil
}

// Eval

func Eval(node *Node) (*Node, error) {
	var result *Node

	if node == nil {
		return nil, nil
	}

	switch node.Kind {
	case ND_NUM:
		return node, nil
	case ND_OP:
		ret, err := calculate(node)
		if err != nil {
			return nil, err
		}
		result = ret
	default:
		result = node
	}
	return result, nil
}

func calculate(node *Node) (*Node, error) {
	op := node.Val.(string)
	node = node.Next
	if node == nil || node.Next == nil {
		return nil, errors.New("invalid number of arguments")
	}

	cur, err := Eval(node)
	if err != nil {
		return nil, err
	}
	if cur.Kind != ND_NUM {
		return nil, errors.New("expected number")
	}

	next, err := Eval(node.Next)
	if err != nil {
		return nil, err
	}

	switch op {
	case "+":
		return &Node{Val: cur.Val.(int) + next.Val.(int)}, nil
	case "-":
		return &Node{Val: cur.Val.(int) - next.Val.(int)}, nil
	case "*":
		return &Node{Val: cur.Val.(int) * next.Val.(int)}, nil
	case "/":
		return &Node{Val: cur.Val.(int) / next.Val.(int)}, nil
	}

	return nil, errors.New("invalid calculation")
}

// Print

func Print(node *Node) {
	printNode(node)
	fmt.Println()
}

func printNode(node *Node) {
	if node == nil {
		fmt.Printf("<nil>")
	} else {
		fmt.Printf("%v ", node.Val)
		if node.Next != nil {
			printNode(node.Next)
		}
	}
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

		result, err := Eval(expr)
		if err != nil {
			fmt.Println(err)
			continue
		}

		Print(result)
	}
}

func main() {
	loop()
}
