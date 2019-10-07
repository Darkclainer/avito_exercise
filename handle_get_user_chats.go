package main

import (
	"fmt"
	"net/http"

	"github.com/Darkclainer/avito_exercise/storage"
	"github.com/sirupsen/logrus"
)

func (s *Server) handleGetUserChats() http.HandlerFunc {
	type Request struct {
		UserId int64 `json:"user" validate:"required,gte=0"`
	}
	type Responce struct {
		Chats []*storage.Chat
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var request Request
		if err := s.decodeAndValidate(w, r, &request); err != nil {
			return
		}
		logger := s.getLogger(r).WithFields(logrus.Fields{
			"user_id": request.UserId,
		})

		chats, err := s.Storage.GetUserChats(request.UserId)
		if err != nil {
			s.respondWithInternalError(w, r, logger.WithField("error",
				fmt.Errorf("GetUserChats failed: %s", err)))
		}
		responce := Responce{chats}
		s.respond(w, r, responce, http.StatusOK)
	}
}
