package main
import (
	"log"
	"net/http"
	"os"
	"recommendation-service/config"
)
func main() {
	cfg := config.Load()
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("recommendation-service is running"))
	})
	log.Println("Recommendation service running on port", cfg.Port)
	
	// Support HTTPS if certificates are provided
	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	if certFile != "" && keyFile != "" {
		log.Println("Starting HTTPS server on port", cfg.Port)
		log.Fatal(http.ListenAndServeTLS(":"+cfg.Port, certFile, keyFile, mux))
	} else {
		log.Println("Starting HTTP server on port", cfg.Port)
		log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
	}
}
