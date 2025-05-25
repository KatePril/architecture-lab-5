package main

import (
	"encoding/json"
	"flag"
	"github.com/KatePril/architecture-lab-5/datastore"
	"github.com/KatePril/architecture-lab-5/httptools"
	"github.com/KatePril/architecture-lab-5/signal"
	"net/http"
	"strings"
)

var port = flag.Int("port", 8081, "server port")

func main() {
	db, err := datastore.Open("db")
	if err != nil {
		// handle the error properly
		panic(err)
	}
	h := new(http.ServeMux)
	h.HandleFunc("/db/", func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/db/")
		if key == "" {
			http.Error(w, "Key is required", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			value, getError := db.Get(key)
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

			db.Put(key, body.Value)
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	server := httptools.CreateServer(*port, h)
	server.Start()
	signal.WaitForTerminationSignal()
}
