package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/FFB6C1/bootdev_webservers/internal/auth"
	"github.com/FFB6C1/bootdev_webservers/internal/database"
	"github.com/google/uuid"
)

type userRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type newUserResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
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

	var expirationTime time.Duration
	if userRequest.ExpiresInSeconds == 0 {
		expirationTime = time.Duration(1) * time.Hour
	} else {
		expirationTime = time.Duration(userRequest.ExpiresInSeconds) * time.Second
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, expirationTime)
	if err != nil {
		handleError("Could not make auth token", err, 500, writer)
		return
	}

	responseJSON, err := makeUserResponseWithToken(user, token)
	if err != nil {
		handleError("could not make response", err, 500, writer)
		return
	}

	writer.WriteHeader(200)
	writer.Write(responseJSON)

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

func makeUserResponseWithToken(user database.User, token string) ([]byte, error) {
	responseStruct := newUserResponse{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	}
	responseJson, err := json.Marshal(responseStruct)
	if err != nil {
		return []byte(""), err
	}
	return responseJson, nil
}
