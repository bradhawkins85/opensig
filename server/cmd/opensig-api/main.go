package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Health struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(Health{Status: "ok", Time: time.Now().UTC().Format(time.RFC3339)})
}

func preview(w http.ResponseWriter, r *http.Request) {
	// Minimal stub: returns a rendered placeholder with fake user data
	type Resp struct {
		HTML string `json:"html"`
		Text string `json:"text"`
	}
	resp := Resp{
		HTML: `<div style="font-family:Segoe UI,Arial,sans-serif"><strong>Jane Doe</strong><br>Senior Engineer<br><img src="https://via.placeholder.com/120x40" alt="Logo"></div>`,
		Text: "Jane Doe\nSenior Engineer",
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/v1/preview", preview)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	log.Printf("OpenSig API listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
