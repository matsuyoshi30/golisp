package main

import (
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {
	testcases := []struct {
		input    string
		expected *Node
	}{
		{"", nil},
		{"1", &Node{Kind: ND_NUM, Val: 1}},
		{"(2 2)", &Node{Kind: ND_NUM, Val: 2,
			Next: &Node{Kind: ND_NUM, Val: 2}}},
		{"(+ 3 3)", &Node{Kind: ND_OP, Val: "+",
			Next: &Node{Kind: ND_NUM, Val: 3,
				Next: &Node{Kind: ND_NUM, Val: 3}}}},
		{"(+ 4 (- 4 4))", &Node{Kind: ND_OP, Val: "+",
			Next: &Node{Kind: ND_NUM, Val: 4,
				Next: &Node{Kind: ND_OP, Val: "-",
					Next: &Node{Kind: ND_NUM, Val: 4,
						Next: &Node{Kind: ND_NUM, Val: 4}}}}}},
	}

	for _, tt := range testcases {
		actual, err := Read(tt.input)
		if err != nil {
			t.Fatalf("unexpected err: %v\n", err)
		}

		if !reflect.DeepEqual(tt.expected, actual) {
			t.Fatalf("expected %v but got %v\n", tt.expected, actual)
		}
	}
}

func TestEval(t *testing.T) {
	testcases := []struct {
		input    string
		expected *Node
	}{
		{"", nil},
		{"1", &Node{Val: 1}},
		{"(2 2)", &Node{Kind: ND_NUM, Val: 2, Next: &Node{Kind: ND_NUM, Val: 2}}},
		{"(+ 3 3)", &Node{Val: 6}},
		{"(+ 7 (- 6 5))", &Node{Val: 8}},
		{"(+ 4 (- 9 (* 2 (/ 6 3))))", &Node{Val: 9}},
	}

	for _, tt := range testcases {
		expr, err := Read(tt.input)
		if err != nil {
			t.Fatalf("unexpected err: %v\n", err)
		}

		actual, err := Eval(expr)
		if err != nil {
			t.Fatalf("unexpected err: %v\n", err)
		}

		if !reflect.DeepEqual(tt.expected, actual) {
			t.Fatalf("expected %v but got %v\n", tt.expected, actual)
		}
	}
}
