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

func TestHandleAddUser(t *testing.T) {
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
			TestName:           "SimpleAdd",
			RequestBody:        `{"username": "user_1"}`,
			ExpectedStatusCode: http.StatusOK,
			MockReturnId:       1,
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
				mock.On("IsUserExists", "user_1").Return(false, nil).Once()
				mock.On("AddUser", "user_1").Return(int64(1), nil).Once()
			},
		},
		&TestCase{
			TestName:           "Error request format",
			RequestBody:        `{"username": "user_1"`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "json decoding error",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Add user with duplicate username",
			RequestBody:        `{"username": "user_1"}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "User with this username is already added",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
				mock.On("IsUserExists", "user_1").Return(true, nil).Once()
			},
		},
		&TestCase{
			TestName:           "Incorrect username",
			RequestBody:        `{"username": "1_user"}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupStorage: func(mock *mocks.Storage, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Too long username",
			RequestBody:        `{"username": "u234567890123456789012345679012345"}`,
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
			request, err := http.NewRequest(http.MethodPost, "/users/add", requestData)
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()

			testCase.SetupStorage(mockStorage, testCase)

			handler := server.handleAddUser()
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
