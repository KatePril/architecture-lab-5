package main

import (
	"encoding/json"
	"flag"
	"github.com/KatePril/architecture-lab-5/datastore"
	"net/http"
	"os"
	"time"

	"github.com/KatePril/architecture-lab-5/httptools"
	"github.com/KatePril/architecture-lab-5/signal"
)

var port = flag.Int("port", 8080, "server port")

const confResponseDelaySec = "CONF_RESPONSE_DELAY_SEC"
const confHealthFailure = "CONF_HEALTH_FAILURE"

func main() {
	h := new(http.ServeMux)
	db, err := datastore.Open("db")
	if err != nil {
		// handle the error properly
		panic(err)
	}
	teamName := "s.k.a.m"
	currentDate := time.Now().Format("2006-01-02")
	db.Put(teamName, currentDate)

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

	report := make(Report)

	h.HandleFunc("/api/v1/some-data", func(rw http.ResponseWriter, r *http.Request) {
		//respDelayString := os.Getenv(confResponseDelaySec)
		//if delaySec, parseErr := strconv.Atoi(respDelayString); parseErr == nil && delaySec > 0 && delaySec < 300 {
		//	time.Sleep(time.Duration(delaySec) * time.Second)
		//}

		report.Process(r)
		key := r.URL.Query().Get("key")
		if key == "" {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		value, err := db.Get(key)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		rw.Header().Set("content-type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode(map[string]string{"value": value})
	})

	h.Handle("/report", report)

	server := httptools.CreateServer(*port, h)
	server.Start()
	signal.WaitForTerminationSignal()
}
