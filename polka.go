package main

import (
	"encoding/json"
	"net/http"

	"github.com/FFB6C1/bootdev_webservers/internal/auth"
	"github.com/google/uuid"
)

type polkaData struct {
	UserId string `json:"user_id"`
}

type polkaRequest struct {
	Event string    `json:"event"`
	Data  polkaData `json:"data"`
}

func (cfg *apiConfig) recievePolkaEvent(writer http.ResponseWriter, request *http.Request) {
	polkaRequest := polkaRequest{}
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&polkaRequest); err != nil {
		handleError("Could not decode header", err, 400, writer)
		return
	}

	apiKey, err := auth.GetAPIKey(request.Header)
	if err != nil {
		handleError("Could not get auth key", err, 401, writer)
		return
	}

	if apiKey != cfg.polkaKey {
		handleError("Unauthorized", nil, 401, writer)
	}

	if polkaRequest.Event != "user.upgraded" {
		writer.WriteHeader(204)
		return
	}

	userId, err := uuid.Parse(polkaRequest.Data.UserId)
	if err != nil {
		handleError("Could not parse userId", err, 400, writer)
		return
	}

	if err := cfg.db.UpgradeByID(request.Context(), userId); err != nil {
		handleError("Could not find and upgrade user", err, 404, writer)
		return
	}

	writer.WriteHeader(204)
}
