package main

import (
	"testing"
)

func TestProfaneRewriteDouble(t *testing.T) {
	text := "I really need a kerfuffle to go to bed sooner, Fornax !"
	result := profaneRewrite(text)
	expected := "I really need a **** to go to bed sooner, **** !"
	if result != expected {
		t.Errorf(`result = %s, expected = %s`, result, expected)
	}
}

func TestProfaneRewriteSimple(t *testing.T) {
	text := "I really need a fornax to go to bed sooner"
	result := profaneRewrite(text)
	expected := "I really need a **** to go to bed sooner"
	if result != expected {
		t.Errorf(`result = %s, expected = %s`, result, expected)
	}
}
