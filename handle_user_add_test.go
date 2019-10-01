package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestHandleUserAdd(t *testing.T) {
	type TestCase struct {
		TestName           string
		RequestBody        string
		ExpectedStatusCode int
		ExpectedErrorMsg   string
		MockReturnId       int64
		SetupMock          func(mock sqlmock.Sqlmock, testCase *TestCase)
	}

	testCases := []*TestCase{
		&TestCase{
			TestName:           "SimpleAdd",
			RequestBody:        `{"username": "user_1"}`,
			ExpectedStatusCode: http.StatusOK,
			MockReturnId:       1,
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("user_1", MatchTime{Start: time.Now()}).
					WillReturnResult(sqlmock.NewResult(testCase.MockReturnId, 1))
			},
		},
		&TestCase{
			TestName:           "Error request format",
			RequestBody:        `{"username": "user_1"`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "json decoding error",
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Add user with duplicate username",
			RequestBody:        `{"username": "user_1"}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "User with this username is already added",
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
				rows := sqlmock.NewRows([]string{"username"}).AddRow("user_1")
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
		},
		&TestCase{
			TestName:           "Incorrect username",
			RequestBody:        `{"username": "1_user"}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Too long username",
			RequestBody:        `{"username": "u234567890123456789012345679012345"}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
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
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal("Can not create stub database connection: ", err)
			}
			defer db.Close()
			server.DB = ServerDB{db}

			requestData := strings.NewReader(testCase.RequestBody)
			request, err := http.NewRequest(http.MethodPost, "/users/add", requestData)
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()

			testCase.SetupMock(mock, testCase)

			handler := server.handleUserAdd()
			handler.ServeHTTP(recorder, request)

			assert.NoError(t, mock.ExpectationsWereMet())
			assert.Equal(t, testCase.ExpectedStatusCode, recorder.Code)

			var responce Responce
			if assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &responce)) {
				assert.Equal(t, testCase.MockReturnId, responce.Id)
				assert.Equal(t, testCase.ExpectedErrorMsg, responce.Error)
			}
		})
	}
}
