package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

type Server struct {
	router    *mux.Router
	Logger    *logrus.Logger
	DB        ServerDB
	validate  *validator.Validate
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
	s.validate = NewValidate()
	return s
}

func (s *Server) getLogger(r *http.Request) *logrus.Entry {
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

// respond sends respond with json data and log if there is any error whyle encoding.
func (s *Server) respond(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			s.getLogger(r).WithFields(logrus.Fields{
				"data":  data,
				"error": err,
			}).Error("json encode error while responding")
		}
	}
}

// decode unmarshal json value from request body and respond with error if there is formating issues.
func (s *Server) decode(w http.ResponseWriter, r *http.Request, value interface{}) error {
	err := json.NewDecoder(r.Body).Decode(value)
	if err != nil {
		s.respondWithError(w, r, s.getLogger(r).WithField("error", err), "json decoding error")
	}
	return err
}

// decodeAndValidate unmarshal json value and validate it. It reports of errors to client if there is one.
func (s *Server) decodeAndValidate(w http.ResponseWriter, r *http.Request, value interface{}) error {
	if err := s.decode(w, r, value); err != nil {
		return err
	}
	err := s.validate.Struct(value)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			s.respondWithInternalError(w, r, s.getLogger(r).WithField("error", err))
			return err
		}
		logger := s.getLogger(r)
		for _, err := range err.(validator.ValidationErrors) {
			logger.WithFields(logrus.Fields{
				"tag":   err.ActualTag(),
				"field": err.StructNamespace(),
				"value": err.Value(),
			}).Debug("Error while validating")
		}
		s.respondWithError(w, r, logger, "invalid input")
		return err
	}
	return nil
}

// respondWithError responds with "error" field and specified msg and log it
// Also it sets status code to http.StatusInternalServerError
func (s *Server) respondWithError(w http.ResponseWriter, r *http.Request, logger *logrus.Entry, msg string) {
	if logger == nil {
		logger = s.getLogger(r)
	}
	logger.WithField("respond_msg", msg).Debug("Server responded with error")

	type Responce struct {
		Error string `json:error`
	}
	s.respond(w, r, Responce{msg}, http.StatusInternalServerError)
}

// respondWithInternalError works as respondWithError, but it has predifined msg.
// Its function used for hiding from client what kind of error happened.
func (s *Server) respondWithInternalError(w http.ResponseWriter, r *http.Request, logger *logrus.Entry) {
	s.respondWithError(w, r, logger, "internal error")
}
