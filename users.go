package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/FFB6C1/bootdev_webservers/internal/auth"
	"github.com/FFB6C1/bootdev_webservers/internal/database"
	"github.com/google/uuid"
)

type userRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type newUserResponse struct {
	Id           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (cfg *apiConfig) newUserHandler(writer http.ResponseWriter, request *http.Request) {
	userRequest := userRequest{}
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&userRequest); err != nil {
		log.Printf("Error decoding user request: %s", err)
		handleError("Could not decode user request", err, 400, writer)
		return
	}

	hashed, err := auth.HashPassword(userRequest.Password)
	if err != nil {
		handleError("Could not hash password", err, 500, writer)
	}

	userParams := database.CreateUserParams{
		Email:          userRequest.Email,
		HashedPassword: hashed,
	}

	newUser, err := cfg.db.CreateUser(request.Context(), userParams)
	if err != nil {
		log.Printf("Error creating new user record: %s", err)
		handleError("Could not create new user record", err, 500, writer)
		return
	}

	response, err := makeUserResponse(newUser)
	if err != nil {
		log.Printf("User created, could not respond: %s", err)
		handleError("Created user record but cannot respond", err, 500, writer)
		return
	}

	writer.WriteHeader(201)
	writer.Write(response)
}

func (cfg *apiConfig) loginHandler(writer http.ResponseWriter, request *http.Request) {
	userRequest := userRequest{}
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&userRequest); err != nil {
		log.Printf("Error decoding user request: %s", err)
		handleError("Could not decode user request", err, 400, writer)
		return
	}

	user, err := cfg.db.GetUserByEmail(context.Background(), userRequest.Email)
	if err != nil {
		handleError("incorrect email or password", err, 401, writer)
		return
	}

	if err := auth.CheckPasswordHash(userRequest.Password, user.HashedPassword); err != nil {
		handleError("incorrect email or password", err, 401, writer)
		return
	}

	expirationTime := time.Duration(1) * time.Hour

	token, err := auth.MakeJWT(user.ID, cfg.secret, expirationTime)
	if err != nil {
		handleError("Could not make auth token", err, 500, writer)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		handleError("Could not make auth token", err, 500, writer)
		return
	}

	refreshTokenParams := database.AddRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	}
	if err := cfg.db.AddRefreshToken(request.Context(), refreshTokenParams); err != nil {
		handleError("could not make refresh token", err, 500, writer)
		return
	}

	responseJSON, err := makeUserResponseWithToken(user, token, refreshToken)
	if err != nil {
		handleError("could not make response", err, 500, writer)
		return
	}

	writer.WriteHeader(200)
	writer.Write(responseJSON)
}

func (cfg *apiConfig) refreshHandler(writer http.ResponseWriter, request *http.Request) {
	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		handleError("Couldn't get token from header", nil, 401, writer)
		return
	}
	tokenFull, err := cfg.db.GetToken(request.Context(), token)
	if err != nil {
		handleError("Token does not exist", err, 401, writer)
		return
	}
	if err := checkTokenValid(tokenFull); err != nil {
		handleError("Token invalid", err, 401, writer)
		return
	}

	newToken, err := auth.MakeJWT(tokenFull.UserID, cfg.secret, time.Hour)
	if err != nil {
		handleError("Could not make token", err, 500, writer)
		return
	}

	response, err := makeTokenResponse(newToken)
	if err != nil {
		handleError("Could not make response", err, 500, writer)
		return
	}
	writer.WriteHeader(200)
	writer.Write(response)
}

func (cfg *apiConfig) revokeHandler(writer http.ResponseWriter, request *http.Request) {
	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		handleError("Could not get token from header", err, 401, writer)
	}
	if err := cfg.db.RevokeToken(request.Context(), token); err != nil {
		handleError("Could not revoke token", err, 401, writer)
	}
	writer.WriteHeader(204)
}

func makeUserResponse(user database.User) ([]byte, error) {
	responseStruct := newUserResponse{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	responseJson, err := json.Marshal(responseStruct)
	if err != nil {
		return []byte(""), err
	}
	return responseJson, nil
}

func makeUserResponseWithToken(user database.User, token string, refreshToken string) ([]byte, error) {
	responseStruct := newUserResponse{
		Id:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}
	responseJson, err := json.Marshal(responseStruct)
	if err != nil {
		return []byte(""), err
	}
	return responseJson, nil
}

func makeTokenResponse(token string) ([]byte, error) {
	responseStruct := newUserResponse{
		Token: token,
	}
	responseJson, err := json.Marshal(responseStruct)
	if err != nil {
		return []byte(""), err
	}
	return responseJson, nil
}

func checkTokenValid(token database.RefreshToken) error {
	if time.Now().After(token.ExpiresAt) {
		return fmt.Errorf("Token Expired")
	}
	if token.RevokedAt.Valid {
		return fmt.Errorf("Token Revoked")
	}
	return nil
}
