package cryptouseruuid

import (
	"io"
	"net/http"

	"github.com/rs/zerolog"
)

// Using standard net/http package
func Example_usage() {
	zl := zerolog.New(io.Discard).With()
	log := zl.Logger()

	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	}

	mux := http.NewServeMux()

	finalHandler := http.HandlerFunc(handler)
	uuidHandler := UserUUID(&log)

	mux.Handle("/", uuidHandler(finalHandler))
}
