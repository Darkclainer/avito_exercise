package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Darkclainer/avito_exercise/mocks"
)

func TestHandleAddMessage(t *testing.T) {
	type TestCase struct {
		TestName           string
		RequestBody        string
		ExpectedStatusCode int
		ExpectedErrorMsg   string
		MockReturnId       int64
		SetupStorage       func(mock *mocks.Storage, testCase *TestCase)
	}

	testCases := []*TestCase{
		&TestCase{
			TestName:           "Add message",
			RequestBody:        `{"chat": 10, "author": 20, "text": "Hello, World!"}`,
			ExpectedStatusCode: http.StatusOK,
			MockReturnId:       50,
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
				mock.On("IsUserInChat", int64(20), int64(10)).Return(true, nil)
				mock.On("AddMessage", int64(20), int64(10), "Hello, World!").Return(testCase.MockReturnId, nil)
			},
		},
		&TestCase{
			TestName:           "Add message with empty text",
			RequestBody:        `{"chat": 10, "author": 20, "text": ""}`,
			ExpectedStatusCode: http.StatusOK,
			MockReturnId:       50,
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
				mock.On("IsUserInChat", int64(20), int64(10)).Return(true, nil)
				mock.On("AddMessage", int64(20), int64(10), "").Return(testCase.MockReturnId, nil)
			},
		},
		&TestCase{
			TestName:           "Add message to nonexistent chat",
			RequestBody:        `{"chat": 10, "author": 20, "text": "Hello, World!"}`,
			ExpectedErrorMsg:   "user is not in the chat",
			ExpectedStatusCode: http.StatusInternalServerError,
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
				mock.On("IsUserInChat", int64(20), int64(10)).Return(false, nil)
			},
		},
		&TestCase{
			TestName:           "Add message without chat",
			RequestBody:        `{"author": 20, "text": "Hello, World!"}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Add message without author",
			RequestBody:        `{"chat": 10, "text": "Hello, World!"}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
			},
		},
	}
	server := NewServer(nil, nil, true)

	type Responce struct {
		Id    int64  `json:"id"`
		Error string `json:"error"`
	}
	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			mockStorage := &mocks.Storage{}
			server.Storage = mockStorage

			requestData := strings.NewReader(testCase.RequestBody)
			request, err := http.NewRequest(http.MethodPost, "/chats/add", requestData)
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()

			testCase.SetupStorage(mockStorage, testCase)

			handler := server.handleAddMessage()
			handler.ServeHTTP(recorder, request)

			mockStorage.AssertExpectations(t)
			assert.Equal(t, testCase.ExpectedStatusCode, recorder.Code)

			var responce Responce
			if assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &responce)) {
				assert.Equal(t, testCase.MockReturnId, responce.Id)
				assert.Equal(t, testCase.ExpectedErrorMsg, responce.Error)
			}
		})
	}
}
