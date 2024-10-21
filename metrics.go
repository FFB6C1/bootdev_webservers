package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(writer, request)
	})
	return handler
}

func (cfg *apiConfig) metricsHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/html")
	writer.Write([]byte(cfg.helperGetMetricsString()))
}

func (cfg *apiConfig) helperGetMetricsString() []byte {
	metricsString := fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>
	`, cfg.fileServerHits.Load())
	return []byte(metricsString)
}

func (cfg *apiConfig) resetHandler(writer http.ResponseWriter, _ *http.Request) {
	cfg.fileServerHits.Store(0)
	writer.Write([]byte("file server hits reset."))
}
