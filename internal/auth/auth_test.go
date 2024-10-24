package auth

import (
	"net/http"
	"testing"
)

func TestGetPassword(t *testing.T) {
	pwText := "password"
	pwHash, err := HashPassword(pwText)
	if pwText == pwHash || err != nil {
		t.Fatal("Password not hashed.")
	}
}

func TestHashPassword(t *testing.T) {
	pwText := "password"
	pwHash, err := HashPassword(pwText)
	if err != nil {
		t.Fatal("Password not hashed.")
	}
	testing := CheckPasswordHash(pwText, pwHash)
	if testing != nil {
		t.Fatal("Passwords do not match.", testing)
	}
}

func TestGetBearerToken(t *testing.T) {
	header := http.Header{}
	header.Add("Authorization", "Bearer ${jwtTokenMike}")
	token, err := GetBearerToken(header)
	if token != "${jwtTokenMike}" || err != nil {
		t.Fatal("Did not get correct auth token:", token)
	}
}
