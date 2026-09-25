package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/example/jev-banking-routing-go/internal/domain"
	"github.com/example/jev-banking-routing-go/internal/jev"
	"github.com/example/jev-banking-routing-go/internal/routing"
	"github.com/example/jev-banking-routing-go/internal/service"
)

func main() {
	client := jev.New(os.Getenv("TYPESAFE_API_KEY"))
	router := service.Router{
		Jev:    client,
		Policy: routing.DefaultPolicy(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("POST /v1/transactions/route", func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		var in domain.EvaluationRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&in); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(req.Context(), 2500*time.Millisecond)
		defer cancel()
		out := router.Route(ctx, in)

		w.Header().Set("Content-Type", "application/json")
		if out.FallbackUsed {
			w.Header().Set("X-Routing-Fallback", "true")
		}
		_ = json.NewEncoder(w).Encode(out)
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("routing API listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
