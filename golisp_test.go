package main

import (
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {
	testcases := []struct {
		input    string
		expected *Cons
	}{
		{"", nil},
	}

	for _, tt := range testcases {
		actual, err := Read(tt.input)
		if err != nil {
			t.Fatalf("unexpected err: %v\n", err)
		}

		if !reflect.DeepEqual(tt.expected, actual) {
			t.Fatalf("expected %#v but got %#v\n", tt.expected, actual)
		}
	}
}

func TestEval(t *testing.T) {
	testcases := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"(+ 1 1)", "2"},
		{"(+ 2 (* 3 4))", "14"},
		{"(- 8 (/ 6 (+ 1 1)))", "5"},
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
