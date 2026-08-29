package main

import (
	"testing"

	"github.com/VortexpluZ/chirpy/internal/auth"
)

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
