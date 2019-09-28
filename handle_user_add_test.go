package main

import (
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

type AfterTime struct {
	created time.Time
}

func (after AfterTime) Match(value driver.Value) bool {
	timeValue, ok := value.(time.Time)
	if !ok {
		return false
	}

	if timeValue.Before(after.created) || timeValue.After(time.Now()) {
		return false
	}
	return true
}

func TestHandleUserAdd(t *testing.T) {
	// create server with mocked db connection
	type TestCase struct {
		Name           string
		RequestPayload string
		Code           int
		ErrorMsg       string
		ReturnedId     int
		SetupMock      func(mock sqlmock.Sqlmock, testCase *TestCase)
	}

	subtestCases := []*TestCase{
		&TestCase{
			Name:           "SimpleAdd",
			RequestPayload: `{"username": "user_1"}`,
			Code:           http.StatusOK,
			ReturnedId:     1,
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("user_1", AfterTime{time.Now()}).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		&TestCase{
			Name:           "Error request format",
			RequestPayload: `{"username": "user_1"`,
			Code:           http.StatusInternalServerError,
			ErrorMsg:       "json decoding error",
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
			},
		},
		&TestCase{
			Name:           "Add user with duplicate username",
			RequestPayload: `{"username": "user_1"}`,
			Code:           http.StatusInternalServerError,
			ErrorMsg:       "User with this username is already added",
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
				rows := sqlmock.NewRows([]string{"username"}).AddRow("user_1")
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
		},
		&TestCase{
			Name:           "Incorrect username",
			RequestPayload: `{"username": "1_user"}`,
			Code:           http.StatusInternalServerError,
			ErrorMsg:       "Username is in unsupported format",
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
			},
		},
		&TestCase{
			Name:           "Too long username",
			RequestPayload: `{"username": "u234567890123456789012345679012345"}`,
			Code:           http.StatusInternalServerError,
			ErrorMsg:       "Username is in unsupported format",
			SetupMock: func(mock sqlmock.Sqlmock, testCase *TestCase) {
			},
		},
	}
	server := NewServer(true)

	type Responce struct {
		Id    int    `json:id`
		Error string `json:error`
	}
	for _, subtest := range subtestCases {
		t.Run(subtest.Name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal("Can not create stub database connection: ", err)
			}
			defer db.Close()
			server.DB = ServerDB{db}

			// create request and response
			requestData := strings.NewReader(subtest.RequestPayload)
			request, err := http.NewRequest(http.MethodPost, "/users/add", requestData)
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()

			// our expected query to sql
			subtest.SetupMock(mock, subtest)

			handler := server.handleUserAdd()
			handler.ServeHTTP(recorder, request)

			assert.NoError(t, mock.ExpectationsWereMet())
			assert.Equal(t, subtest.Code, recorder.Code)

			var responce Responce
			if assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &responce)) {
				assert.Equal(t, subtest.ReturnedId, responce.Id)
				assert.Equal(t, subtest.ErrorMsg, responce.Error)
			}
		})
	}
}
