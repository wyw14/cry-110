package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	system *System
	router chi.Router
}

func NewServer(system *System) *Server {
	server := &Server{system: system, router: chi.NewRouter()}
	server.routes()
	return server
}

func (s *Server) routes() {
	s.router.Get("/healthz", s.health)
	s.router.Route("/api", func(router chi.Router) {
		router.Get("/isolation", s.getIsolation)
		router.Post("/isolation", s.startIsolation)
		router.Put("/isolation", s.advanceIsolation)
		router.Get("/pitch", s.getPitch)
		router.Post("/pitch", s.startPitch)
		router.Get("/yaw", s.getYaw)
		router.Post("/yaw", s.startYaw)
		router.Put("/yaw", s.finishYaw)
		router.Get("/crane", s.getCrane)
		router.Post("/crane", s.startCrane)
		router.Put("/crane", s.updateCrane)
		router.Get("/turning-gear", s.getTurningGear)
		router.Post("/turning-gear", s.startTurningGear)
		router.Put("/turning-gear", s.updateTurningGear)
		router.Get("/hydraulic", s.getHydraulic)
		router.Post("/hydraulic", s.updateHydraulic)
		router.Get("/generator", s.getGenerator)
		router.Post("/generator", s.updateGenerator)
		router.Get("/ventilation", s.getVentilation)
		router.Post("/ventilation", s.updateVentilation)
		router.Get("/access", s.getAccess)
		router.Post("/access", s.updateAccess)
		router.Get("/incidents", s.getIncidents)
		router.Post("/incidents", s.startIncident)
		router.Put("/incidents", s.updateIncident)
	})
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func respondJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func respondError(writer http.ResponseWriter, status int, err error) {
	respondJSON(writer, status, map[string]string{"error": err.Error()})
}

func decodeJSON(request *http.Request, value any) error {
	decoder := json.NewDecoder(io.LimitReader(request.Body, 1024*1024))
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}

func (s *Server) health(writer http.ResponseWriter, request *http.Request) {
	respondJSON(writer, http.StatusOK, map[string]any{
		"status": "ok", "journal": s.system.Journal.Path(), "time": time.Now().UTC(),
	})
}
