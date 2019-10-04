package storage

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

var sqlStorageInstance SqlStorage

func TestMain(m *testing.M) {
	sqlStorage, teardown, err := openDb()
	if err != nil {
		fmt.Println("Error while setuping db: ", err)
	}
	defer teardown()
	sqlStorageInstance = sqlStorage
	os.Exit(m.Run())
}

func openDb() (sqlStorage SqlStorage, teardown func(), err error) {
	teardown = func() {}
	tempFile, err := ioutil.TempFile("", "sql-db-test-db-")
	if err != nil {
		err = fmt.Errorf("Creation temp file for db failed: %s", err)
		return
	}
	tempFile.Close()
	db, err := sql.Open("sqlite3", tempFile.Name())
	if err != nil {
		err = fmt.Errorf("Open db failed: %s", err)
		return
	}
	sqlStorage = SqlStorage{db}
	err = sqlStorage.Setup()
	if err != nil {
		err = fmt.Errorf("Setup failed: %s", err)
		return
	}
	teardown = func() {
		db.Close()
		os.Remove(tempFile.Name())
	}
	return
}

func getSqlStorage(t *testing.T, clearTables []string) (SqlStorage, func()) {
	sqlStorage := sqlStorageInstance

	return sqlStorage, func() {
		for _, tableName := range clearTables {
			_, err := sqlStorage.Exec("DELETE FROM " + tableName)
			if err != nil {
				t.Fatal("Delete failed: ", err)
			}
		}
	}
}

func TestSetup(t *testing.T) {
	sqlStorage, teardown := getSqlStorage(t, []string{})
	defer teardown()
	rows, err := sqlStorage.Query(`SELECT name FROM sqlite_master WHERE type = "table"`)
	if err != nil {
		t.Fatal("Failed to query table names: ", err)
	}
	defer rows.Close()
	tablesShouldExist := []string{
		"users",
		"chats",
		"users_chats",
	}
	tablesPresented := make([]string, 0, len(tablesShouldExist))
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			t.Fatal("Scan table name failed: ", err)
		}
		tablesPresented = append(tablesPresented, tableName)
	}
	sort.Strings(tablesPresented)
	sort.Strings(tablesShouldExist)
	assert.Equal(t, tablesShouldExist, tablesPresented)
}

func TestIsUserExists(t *testing.T) {
	sqlStorage, teardown := getSqlStorage(t, []string{"users"})
	defer teardown()
	_, err := sqlStorage.Exec(`INSERT INTO users(username, created_at) VALUES ("user1", "2019-01-01 10:00:00")`)
	if err != nil {
		t.Fatal("Insert failed: ", err)
	}

	isExist, err := sqlStorage.IsUserExists("user1")
	if err != nil {
		t.Fatal("IsUserExists failed: ", err)
	}
	assert.True(t, isExist)

	isExist, err = sqlStorage.IsUserExists("not_user")
	if err != nil {
		t.Fatal("IsUserExists failed: ", err)
	}
	assert.False(t, isExist)
}

func TestAreUsersExist(t *testing.T) {
	sqlStorage, teardown := getSqlStorage(t, []string{"users"})
	defer teardown()
	_, err := sqlStorage.Exec(`INSERT INTO users(id, username, created_at) VALUES 
		(1, "user1", "2019-01-01 10:00:00"),
		(2, "user2", "2019-01-01 10:00:00"),
		(3, "user3", "2019-01-01 10:00:00")`)
	if err != nil {
		t.Fatal("Insert failed: ", err)
	}

	cases := []struct {
		Ids      []int64
		AreExist bool
	}{
		{[]int64{1, 2, 3}, true},
		{[]int64{1, 3}, true},
		{[]int64{3}, true},
		{[]int64{1, 2, 4}, false},
		{[]int64{5}, false},
	}
	for _, testCase := range cases {
		testCase := testCase
		t.Run(fmt.Sprintf("Users: %v", testCase.Ids), func(t *testing.T) {
			areExist, err := sqlStorage.AreUsersExistByIds(testCase.Ids)
			assert.NoError(t, err)
			assert.Equal(t, testCase.AreExist, areExist)
		})
	}

}

