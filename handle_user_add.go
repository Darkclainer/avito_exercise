package main

import (
	"fmt"
	"net/http"
)

/* handleUserAdd return handle that andd new user on POST method

Request body must be json dictionary with field "username" and value string.
Valid username must start with ASCII letter and continue with letter, number or underscore.
Maximum length is 32 characters.
Handler return id of new user or error msg.
*/
func (s *Server) handleUserAdd() http.HandlerFunc {
	type Request struct {
		Username string `json:"username" validate:"username"`
	}
	type Responce struct {
		Id int64 `json:id`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var request Request
		if err := s.decodeAndValidate(w, r, &request); err != nil {
			return
		}
		logger := s.getLogger(r).WithField("username", request.Username)
		if isSameUser, _ := s.DB.IsUserExists(request.Username); isSameUser {
			s.respondWithError(w, r, logger, "User with this username is already added")
			return
		}
		result, err := s.DB.AddUser(request.Username)
		if err != nil {
			s.respondWithInternalError(w, r, logger.WithField("error",
				fmt.Errorf("INSERT user failed unexpectedly: %v", err)))
			return
		}
		lastId, err := result.LastInsertId()
		if err != nil {
			s.respondWithInternalError(w, r,
				logger.WithField("error", fmt.Errorf("LastInsertId failed: %v", err)))
			return
		}
		responce := Responce{lastId}
		s.respond(w, r, responce, http.StatusOK)
	}
}
