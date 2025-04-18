package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	storage "github.com/MPoline/alert_service_yp/cmd/server/memstorage"

	"net/http"
)

var memStorage = storage.NewMemStorage()

// func middleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodPost {
// 			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
// 			return
// 		}
// 		if r.Header.Get("Content-Type") != "text/plain" {
// 			http.Error(w, "Only Content-Type:text/plain are allowed!", http.StatusUnsupportedMediaType)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }

func updateMetric(w http.ResponseWriter, r *http.Request) {

	fmt.Println("URL: ", r.URL.Path)

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Only Content-Type:text/plain are allowed!", http.StatusUnsupportedMediaType)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 || parts[1] != "update" {
		http.Error(w, "Invalid URL format", http.StatusNotFound)
		return
	}

	metricType := parts[2]
	metricName := parts[3]
	metricValue := parts[4]

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}
	contentType := r.Header.Get("Content-Type")
	contentLength := r.Header.Get("Content-Length")
	date := time.Now().UTC().Format(http.TimeFormat)

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		memStorage.SetGauge(metricName, value)
		newValue, checkValue := memStorage.GetGauge(metricName)
		fmt.Println("newValue: ", newValue)
		fmt.Println("checkValue: ", checkValue)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Date", date)
		w.Header().Set("Content-Length", contentLength)
		w.Header().Set("Content-Type", contentType)

	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		memStorage.IncrementCounter(metricName, value)
		newValue, checkValue := memStorage.GetCounter(metricName)
		fmt.Println("newValue: ", newValue)
		fmt.Println("checkValue: ", checkValue)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Date", date)
		w.Header().Set("Content-Length", contentLength)
		w.Header().Set("Content-Type", contentType)

	default:
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
	}

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, updateMetric)

	fmt.Println("Starting server on :8080")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
