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

func TestHandleCreateChat(t *testing.T) {
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
			TestName:           "Add chat with one user",
			RequestBody:        `{"name": "chat_1", "users": [1]}`,
			ExpectedStatusCode: http.StatusOK,
			MockReturnId:       2,
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
				mock.On("IsChatExists", "chat_1").Return(false, nil)
				mock.On("AreUsersExistByIds", []int64{1}).Return(true, nil)
				mock.On("AddChat", "chat_1", []int64{1}).Return(testCase.MockReturnId, nil)
			},
		},
		&TestCase{
			TestName:           "Add chat with three users",
			RequestBody:        `{"name": "chat_1", "users": [1, 2, 3]}`,
			ExpectedStatusCode: http.StatusOK,
			MockReturnId:       4,
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
				mock.On("IsChatExists", "chat_1").Return(false, nil)
				mock.On("AreUsersExistByIds", []int64{1, 2, 3}).Return(true, nil)
				mock.On("AddChat", "chat_1", []int64{1, 2, 3}).Return(testCase.MockReturnId, nil)
			},
		},
		&TestCase{
			TestName:           "Add chat with nonexistent user",
			RequestBody:        `{"name": "chat_1", "users": [1, 123]}`,
			ExpectedErrorMsg:   "nonexistent user",
			ExpectedStatusCode: http.StatusInternalServerError,
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
				mock.On("IsChatExists", "chat_1").Return(false, nil)
				mock.On("AreUsersExistByIds", []int64{1, 123}).Return(false, nil)
			},
		},
		&TestCase{
			TestName:           "Add chat with duplicated name",
			RequestBody:        `{"name": "chat_1", "users": [1, 2]}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "chat with the same name is already exists",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
				mock.On("IsChatExists", "chat_1").Return(true, nil)
			},
		},
		&TestCase{
			TestName:           "Query without users",
			RequestBody:        `{"name": "chat"}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Query without chat name",
			RequestBody:        `{"users": [1, 2]}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Query with empty users list",
			RequestBody:        `{"name": "chat", "users": []}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Query with duplicated users id",
			RequestBody:        `{"name": "chat", "users": [2, 1, 1, 3]}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Query with invalid chat name",
			RequestBody:        `{"name": "12chat", "users": [1, 2]}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Query with invalid user id type",
			RequestBody:        `{"name": "chat", "users": [1, "sdf"]}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "json decoding error",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Query with below zero user id",
			RequestBody:        `{"name": "chat", "users": [1, -1]}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
			},
		},
	}
	server := NewServer(nil, nil, true)

	type Responce struct {
		Id    int64  `json:id`
		Error string `json:error`
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

			handler := server.handleChatAdd()
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
