package requestlogger

import (
	"io"
	"net/http"

	"github.com/rs/zerolog"
)

// Using standard net/http package
func Example_usage() {
	log := zerolog.New(io.Discard)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux := http.NewServeMux()

	loggerHandler := RequestLogger(&log)

	mux.Handle("/", loggerHandler(nextHandler))
}
