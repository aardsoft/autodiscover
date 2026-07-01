package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/aardsoft/autodiscover/internal/config"
	"github.com/aardsoft/autodiscover/internal/handlers"
)

var version = "dev"

func main() {
	addr := flag.String("addr", ":8080", "TCP address to listen on")
	configPath := flag.String("config", "settings.json", "Path to settings.json")
	flag.Parse()

	settings, err := config.LoadSettings(*configPath)
	if err != nil {
		log.Fatalf("Failed to load settings: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /autodiscover/autodiscover.xml", handlers.HandleMSAutodiscover(settings))
	mux.HandleFunc("GET /autodiscover/autodiscover.xml", handlers.HandleMSAutodiscoverGET(settings))
	mux.HandleFunc("GET /mail/config-v1.1.xml", handlers.HandleThunderbird(settings))
	mux.HandleFunc("GET /.well-known/autoconfig/mail/config-v1.1.xml", handlers.HandleThunderbird(settings))
	mux.HandleFunc("GET /profile.mobileconfig", handlers.HandleAppleProfile(settings))
	mux.HandleFunc("GET /.well-known/caldav", handlers.HandleWellKnownCalDAV(settings))
	mux.HandleFunc("GET /.well-known/carddav", handlers.HandleWellKnownCardDAV(settings))

	log.Printf("Autodiscover server v%s listening on %s", version, *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
