package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/FFB6C1/bootdev_webservers/internal/database"
	"github.com/google/uuid"
)

type newUserRequest struct {
	Email string `json:"email"`
}

type newUserResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) newUserHandler(writer http.ResponseWriter, request *http.Request) {
	userRequest := newUserRequest{}
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&userRequest); err != nil {
		log.Printf("Error decoding user request: %s", err)
		handleError("Could not decode user request", err, 400, writer)
		return
	}

	newUser, err := cfg.db.CreateUser(request.Context(), userRequest.Email)
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
