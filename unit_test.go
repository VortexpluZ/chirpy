package main

import (
	"testing"
	"time"

	"github.com/google/uuid"

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

func TestValidateJwtToken(t *testing.T) {
	duration, err := time.ParseDuration("3s")
	expectedUserId := uuid.New()
	if err != nil {
		t.Error(err)
	}
	jwtToken, err := auth.MakeJWT(expectedUserId, auth.Secret, duration)
	if err != nil {
		t.Error(err)
		return
	}
	userId, err := auth.ValidateJWT(jwtToken, auth.Secret)
	if err != nil {
		t.Error(err)
		return
	}
	if expectedUserId != userId {
		t.Errorf(`expectedUserId = %s is equal to userId = %s`, expectedUserId, userId)
		return
	}
}

func TestInvalidSecretForJwtToken(t *testing.T) {
	duration, err := time.ParseDuration("3s")
	expectedUserId := uuid.New()
	if err != nil {
		t.Error(err)
		return
	}
	jwtToken, err := auth.MakeJWT(expectedUserId, "wrong signed secret", duration)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = auth.ValidateJWT(jwtToken, auth.Secret)
	if err == nil {
		t.Error("No error was thrown despite wrong signed secret")
		return
	}
}

func TestExpiredJwtToken(t *testing.T) {
	duration, err := time.ParseDuration("0.5s")
	expectedUserId := uuid.New()
	if err != nil {
		t.Error(err)
		return
	}
	jwtToken, err := auth.MakeJWT(expectedUserId, auth.Secret, duration)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)
	_, err = auth.ValidateJWT(jwtToken, auth.Secret)
	if err == nil {
		t.Error("Expected expired token")
		return
	}
}
