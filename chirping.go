package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/FFB6C1/bootdev_webservers/internal/auth"
	"github.com/FFB6C1/bootdev_webservers/internal/database"
	"github.com/google/uuid"
)

const badResponseError = `{
  "error": "Something went wrong"
}`

type chirp struct {
	Body   string    `json:"body"`
	UserId uuid.UUID `json:"user_id"`
}

type chirpResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) postNewChirpHandler(writer http.ResponseWriter, request *http.Request) {
	chirp := chirp{}
	decoder := json.NewDecoder(request.Body)

	if err := decoder.Decode(&chirp); err != nil {
		log.Printf("Error decoding chirp: %s", err)
		handleError("Error decoding chirp", err, 500, writer)
		return
	}

	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		handleError("Couldn't get token from header", err, 401, writer)
		return
	}

	tokenUUID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		handleError("Unauthorized", err, 401, writer)
		return
	}

	if !checkChirpLength(chirp.Body) {
		handleError("Chirp too long", nil, 400, writer)
		return
	}

	cleanBody := checkChirpProfanity(chirp.Body)

	chirpToAdd := database.CreateChirpParams{
		Body:   cleanBody,
		UserID: tokenUUID,
	}

	addedChirp, err := cfg.db.CreateChirp(request.Context(), chirpToAdd)
	if err != nil {
		handleError("Error creating chirp", err, 500, writer)
		return
	}

	response, err := makeChirpJSON(addedChirp)
	if err != nil {
		handleError("Error creating response", err, 500, writer)
		return
	}

	writer.WriteHeader(201)
	writer.Write([]byte(response))
	return
}

func (cfg *apiConfig) getChirpsHandler(writer http.ResponseWriter, request *http.Request) {
	chirps, err := cfg.db.GetChirps(request.Context())
	if err != nil {
		handleError("Failed to get chirps", err, 500, writer)
		return
	}
	writer.WriteHeader(200)
	jsonChirps := []string{}
	for _, chirp := range chirps {
		jsonChirp, err := makeChirpJSON(chirp)
		if err != nil {
			handleError("Failed to marshal chirp into json", err, 500, writer)
			return
		}
		jsonChirps = append(jsonChirps, string(jsonChirp))
	}
	body := "[" + strings.Join(jsonChirps, ", ") + "]"
	writer.Write([]byte(body))
}

func (cfg *apiConfig) getChirpByIdHandler(writer http.ResponseWriter, request *http.Request) {
	id := request.PathValue("chirpID")
	idUUID, err := uuid.Parse(id)
	if err != nil {
		handleError("Bad UUID", err, 400, writer)
		return
	}
	chirp, err := cfg.db.GetChirpById(request.Context(), idUUID)
	if err != nil {
		handleError("Couldn't find chirp", err, 404, writer)
		return
	}

	chirpJSON, err := makeChirpJSON(chirp)
	if err != nil {
		handleError("Couldn't marshal chirp into JSON", err, 500, writer)
		return
	}

	writer.WriteHeader(200)
	writer.Write(chirpJSON)
}

func (cfg *apiConfig) deleteChirpByIdHandler(writer http.ResponseWriter, request *http.Request) {
	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		handleError("Couldn't get token from header", err, 401, writer)
		return
	}

	tokenUUID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		handleError("Unauthorized", err, 401, writer)
		return
	}

	chirpID, err := uuid.Parse(request.PathValue("chirpID"))
	if err != nil {
		handleError("Could not parse chirp ID", err, 400, writer)
		return
	}

	chirp, err := cfg.db.GetChirpById(request.Context(), chirpID)
	if err != nil {
		handleError("Could not find chirp", err, 404, writer)
		return
	}

	if tokenUUID != chirp.UserID {
		handleError("Unauthorized", nil, 403, writer)
		return
	}

	if err := cfg.db.DeleteChirp(request.Context(), chirp.ID); err != nil {
		handleError("Could not delete chirp", err, 500, writer)
		return
	}

	writer.WriteHeader(204)
}

func checkChirpLength(text string) bool {
	return len(text) <= 140
}

func checkChirpProfanity(text string) string {
	words := strings.Split(text, " ")
	badWords := getBadWords()
	for index, word := range words {
		if _, ok := badWords[strings.ToLower(word)]; ok {
			words[index] = "****"
		}
	}
	return strings.Join(words, " ")

}

func makeChirpJSON(chirp database.Chirp) ([]byte, error) {
	chirpJson := chirpResponse{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}
	responseJson, err := json.Marshal(chirpJson)
	if err != nil {
		return []byte(""), err
	}
	return responseJson, nil
}

func getBadWords() map[string]int {
	return map[string]int{
		"kerfuffle": 0,
		"sharbert":  0,
		"fornax":    0,
	}
}
