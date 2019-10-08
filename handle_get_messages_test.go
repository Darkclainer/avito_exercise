package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Darkclainer/avito_exercise/mocks"
	"github.com/Darkclainer/avito_exercise/storage"
)

func TestHandleGetMessages(t *testing.T) {
	type TestCase struct {
		TestName           string
		RequestBody        string
		ExpectedStatusCode int
		ExpectedErrorMsg   string
		Responce           []*storage.Message
		SetupStorage       func(mock *mocks.Storage, testCase *TestCase)
	}

	testCases := []*TestCase{
		&TestCase{
			TestName:           "OK",
			RequestBody:        `{"chat": 10}`,
			ExpectedStatusCode: http.StatusOK,
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
				mock.On("GetMessagesFromChat", int64(10)).Return(testCase.Responce, nil)
			},
			Responce: []*storage.Message{
				&storage.Message{
					Id: 21, ChatId: 10, AuthorId: 1,
					Text:      "Hello",
					CreatedAt: time.Date(2019, time.January, 1, 10, 0, 0, 0, time.UTC),
				},
				&storage.Message{
					Id: 23, ChatId: 10, AuthorId: 1,
					Text:      "I can travel in time",
					CreatedAt: time.Date(2019, time.January, 1, 10, 0, 30, 0, time.UTC),
				},
			},
		},
		&TestCase{
			TestName:           "Message list empty",
			RequestBody:        `{"chat": 11}`,
			ExpectedStatusCode: http.StatusOK,
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
				mock.On("GetMessagesFromChat", int64(11)).Return(testCase.Responce, nil)
			},
			Responce: []*storage.Message{},
		},
		&TestCase{
			TestName:           "Without chat id",
			RequestBody:        `{"user": 123}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
			},
		},
	}
	server := NewServer(nil, nil, true)

	type Responce struct {
		Messages []*storage.Message `json:"messages"`
		Error    string             `json:"error"`
	}
	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			mockStorage := &mocks.Storage{}
			server.Storage = mockStorage

			requestData := strings.NewReader(testCase.RequestBody)
			request, err := http.NewRequest(http.MethodPost, "/messages/get", requestData)
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()

			testCase.SetupStorage(mockStorage, testCase)

			handler := server.handleGetMessages()
			handler.ServeHTTP(recorder, request)

			mockStorage.AssertExpectations(t)
			assert.Equal(t, testCase.ExpectedStatusCode, recorder.Code)

			var responce Responce
			if assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &responce)) {
				assert.Equal(t, testCase.Responce, responce.Messages)
				assert.Equal(t, testCase.ExpectedErrorMsg, responce.Error)
			}
		})
	}
}
