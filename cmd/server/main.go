package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"local/KOrche/internal/api"
	"local/KOrche/internal/placer"
)

func main() {
	log.SetOutput(os.Stdout)

	http.HandleFunc("/place", placeHandler)
	http.HandleFunc("/health", healthHandler) // NEW

	port := "8080"
	fmt.Printf("Server listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// -----------------------------------------------------------------------------
// /health - Simple probe
// -----------------------------------------------------------------------------
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// -----------------------------------------------------------------------------
// /place - Main placement endpoint
// Optional query ?output=file -> saves result to a JSON file
// -----------------------------------------------------------------------------
func placeHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Log ricezione richiesta
	log.Printf(
		"[INFO] Request: %s %s from %s at %s (UTC)",
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		start.Format(time.RFC3339),
	)

	// Log finale garantito
	defer func() {
		end := time.Now()
		log.Printf(
			"[INFO] Response: %s %s completed at %s (UTC). Duration: %s",
			r.Method,
			r.URL.Path,
			end.Format(time.RFC3339),
			end.Sub(start),
		)
	}()

	// resto della funzione
	if r.Method != http.MethodPost {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var req api.PlacementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := placer.Place(req)
	if err != nil {
		http.Error(w, "error placing: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// CHECK OUTPUT MODE
	outputMode := r.URL.Query().Get("output")
	if outputMode == "file" {
		fileName, err := saveResponseToFile(resp)
		if err != nil {
			http.Error(w, "file save error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"saved": fileName,
		})
		return
	}

	// DEFAULT: respond over HTTP
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resp); err != nil {
		http.Error(w, "encode error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// -----------------------------------------------------------------------------
// Utility: Save JSON Response to File
// -----------------------------------------------------------------------------
func saveResponseToFile(resp *api.PlacementResult) (string, error) {
	// timestamp := time.Now().Format("20060102-150405")
	podID := "unknown-pod"
	if resp.PodID != "" { // if you have resp.PodID; adjust accordingly
		podID = resp.PodID
	}

	// Ensure directory exists
	outDir := getOutputDir()
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}

	// fileName := fmt.Sprintf("%s/placement-%s-%s.json", outDir, podID, timestamp)
	fileName := fmt.Sprintf("placement-%s.json", podID)
	fileName = filepath.Join(outDir, fileName)

	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resp); err != nil {
		return "", err
	}

	return fileName, nil
}

func getOutputDir() string {
	return "out"
}

// docker build -t korche-placer .
// docker run -p 8080:8080 korche-placer
// or docker run -p 8080:8080 -v ${PWD}\out:/app/out korche-placer

//or

// go build ./cmd/server
// ./server.exe

//or

// docker save -o korche-placer.tar korche-placer

//or

// docker load -i "korche-placer.tar"

// RUN con:
// curl -X POST -H "Content-Type: application/json" --data-binary @request.json "http://<IP>:8080/place"
// curl -X POST -H "Content-Type: application/json" --data-binary @request.json "http://<IP>:8080/place?output=file"
