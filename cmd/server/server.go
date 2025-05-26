package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/KatePril/architecture-lab-5/httptools"
	"github.com/KatePril/architecture-lab-5/signal"
)

var port = flag.Int("port", 8080, "server port")

const (
	dbServiceURL      = "http://database-server:8091/db/"
	healthCheckURL    = "http://database-server:8091/health"
	maxRetryAttempts  = 10
	confHealthFailure = "CONF_HEALTH_FAILURE"
)

func waitForServerReady() {
	for i := 0; i < maxRetryAttempts; i++ {
		resp, err := http.Get(healthCheckURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			fmt.Println("Server is healthy. Proceeding...")
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		fmt.Printf("Health check failed (attempt %d/%d). Retrying...\n", i+1, maxRetryAttempts)
		time.Sleep(1 * time.Second)
	}
	panic("Server did not become healthy in time")
}

func postTeamName() {
	waitForServerReady()

	teamName := "s.k.a.m"
	currentDate := time.Now().Format("2006-01-02")
	requestBody := map[string]interface{}{
		"value": currentDate,
	}
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		panic(err)
	}

	url := dbServiceURL + teamName
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBytes))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("POST attempt %d failed: %v\n", i+1, err)
			time.Sleep(1 * time.Second)
			continue
		}
		defer resp.Body.Close()
		fmt.Printf("POST succeeded with status: %s\n", resp.Status)
		return
	}

	panic("Failed to POST after multiple attempts")
}

func main() {
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

	report := make(Report)

	h.HandleFunc("/api/v1/some-data", func(rw http.ResponseWriter, r *http.Request) {
		report.Process(r)
		key := r.URL.Query().Get("key")
		if key == "" {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		resp, err := http.DefaultClient.Get(dbServiceURL + key)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		if resp.StatusCode != http.StatusOK {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		var result struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			http.Error(rw, "Failed to decode db response", http.StatusInternalServerError)
			return
		}

		rw.Header().Set("content-type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode(map[string]string{"value": result.Value})
	})

	h.Handle("/report", report)

	server := httptools.CreateServer(*port, h)
	server.Start()
	waitForServerReady()
	postTeamName()
	signal.WaitForTerminationSignal()
}
