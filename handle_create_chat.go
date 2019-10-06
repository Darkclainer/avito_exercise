package main

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func (s *Server) handleAddChat() http.HandlerFunc {
	type Request struct {
		Name    string  `json:"name" validate:"chatname"`
		UserIds []int64 `json:"users" validate:"gt=0,unique,dive,gte=0,required"`
	}
	type Responce struct {
		Id int64 `json:id`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var request Request
		if err := s.decodeAndValidate(w, r, &request); err != nil {
			return
		}
		logger := s.getLogger(r).WithFields(logrus.Fields{
			"chat_name": request.Name,
			"users":     request.UserIds,
		})
		if isChatExists, _ := s.Storage.IsChatExists(request.Name); isChatExists {
			s.respondWithError(w, r, logger, "chat with the same name is already exists")
			return
		}
		if areUsersExist, _ := s.Storage.AreUsersExistByIds(request.UserIds); !areUsersExist {
			s.respondWithError(w, r, logger, "nonexistent user")
			return
		}
		chatId, err := s.Storage.AddChat(request.Name, request.UserIds)
		if err != nil {
			s.respondWithInternalError(w, r, logger.WithField("error",
				fmt.Errorf("AddChat failed: %s", err)))
			return
		}
		responce := Responce{chatId}
		s.respond(w, r, responce, http.StatusOK)
	}
}
