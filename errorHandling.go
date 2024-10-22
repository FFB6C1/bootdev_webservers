package main

import (
	"fmt"
	"net/http"
)

func handleError(text string, err error, code int, writer http.ResponseWriter) {
	writer.WriteHeader(code)
	if err == nil {
		response := fmt.Sprintf(` {
		"error": %s
		}`, text)
		writer.Write([]byte(response))
		return
	}
	response := fmt.Sprintf(` {
	"error": %s: %v
	}`, text, err)
	writer.Write([]byte(response))
}
