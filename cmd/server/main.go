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

func getMetric(w http.ResponseWriter, r *http.Request) {

	fmt.Println("URL: ", r.URL.Path)

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 || parts[1] != "value" {
		http.Error(w, "Invalid URL format", http.StatusNotFound)
		return
	}

	metricType := parts[2]
	metricName := parts[3]

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	var value string
	var found bool
	switch metricType {
	case "gauge":
		if val, ok := memStorage.GetGauge(metricName); ok {
			value = fmt.Sprintf("%f", val)
			found = true
		}
	case "counter":
		if val, ok := memStorage.GetCounter(metricName); ok {
			value = fmt.Sprintf("%d", val)
			found = true
		}
	default:
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}

	if !found {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	w.Write([]byte(value))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
}

func getAllMetrics(w http.ResponseWriter, r *http.Request) {

	fmt.Println("URL: ", r.URL.Path)

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var sb strings.Builder
	sb.WriteString("<html><body><h1>Metrics</h1><ul>")

	for metricName, metricValue := range memStorage.Gauges {
		sb.WriteString(fmt.Sprintf("<li> Gauge metrics: %s - %f</li>", metricName, metricValue))
	}

	for metricName, metricValue := range memStorage.Counters {
		sb.WriteString(fmt.Sprintf("<li> Counter metrics: %s - %d</li>", metricName, metricValue))
	}

	sb.WriteString("</ul></body></html>")

	w.Write([]byte(sb.String()))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, updateMetric)
	mux.HandleFunc(`/value/`, getMetric)
	mux.HandleFunc(`/`, getAllMetrics)

	fmt.Println("Starting server on :8080")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