func TestAddUser(t *testing.T) {
	sqlStorage, teardown := getSqlStorage(t, []string{"users"})
	defer teardown()
	type UserData struct {
		Id   int64
		Name string
	}
	users := []UserData{
		UserData{0, "test_user"},
		UserData{0, "some_another"},
		UserData{0, "one_spare"},
	}
	timeBeforeInserting := time.Now()
	usersMap := make(map[int64]UserData)
	for _, user := range users {
		id, err := sqlStorage.AddUser(user.Name)
		assert.NoError(t, err, "Add user failed")
		user.Id = id
		usersMap[user.Id] = user
	}
	rows, err := sqlStorage.Query("SELECT id, username, created_at FROM users")
	assert.NoError(t, err, "SELECT query failed")
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name string
		var createdAt time.Time
		err := rows.Scan(&id, &name, &createdAt)
		assert.NoError(t, err, "Scan failed")
		user, ok := usersMap[id]
		assert.True(t, ok, "Retrieve from users map failed")
		assert.Equal(t, user.Name, name)
		assert.False(t, createdAt.Before(timeBeforeInserting) || createdAt.After(time.Now()))
	}

	var numberOfUsers int
	err = sqlStorage.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&numberOfUsers)
	assert.NoError(t, err)
	assert.Equal(t, len(users), numberOfUsers, "Number of inserted users not equal to number of users expected to be inserted")

}
func TestIsChatExists(t *testing.T) {
	sqlStorage, teardown := getSqlStorage(t, []string{"chats"})
	defer teardown()
	_, err := sqlStorage.Exec(`INSERT INTO chats(name, created_at) VALUES ("chat1", "2019-01-01 10:00:00")`)
	if err != nil {
		t.Fatal("Insert failed: ", err)
	}

	isExist, err := sqlStorage.IsChatExists("chat1")
	if err != nil {
		t.Fatal("IsChatExists failed: ", err)
	}
	assert.True(t, isExist)

	isExist, err = sqlStorage.IsChatExists("not_chat")
	if err != nil {
		t.Fatal("IsChatExists failed: ", err)
	}
	assert.False(t, isExist)
}

func TestAddChat(t *testing.T) {
	sqlStorage, teardown := getSqlStorage(t, []string{"users_chats", "users", "chats"})
	defer teardown()
	//add users
	_, err := sqlStorage.Exec(`INSERT INTO users(id, username, created_at) VALUES 
		(1, "my_favorite", "2019-01-01 10:00:00"),
		(2, "another_one", "2019-01-01 10:00:00"),
		(3, "one_I_dont_really_like", "2019-01-01 10:00:00")`)
	if err != nil {
		t.Fatal("Insert users failed: ", err)
	}
	type TestCase struct {
		TestName   string
		ChatName   string
		UserIds    []int64
		ShouldFail bool
	}
	testCases := []TestCase{
		TestCase{
			TestName: "Add chat with all 3 users",
			ChatName: "Telegram_news",
			UserIds:  []int64{1, 2, 3},
		},
		TestCase{
			TestName: "Add chat with single user",
			ChatName: "ethereum_future",
			UserIds:  []int64{2},
		},
		TestCase{
			TestName:   "Chat with non existent user",
			ChatName:   "ethereum_future",
			UserIds:    []int64{1, 3, 5},
			ShouldFail: true,
		},
	}
	var chatsCreated = 0
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			timeBeforeInserting := time.Now()
			chatId, err := sqlStorage.AddChat(testCase.ChatName, testCase.UserIds)
			if testCase.ShouldFail {
				assert.Error(t, err)
				var chatsExist int
				err = sqlStorage.QueryRow(`SELECT COUNT(*) FROM chats`).Scan(&chatsExist)
				assert.NoError(t, err)
				assert.Equal(t, chatsCreated, chatsExist)
				return
			} else {
				if err != nil {
					t.Fatal("AddChat failed, but should not")
				}
			}
			chatsCreated++

			var chatName string
			var chatCreatedAt time.Time
			err = sqlStorage.QueryRow("SELECT name, created_at FROM chats WHERE id = ?", chatId).
				Scan(&chatName, &chatCreatedAt)
			if err != nil {
				t.Fatal("Select chat failed: ", err)
			}
			assert.Equal(t, testCase.ChatName, chatName)
			assert.False(t, chatCreatedAt.Before(timeBeforeInserting) || chatCreatedAt.After(time.Now()))

			rows, err := sqlStorage.Query("SELECT user_id FROM users_chats WHERE chat_id = ?", chatId)
			if err != nil {
				t.Fatal("Query users_chats failed: ", err)
			}
			defer rows.Close()
			for rows.Next() {
				var userId int64
				err := rows.Scan(&userId)
				if !assert.NoError(t, err) {
					continue
				}
				assert.Contains(t, testCase.UserIds, userId)
			}
			var numberOfUsers int
			err = sqlStorage.QueryRow(`SELECT COUNT(*) FROM users_chats WHERE chat_id = ?`, chatId).Scan(&numberOfUsers)
			assert.NoError(t, err)
			assert.Equal(t, len(testCase.UserIds), numberOfUsers)
		})
	}
}
