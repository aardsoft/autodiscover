package handlers

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
)

func writeXML(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(status)
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(v); err != nil {
		log.Printf("XML encode error: %v", err)
	}
}

func writeMobileConfig(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/x-apple-aspen-config")
	w.Header().Set("Content-Disposition", "attachment; filename=\"profile.mobileconfig\"")
	w.WriteHeader(status)
	w.Write(data)
}

func handleNotConfigured(w http.ResponseWriter, service string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s is not configured\n", service)
}
