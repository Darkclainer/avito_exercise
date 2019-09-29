package main

import (
	"net/http"
	"regexp"
)

/* handleUserAdd return handle that andd new user on POST method

Request body must be json dictionary with field "username" and value string.
Valid username must start with ASCII letter and continue with letter, number or underscore.
Maximum length is 32 characters.
Handler return id of new user or error msg.
*/
func (s *Server) handleUserAdd() http.HandlerFunc {
	type Request struct {
		Username string `json:username`
	}
	type Responce struct {
		Id int64 `json:id`
	}
	validUsername := regexp.MustCompile(`^[a-zA-Z]\w{0,31}$`)
	return func(w http.ResponseWriter, r *http.Request) {
		var request Request
		if err := s.decode(w, r, &request); err != nil {
			return
		}
		if !validUsername.MatchString(request.Username) {
			s.respondError(w, r, "Username is in unsupported format", nil)
			return
		}
		if isSameUser, _ := s.DB.IsUserExists(request.Username); isSameUser {
			s.respondError(w, r, "User with this username is already added", nil)
			return
		}
		result, err := s.DB.AddUser(request.Username)
		if err != nil {
			s.respondInternalError(w, r, "INSERT user failed unexpectedly", err)
			return
		}
		lastId, err := result.LastInsertId()
		if err != nil {
			s.respondInternalError(w, r, "LastInsertId failed", err)
			return
		}
		responce := Responce{lastId}
		s.respond(w, r, responce, http.StatusOK)
	}
}
