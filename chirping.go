package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

const badResponseError = `{
  "error": "Something went wrong"
}`

type chirp struct {
	Body string `json:"body"`
}

type chirpResponse struct {
	Error     string `json:"error"`
	CleanBody string `json:"cleaned_body"`
}

func validateChirpHandler(writer http.ResponseWriter, request *http.Request) {
	chirp := chirp{}
	decoder := json.NewDecoder(request.Body)

	if err := decoder.Decode(&chirp); err != nil {
		log.Printf("Error decoding chirp: %s", err)
		writer.WriteHeader(500)
		response, err := makeChirpResponse("Error decoding chirp", "")
		if err != nil {
			log.Printf("Error writing response: %s", err)
			writer.Write([]byte(badResponseError))
			return
		}
		writer.Write(response)
		return
	}

	if !checkChirpLength(chirp.Body) {
		response, err := makeChirpResponse("Chirp is too long", "")
		if err != nil {
			writer.WriteHeader(500)
			log.Printf("Error writing response: %s", err)
			writer.Write([]byte(badResponseError))
			return
		}
		writer.WriteHeader(400)
		writer.Write(response)
		return
	}

	cleanBody := checkChirpProfanity(chirp.Body)

	response, err := makeChirpResponse("", cleanBody)
	if err != nil {
		writer.WriteHeader(500)
		log.Printf("Error writing response: %s", err)
		writer.Write([]byte(badResponseError))
		return
	}
	writer.WriteHeader(200)
	writer.Write(response)
	return

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

func makeChirpResponse(errors string, cleanBody string) ([]byte, error) {
	responseStruct := chirpResponse{
		Error:     errors,
		CleanBody: cleanBody,
	}
	responseJson, err := json.Marshal(responseStruct)
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
