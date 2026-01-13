package main

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"api-gateway/config"
)

// proxyRequest prosleđuje zahtev ka backend servisu
func proxyRequest(w http.ResponseWriter, r *http.Request, targetURL string) {
	// Čitanje body-ja zahteva
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Kreiranje novog zahteva ka backend servisu
	req, err := http.NewRequest(r.Method, targetURL, bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Kopiranje headers-a
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Slanje zahteva
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Kopiranje status koda i headers-a
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Kopiranje body-ja odgovora
	responseBody, _ := io.ReadAll(resp.Body)
	w.Write(responseBody)
}

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()

	// USERS SERVICE ROUTES
	mux.HandleFunc("/api/users/health", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/health")
	})

	mux.HandleFunc("/api/users/register", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/register")
	})

	mux.HandleFunc("/api/users/login/request-otp", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/login/request-otp")
	})

	mux.HandleFunc("/api/users/login/verify-otp", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/login/verify-otp")
	})

	mux.HandleFunc("/api/users/password/change", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/password/change")
	})

	mux.HandleFunc("/api/users/password/reset", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.UsersServiceURL+"/password/reset")
	})

	// CONTENT SERVICE ROUTES
	mux.HandleFunc("/api/content/health", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, cfg.ContentServiceURL+"/health")
	})

	log.Println("API Gateway running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
