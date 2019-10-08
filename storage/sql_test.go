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
		"messages",
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
func TestGetUserChats(t *testing.T) {
	sqlStorage, teardown := getSqlStorage(t, []string{"users_chats", "chats", "users"})
	defer teardown()
	_, err := sqlStorage.Exec(`INSERT INTO users(id, username, created_at) VALUES 
		(1, "user1", "2019-01-01 10:00:00"),
		(2, "user2", "2019-01-01 10:00:00")`)
	if err != nil {
		t.Fatal("Insert user failed: ", err)
	}
	chats := []*Chat{
		&Chat{11, "chat1", time.Date(2019, time.January, 1, 10, 0, 0, 0, time.UTC), nil},
		&Chat{12, "chat2", time.Date(2019, time.January, 1, 11, 0, 0, 0, time.UTC), nil},
		&Chat{13, "chat3", time.Date(2019, time.January, 1, 12, 0, 0, 0, time.UTC), nil},
		&Chat{14, "chat4", time.Date(2019, time.January, 1, 13, 0, 0, 0, time.UTC), nil},
	}
	for _, chat := range chats {
		_, err := sqlStorage.Exec(`INSERT INTO chats(id, name, created_at) VALUES (?, ?, ?)`, chat.Id, chat.Name, chat.CreatedAt)
		if err != nil {
			t.Fatal("Insert chats failed: ", err)
		}
	}
	_, err = sqlStorage.Exec(`INSERT INTO users_chats(user_id, chat_id) VALUES
		(1, 11),
		(1, 12),
		(1, 14),
		(2, 12),
		(2, 13)`)
	if err != nil {
		t.Fatal("Insert users_chats failed: ", err)
	}

	assertChats := func(expected, actual []*Chat) {
		assert.Equal(t, len(expected), len(actual))
		for i, chat := range expected {
			assert.Equal(t, chat.Id, actual[i].Id)
			assert.Equal(t, chat.Name, actual[i].Name)
			assert.Equal(t, chat.CreatedAt, actual[i].CreatedAt)
		}

	}
	user1Chats, err := sqlStorage.GetUserChats(1)
	assert.NoError(t, err)
	assertChats([]*Chat{chats[0], chats[1], chats[3]}, user1Chats)

	user2Chats, err := sqlStorage.GetUserChats(2)
	assert.NoError(t, err)
	assertChats([]*Chat{chats[1], chats[2]}, user2Chats)

}
func TestGetUserIdsFromChat(t *testing.T) {
	sqlStorage, teardown := getSqlStorage(t, []string{"users_chats", "chats", "users"})
	defer teardown()
	_, err := sqlStorage.Exec(`INSERT INTO users(id, username, created_at) VALUES 
		(1, "user1", "2019-01-01 10:00:00"),
		(2, "user2", "2019-01-01 10:00:00"),
		(3, "user3", "2019-01-01 10:00:00"),
		(4, "user4", "2019-01-01 10:00:00")`)
	if err != nil {
		t.Fatal("Insert user failed: ", err)
	}
	_, err = sqlStorage.Exec(`INSERT INTO chats(id, name, created_at) VALUES 
		(10, "chat1", "2019-01-01 10:00:00"),
		(11, "chat2", "2019-01-01 10:00:00")`)
	if err != nil {
		t.Fatal("Insert chats failed: ", err)
	}
	_, err = sqlStorage.Exec(`INSERT INTO users_chats(user_id, chat_id) VALUES
		(1, 10),
		(2, 10),
		(4, 10),
		(3, 11),
		(4, 11)`)
	if err != nil {
		t.Fatal("Insert users_chats failed: ", err)
	}

	userIds, err := sqlStorage.getUserIdsFromChat(10)
	assert.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 4}, userIds)

	userIds, err = sqlStorage.getUserIdsFromChat(11)
	assert.NoError(t, err)
	assert.Equal(t, []int64{3, 4}, userIds)

}
func TestSortChatsByLastMessage(t *testing.T) {
	sqlStorage, teardown := getSqlStorage(t, []string{"messages", "chats", "users"})
	defer teardown()
	_, err := sqlStorage.Exec(`INSERT INTO users(id, username, created_at) VALUES (10, "user", "2019-01-01 10:00:00")`)
	if err != nil {
		t.Fatal("Insert user failed: ", err)
	}
	chats := []*Chat{
		&Chat{5, "chat5", time.Date(2019, time.January, 1, 14, 0, 0, 0, time.UTC), nil},
		&Chat{4, "chat4", time.Date(2019, time.January, 1, 13, 0, 0, 0, time.UTC), nil},
		&Chat{3, "chat3", time.Date(2019, time.January, 1, 12, 0, 0, 0, time.UTC), nil},
		&Chat{2, "chat2", time.Date(2019, time.January, 1, 11, 0, 0, 0, time.UTC), nil},
		&Chat{1, "chat1", time.Date(2019, time.January, 1, 10, 0, 0, 0, time.UTC), nil},
	}
	for _, chat := range chats {
		_, err := sqlStorage.Exec(`INSERT INTO chats(id, name, created_at) VALUES (?, ?, ?)`, chat.Id, chat.Name, chat.CreatedAt)
		if err != nil {
			t.Fatal("Insert chats failed: ", err)
		}
	}

	testOrder := func(expected []int64) {
		actualOrder := make([]int64, len(chats))
		for i, chat := range chats {
			actualOrder[i] = chat.Id
		}
		assert.Equal(t, expected, actualOrder)
	}

	sqlStorage.sortChatsByLastMessage(chats)
	testOrder([]int64{1, 2, 3, 4, 5})

	_, err = sqlStorage.Exec(`INSERT INTO messages(chat_id, author_id, text, created_at) VALUES(2, 10, "", "2019-01-01 16:00:00")`)
	if err != nil {
		t.Fatal("Insert message failed: ", err)
	}
	sqlStorage.sortChatsByLastMessage(chats)
	testOrder([]int64{1, 3, 4, 5, 2})

	_, err = sqlStorage.Exec(`INSERT INTO messages(chat_id, author_id, text, created_at) VALUES(1, 10, "", "2019-01-01 15:00:00")`)
	if err != nil {
		t.Fatal("Insert message failed: ", err)
	}
	sqlStorage.sortChatsByLastMessage(chats)
	testOrder([]int64{3, 4, 5, 1, 2})

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
func TestIsUserInChat(t *testing.T) {
	sqlStorage, teardown := getSqlStorage(t, []string{"users_chats", "users", "chats"})
	defer teardown()
	userNotInChat, err := sqlStorage.AddUser("user_2")
	if err != nil {
		t.Fatal("Can not add user: ", err)
	}
	userInChat, err := sqlStorage.AddUser("user_1")
	if err != nil {
		t.Fatal("Can not add user: ", err)
	}
	chatId, err := sqlStorage.AddChat("chat_1", []int64{userInChat})
	if err != nil {
		t.Fatal("Can not create chat: ", err)
	}

	isInChat, err := sqlStorage.IsUserInChat(userInChat, chatId)
	assert.NoError(t, err)
	assert.True(t, isInChat)

	isInChat, err = sqlStorage.IsUserInChat(userNotInChat, chatId)
	assert.NoError(t, err)
	assert.False(t, isInChat)
}
func TestAddMessage(t *testing.T) {
	sqlStorage, teardown := getSqlStorage(t, []string{"messages", "users_chats", "users", "chats"})
	defer teardown()
	//add users
	userNotInChat, err := sqlStorage.AddUser("user_2")
	if err != nil {
		t.Fatal("Can not add user: ", err)
	}
	userInChat, err := sqlStorage.AddUser("user_1")
	if err != nil {
		t.Fatal("Can not add user: ", err)
	}
	chatId, err := sqlStorage.AddChat("chat_1", []int64{userInChat})
	if err != nil {
		t.Fatal("Can not create chat: ", err)
	}
	type TestCase struct {
		TestName   string
		ChatId     int64
		AuthorId   int64
		Text       string
		ShouldFail bool
	}
	testCases := []TestCase{
		TestCase{
			TestName: "Add message",
			ChatId:   chatId,
			AuthorId: userInChat,
			Text:     "Hello, World!",
		},
		TestCase{
			TestName:   "User not in chat",
			ChatId:     chatId,
			AuthorId:   userNotInChat,
			Text:       "Hello, Hijackers!",
			ShouldFail: true,
		},
		TestCase{
			TestName:   "Nonexistent user",
			ChatId:     chatId,
			AuthorId:   1203,
			Text:       "I'm not exist",
			ShouldFail: true,
		},
		TestCase{
			TestName:   "Nonexistent chat",
			ChatId:     2313,
			AuthorId:   userInChat,
			Text:       "Helloo",
			ShouldFail: true,
		},
	}
	messagesAlreadyCreated := 0
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			timeBeforeInserting := time.Now()
			messageId, err := sqlStorage.AddMessage(testCase.AuthorId, testCase.ChatId, testCase.Text)
			if testCase.ShouldFail {
				assert.Error(t, err)
				var messagesExist int
				err = sqlStorage.QueryRow(`SELECT COUNT(*) FROM messages`).Scan(&messagesExist)
				assert.NoError(t, err)
				assert.Equal(t, messagesAlreadyCreated, messagesExist)
				return
			} else {
				if err != nil {
					t.Error("AddMessage failed: ", err)
					return
				}
			}
			messagesAlreadyCreated++

			var chatId, authorId int64
			var text string
			var createdAt time.Time
			err = sqlStorage.QueryRow("SELECT chat_id, author_id, text, created_at FROM messages WHERE id = ?", messageId).
				Scan(&chatId, &authorId, &text, &createdAt)
			if err != nil {
				t.Fatal("Select message failed: ", err)
			}
			assert.Equal(t, testCase.ChatId, chatId)
			assert.Equal(t, testCase.AuthorId, authorId)
			assert.Equal(t, testCase.Text, text)
			assert.False(t, createdAt.Before(timeBeforeInserting) || createdAt.After(time.Now()))

			var messagesExist int
			err = sqlStorage.QueryRow(`SELECT COUNT(*) FROM messages`).Scan(&messagesExist)
			assert.NoError(t, err)
			assert.Equal(t, messagesAlreadyCreated, messagesExist)
		})
	}
}
func TestGetMessagesFromChat(t *testing.T) {
	sqlStorage, teardown := getSqlStorage(t, []string{"messages", "chats", "users"})
	defer teardown()
	_, err := sqlStorage.Exec(`INSERT INTO users(id, username, created_at) VALUES
		(1, "my_favorite", "2019-01-01 10:00:00"),
		(2, "another_one", "2019-01-01 10:00:00")`)
	if err != nil {
		t.Fatal("Insert into users failed: ", err)
	}
	_, err = sqlStorage.Exec(`INSERT INTO chats(id, name, created_at) VALUES
		(10, "chat_1", "2019-01-01 10:00:00"),
		(11, "chat_2", "2019-01-01 10:00:00")`)
	if err != nil {
		t.Fatal("Insert into chats failed: ", err)
	}
	expectedMessages := []*Message{
		&Message{
			Id: 21, ChatId: 10, AuthorId: 1,
			Text:      "Hello",
			CreatedAt: time.Date(2019, time.January, 1, 10, 0, 0, 0, time.UTC),
		},
		&Message{
			Id: 23, ChatId: 10, AuthorId: 1,
			Text:      "I can travel in time",
			CreatedAt: time.Date(2019, time.January, 1, 10, 0, 30, 0, time.UTC),
		},
		&Message{
			Id: 22, ChatId: 10, AuthorId: 2,
			Text:      "Hello, how are you?",
			CreatedAt: time.Date(2019, time.January, 1, 10, 1, 0, 0, time.UTC),
		},
		&Message{
			Id: 24, ChatId: 10, AuthorId: 2,
			Text:      "Are you kidding?",
			CreatedAt: time.Date(2019, time.January, 1, 10, 1, 30, 0, time.UTC),
		},
		&Message{
			Id: 25, ChatId: 11, AuthorId: 1,
			Text:      "Another chat",
			CreatedAt: time.Date(2019, time.January, 1, 9, 0, 0, 0, time.UTC),
		},
	}
	for _, message := range expectedMessages {
		_, err := sqlStorage.Exec(`INSERT INTO messages(id, chat_id, author_id, text, created_at) VALUES(?, ?, ?, ?, ?)`,
			message.Id, message.ChatId, message.AuthorId, message.Text, message.CreatedAt)
		if err != nil {
			t.Fatal("Insert into messages failed: ", err)
		}
	}
	// last messages from another chat
	expectedMessages = expectedMessages[:len(expectedMessages)-1]
	actualMessages, err := sqlStorage.GetMessagesFromChat(10)
	if err != nil {
		t.Fatal("GetMessagesFromChat failed: ", err)
	}
	assert.Equal(t, expectedMessages, actualMessages)

}
