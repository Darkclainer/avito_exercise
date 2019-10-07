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

func TestHandleGetUserChats(t *testing.T) {
	type TestCase struct {
		TestName           string
		RequestBody        string
		ExpectedStatusCode int
		ExpectedErrorMsg   string
		Responce           []*storage.Chat
		SetupStorage       func(mock *mocks.Storage, testCase *TestCase)
	}

	testCases := []*TestCase{
		&TestCase{
			TestName:           "OK",
			RequestBody:        `{"user": 1}`,
			ExpectedStatusCode: http.StatusOK,
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
				mock.On("GetUserChats", int64(1)).Return(testCase.Responce, nil)
			},
			Responce: []*storage.Chat{
				&storage.Chat{
					Id:        10,
					Name:      "chat_1",
					CreatedAt: time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC),
					UserIds:   []int64{1, 2, 3},
				},
				&storage.Chat{
					Id:        11,
					Name:      "chat_2",
					CreatedAt: time.Date(2019, time.January, 1, 0, 0, 1, 30, time.UTC),
					UserIds:   []int64{1, 3, 5},
				},
			},
		},
		&TestCase{
			TestName:           "Chat list empty",
			RequestBody:        `{"user": 3}`,
			ExpectedStatusCode: http.StatusOK,
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
				mock.On("GetUserChats", int64(3)).Return(testCase.Responce, nil)
			},
			Responce: []*storage.Chat{},
		},
		&TestCase{
			TestName:           "Without user id",
			RequestBody:        `{"username": "user1"}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
			},
		},
	}
	server := NewServer(nil, nil, true)

	type Responce struct {
		Chats []*storage.Chat `json:"chats"`
		Error string          `json:"error"`
	}
	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			mockStorage := &mocks.Storage{}
			server.Storage = mockStorage

			requestData := strings.NewReader(testCase.RequestBody)
			request, err := http.NewRequest(http.MethodPost, "/chats/get", requestData)
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()

			testCase.SetupStorage(mockStorage, testCase)

			handler := server.handleGetUserChats()
			handler.ServeHTTP(recorder, request)

			mockStorage.AssertExpectations(t)
			assert.Equal(t, testCase.ExpectedStatusCode, recorder.Code)

			var responce Responce
			if assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &responce)) {
				assert.Equal(t, testCase.Responce, responce.Chats)
				assert.Equal(t, testCase.ExpectedErrorMsg, responce.Error)
			}
		})
	}
}
