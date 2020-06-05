package main

import (
	"reflect"
	"testing"
)

func TestEval(t *testing.T) {
	testcases := []struct {
		input    string
		expected *Atom
	}{
		{"", nil},
		{"(+ 1 2)", &Atom{Kind: TypeNum, Val: 3}},
		{"(+ 1 2 3)", &Atom{Kind: TypeNum, Val: 6}},
		{"(+ 1 2 (+ 3 4))", &Atom{Kind: TypeNum, Val: 10}},
		{"(+ 1 (+ 2 3 4) 5)", &Atom{Kind: TypeNum, Val: 15}},
		{"(+ (+ 1 2 3) 4 5)", &Atom{Kind: TypeNum, Val: 15}},
		{"(- 1 2)", &Atom{Kind: TypeNum, Val: -1}},
		{"(+ 1 2 (* 3 4))", &Atom{Kind: TypeNum, Val: 15}},
		{"(+ 1 (/ 8 2 ) 5)", &Atom{Kind: TypeNum, Val: 10}},
		{"(* (+ 1 2 3) 4 5)", &Atom{Kind: TypeNum, Val: 120}},
		{"(+ 1 (/ 8 2 2) 3)", &Atom{Kind: TypeNum, Val: 6}},
	}

	for _, tt := range testcases {
		expr, err := Read(tt.input)
		if err != nil {
			t.Fatalf("unexpected err: %v\n", err)
		}

		actual, err := expr.Eval()
		if err != nil {
			t.Fatalf("unexpected err: %v\n", err)
		}

		if !reflect.DeepEqual(tt.expected, actual) {
			t.Fatalf("expected %#v but got %#v\n", tt.expected, actual)
		}
	}
}
