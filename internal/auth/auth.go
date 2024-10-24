package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	issuedTime := jwt.NumericDate{
		Time: time.Now(),
	}
	expiresTime := jwt.NumericDate{
		Time: issuedTime.Time.Add(expiresIn),
	}
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  &issuedTime,
		ExpiresAt: &expiresTime,
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("unexpected claims type")
	}
	userString := claims.Subject
	userUuid, err := uuid.Parse(userString)
	if err != nil {
		return uuid.Nil, err
	}
	return userUuid, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	prefix := "Bearer "
	token, ok := strings.CutPrefix(authHeader, prefix)
	if !ok {
		return "", fmt.Errorf("Bad authorization header.")
	}
	return token, nil
}

func MakeRefreshToken() (string, error) {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}
