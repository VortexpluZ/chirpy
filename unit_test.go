package main

import (
	"testing"

	"github.com/VortexpluZ/chirpy/internal/auth"
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

func TestCheckPasswordHash(t *testing.T) {
	pass := "thisismypass"
	hash, _ := auth.HashPassword(pass)

	match, _ := auth.CheckPasswordHash(pass, hash)

	if !match {
		t.Errorf(`pass = %s is different than hash = %s`, pass, hash)
	}
}

func TestFailedCheckPasswordHash(t *testing.T) {
	pass := "thisismypass"
	notPass := "thisisnotmypass"
	hash, _ := auth.HashPassword(pass)

	match, _ := auth.CheckPasswordHash(notPass, hash)

	if match {
		t.Errorf(`pass = %s is equal to notPass = %s`, pass, notPass)
	}
}
