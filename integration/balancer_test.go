package integration

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

const baseAddress = "http://balancer:8090"

var client = http.Client{
	Timeout: 3 * time.Second,
}

func TestBalancer(t *testing.T) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		t.Skip("Integration test is not enabled")
	}

	// Перевіряємо, чи балансувальник правильно розподіляє запити між серверами
	servers := map[string]int{
		"server1:8080": 0,
		"server2:8080": 0,
		"server3:8080": 0,
	}

	for i := 0; i < 10; i++ {
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		from := resp.Header.Get("lb-from")
		if from == "" {
			t.Errorf("Missing 'lb-from' header in response")
		} else {
			servers[from]++
		}
	}

	// Перевіряємо, чи всі сервери отримали хоча б один запит
	for server, count := range servers {
		if count == 0 {
			t.Errorf("Server %s did not receive any requests", server)
		}
	}
}

func BenchmarkBalancer(b *testing.B) {
	// TODO: Реалізуйте інтеграційний бенчмарк для балансувальникка.
}
