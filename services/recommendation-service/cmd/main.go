package main
import (
	"log"
	"net/http"
	"recommendation-service/config"
)
func main() {
	cfg := config.Load()
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("recommendation-service is running"))
	})
	log.Println("Recommendation service running on port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
