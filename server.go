package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Server struct {
	router    *mux.Router
	Logger    *logrus.Logger
	DB        ServerDB
	isTesting bool
}

func NewServer(db *sql.DB, logger *logrus.Logger, isTesting bool) *Server {
	s := &Server{
		router:    mux.NewRouter(),
		Logger:    logger,
		DB:        ServerDB{db},
		isTesting: isTesting,
	}
	if logger == nil {
		s.Logger = logrus.New()
	}
	if isTesting {
		s.Logger.Level = logrus.ErrorLevel
	} else {
		s.Logger.Level = logrus.DebugLevel
	}
	s.routes()
	return s
}

func (s *Server) getRequestLogger(r *http.Request) *logrus.Entry {
	return s.Logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"method": r.Method,
	})
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"method": r.Method,
	}).Debug("New request")
	s.router.ServeHTTP(w, r)
}

func (s *Server) respond(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			s.Logger.WithFields(logrus.Fields{
				"url":   r.URL,
				"data":  data,
				"error": err,
			}).Error("json encode error while responding")
		}
	}
}

func (s *Server) decode(w http.ResponseWriter, r *http.Request, value interface{}) error {
	err := json.NewDecoder(r.Body).Decode(value)
	if err != nil {
		s.respondError(w, r, "json decoding error", err)
	}
	return err
}

func (s *Server) respondError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	logger := s.getRequestLogger(r)
	logger.WithField("error", err).Debug("Server responded with error")

	msgValue := map[string]string{"error": msg}
	s.respond(w, r, msgValue, http.StatusInternalServerError)
}
func (s *Server) respondInternalError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	logger := s.getRequestLogger(r)
	logger.WithField("error", err).Error("Server responded with internal error")

	msgValue := map[string]string{"error": "Internal error"}
	s.respond(w, r, msgValue, http.StatusInternalServerError)
}
