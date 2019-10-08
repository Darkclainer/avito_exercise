package main

import (
	"fmt"
	"net/http"

	"github.com/Darkclainer/avito_exercise/storage"
	"github.com/sirupsen/logrus"
)

func (s *Server) handleGetMessages() http.HandlerFunc {
	type Request struct {
		ChatId int64 `json:"chat" validate:"required,gte=0"`
	}
	type Responce struct {
		Messages []*storage.Message `json:"messages"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var request Request
		if err := s.decodeAndValidate(w, r, &request); err != nil {
			return
		}
		logger := s.getLogger(r).WithFields(logrus.Fields{
			"chat_id": request.ChatId,
		})

		messages, err := s.Storage.GetMessagesFromChat(request.ChatId)
		if err != nil {
			s.respondWithInternalError(w, r, logger.WithField("error",
				fmt.Errorf("GetMessagesFromChat failed: %s", err)))
		}
		responce := Responce{messages}
		s.respond(w, r, responce, http.StatusOK)
	}
}
