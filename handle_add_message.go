package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (s *Server) handleAddMessage() http.HandlerFunc {
	type Request struct {
		ChatId   int64  `json:"chat" validate:"required,gte=0"`
		AuthorId int64  `json:"author" validate:"required,gte=0"`
		Text     string `json:"text"`
	}
	type Responce struct {
		Id int64 `json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var request Request
		if err := s.decodeAndValidate(w, r, &request); err != nil {
			return
		}
		logger := s.getLogger(r).WithFields(logrus.Fields{
			"chat_id":   request.ChatId,
			"author_id": request.AuthorId,
			"msg_text":  request.Text,
		})
		if isUserInChat, _ := s.Storage.IsUserInChat(request.AuthorId, request.ChatId); !isUserInChat {
			s.respondWithError(w, r, logger, "user is not in the chat")
			return
		}
		messageId, err := s.Storage.AddMessage(request.AuthorId, request.ChatId, request.Text)
		if err != nil {
			s.respondWithInternalError(w, r, logger.WithField("error",
				fmt.Errorf("AddMessage failed: %s", err)))
			return
		}
		responce := Responce{messageId}
		s.respond(w, r, responce, http.StatusOK)
	}
}
