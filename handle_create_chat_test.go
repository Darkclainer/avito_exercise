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

func TestHandleCreateChat(t *testing.T) {
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
			TestName:           "Add chat with one user",
			RequestBody:        `{"name": "chat_1", "users": [1]}`,
			ExpectedStatusCode: http.StatusOK,
			MockReturnId:       2,
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery("SELECT id FROM users WHERE id IN (1)").WillReturnRows(rows)
				mock.ExpectExec("INSERT INTO chats(name, created_at)").
					WithArgs("chat_1", MatchTime{Start: time.Now()}).
					WillReturnResult(sqlmock.NewResult(testCase.MockReturnId, 0))
				mock.ExpectExec("INSERT INTO users_chats (user_id, chat_id) VALUES").
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
		},
		&TestCase{
			TestName:           "Add chat with three users",
			RequestBody:        `{"name": "chat_1", "users": [1, 2, 3]}`,
			ExpectedStatusCode: http.StatusOK,
			MockReturnId:       4,
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2).AddRow(3)
				mock.ExpectQuery("SELECT id FROM users WHERE id IN (1, 2, 3)").WillReturnRows(rows)
				mock.ExpectExec("INSERT INTO chats(name, created_at)").
					WithArgs("chat_1", MatchTime{Start: time.Now()}).
					WillReturnResult(sqlmock.NewResult(testCase.MockReturnId, 0))
				mock.ExpectExec("INSERT INTO users_chats (user_id, chat_id) VALUES").
					WithArgs(1, 4).
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectExec("INSERT INTO users_chats (user_id, chat_id) VALUES").
					WithArgs(2, 4).
					WillReturnResult(sqlmock.NewResult(1, 0))
				mock.ExpectExec("INSERT INTO users_chats (user_id, chat_id) VALUES").
					WithArgs(3, 4).
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
		},
		&TestCase{
			TestName:           "Add chat with nonexistent user",
			RequestBody:        `{"name": "chat_1", "users": [1, 123]}`,
			ExpectedErrorMsg:   "nonexistent user",
			ExpectedStatusCode: http.StatusInternalServerError,
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery("SELECT id FROM users WHERE id IN (1, 123)").WillReturnRows(rows)
			},
		},
		&TestCase{
			TestName:           "Add chat with duplicated name",
			RequestBody:        `{"name": "chat_1", "users": [1, 2]}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "chat with the same name is already exists",
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2)
				mock.ExpectQuery("SELECT id FROM users WHERE id IN (1, 2, 3)").WillReturnRows(rows)
				rows = sqlmock.NewRows([]string{"name"}).AddRow("chat_1")
				mock.ExpectQuery("SELECT name FROM chats WHERE name").WillReturnRows(rows)
			},
		},
		&TestCase{
			TestName:           "Query without users",
			RequestBody:        `{"name": "chat"}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Query without chat name",
			RequestBody:        `{"users": [1, 2]}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Query with empty users list",
			RequestBody:        `{"name": "chat", "users": []}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Query with invalid chat name",
			RequestBody:        `{"name": "12chat", "users": [1, 2]}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Query with invalid user id type",
			RequestBody:        `{"name": "chat", "users": [1, "sdf"]}`,
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMsg:   "invalid input",
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
			},
		},
		&TestCase{
			TestName:           "Query with below zero user id",
			RequestBody:        `{"name": "chat", "users": [1, -1]}`,
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
			request, err := http.NewRequest(http.MethodPost, "/chats/add", requestData)
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()

			testCase.SetupMock(mock, testCase)

			handler := server.handleChatAdd()
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
