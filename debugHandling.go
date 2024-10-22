package main

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) resetHandler(writer http.ResponseWriter, request *http.Request) {
	if cfg.platform != "dev" {
		handleError("Forbidden.", nil, 403, writer)
		return
	}
	if err := cfg.db.ResetUsers(request.Context()); err != nil {
		handleError("Could not reset users database", err, 500, writer)
		return
	}
	cfg.fileServerHits.Store(0)
	log.Println("Users database reset. Fileserver hits reset.")
	writer.WriteHeader(200)
	writer.Write([]byte("Users database reset. Fileserver hits reset."))
}
