package handlers

import (
	"net/http"

	"github.com/aardsoft/autodiscover/internal/config"
)

// HandleWellKnownMatrixServer serves /.well-known/matrix/server for
// server-name Federation delegation (m.server).
// See https://spec.matrix.org/latest/server-server-api/#server-discovery
func HandleWellKnownMatrixServer(s *config.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.Matrix == nil || s.Matrix.Server == "" {
			handleNotConfigured(w, "Matrix server delegation")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"m.server": s.Matrix.Server,
		})
	}
}

// HandleWellKnownMatrixClient serves /.well-known/matrix/client for
// homeserver discovery by clients logging in with a full Matrix ID.
// See https://spec.matrix.org/latest/client-server-api/#well-known-uri
func HandleWellKnownMatrixClient(s *config.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.Matrix == nil || s.Matrix.Homeserver == "" {
			handleNotConfigured(w, "Matrix client discovery")
			return
		}
		identity := s.Matrix.Homeserver
		if s.Matrix.IdentityServer != "" {
			identity = s.Matrix.IdentityServer
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"m.homeserver": map[string]string{
				"base_url": s.Matrix.Homeserver,
			},
			"m.identity_server": map[string]string{
				"base_url": identity,
			},
		})
	}
}
