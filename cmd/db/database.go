package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/KatePril/architecture-lab-5/datastore"
	"github.com/KatePril/architecture-lab-5/httptools"
	"github.com/KatePril/architecture-lab-5/safestorage"
	"github.com/KatePril/architecture-lab-5/signal"
	"log"
	"net/http"
	"os"
	"strings"
)

var port = flag.Int("port", 8091, "server port")

const confHealthFailure = "CONF_HEALTH_FAILURE"

func main() {
	db, err := datastore.Open("db1/")
	if err != nil {
		fmt.Println("Error opening database: ", err)
		os.Exit(1)
	}
	ss := safestorage.Init(db)

	h := new(http.ServeMux)

	h.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("content-type", "text/plain")
		if failConfig := os.Getenv(confHealthFailure); failConfig == "true" {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte("FAILURE"))
		} else {
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write([]byte("OK"))
		}
	})

	h.HandleFunc("/db/", func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/db/")
		if key == "" {
			http.Error(w, "Key is required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			value, getError := ss.Get(key)
			if getError != nil {
				http.Error(w, "Key not found", http.StatusNotFound)
				return
			}
			response := map[string]any{
				"key":   key,
				"value": value,
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		case http.MethodPost:
			var body struct {
				Value string `json:"value"`
			}

			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "Invalid JSON body", http.StatusBadRequest)
				return
			}

			ss.Put(key, body.Value)
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	server := httptools.CreateServer(*port, h)
	server.Start()
	log.Printf("Starting server on port %d...", *port)
	signal.WaitForTerminationSignal()
}
