package services

import (
	"fmt"
	"net/http/httputil"
	"time"

	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/go-resty/resty/v2"
)

var (
	serverURL = "http://localhost:8080/update"
	nRetries  = 3
)

func CreateURLS(s *storage.MemStorage) (URLStorage []string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for key, value := range s.Gauges {
		url := fmt.Sprintf("%s/gauge/%s/%f", serverURL, key, value)
		URLStorage = append(URLStorage, url)
	}

	for key, value := range s.Counters {
		url := fmt.Sprintf("%s/counter/%s/%d", serverURL, key, value)
		URLStorage = append(URLStorage, url)
	}
	return
}

func SendMetrics(s *storage.MemStorage, URLStorage []string) {
	client := resty.New()

	for _, URL := range URLStorage {
		nAttempts := 0
		for nAttempts < nRetries {
			req := client.R().SetHeader("Content-Type", "text/plain")
			req.Body = ""
			req.URL = URL
			req.Method = "POST"
			resp, err := req.Send()

			if err != nil {
				fmt.Println("Error sending request:", err)
				nAttempts++
				time.Sleep(2 * time.Second)
				dump, _ := httputil.DumpRequest(req.RawRequest, true)
				fmt.Printf("Оригинальный запрос:\n\n%s", dump)
				continue
			}
			if resp.IsError() {
				fmt.Println("Error response:", resp.Status())
			}
			break
		}
		if nAttempts == nRetries {
			fmt.Println("All retries failed for URL:", URL)
		}
	}
}
