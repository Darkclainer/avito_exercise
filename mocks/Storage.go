// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	storage "github.com/Darkclainer/avito_exercise/storage"
	mock "github.com/stretchr/testify/mock"
)

// Storage is an autogenerated mock type for the Storage type
type Storage struct {
	mock.Mock
}

// AddChat provides a mock function with given fields: chatname, userIds
func (_m *Storage) AddChat(chatname string, userIds []int64) (int64, error) {
	ret := _m.Called(chatname, userIds)

	var r0 int64
	if rf, ok := ret.Get(0).(func(string, []int64) int64); ok {
		r0 = rf(chatname, userIds)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, []int64) error); ok {
		r1 = rf(chatname, userIds)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AddMessage provides a mock function with given fields: chatId, authorId, text
func (_m *Storage) AddMessage(chatId int64, authorId int64, text string) (int64, error) {
	ret := _m.Called(chatId, authorId, text)

	var r0 int64
	if rf, ok := ret.Get(0).(func(int64, int64, string) int64); ok {
		r0 = rf(chatId, authorId, text)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64, int64, string) error); ok {
		r1 = rf(chatId, authorId, text)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AddUser provides a mock function with given fields: username
func (_m *Storage) AddUser(username string) (int64, error) {
	ret := _m.Called(username)

	var r0 int64
	if rf, ok := ret.Get(0).(func(string) int64); ok {
		r0 = rf(username)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(username)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AreUsersExistByIds provides a mock function with given fields: userIds
func (_m *Storage) AreUsersExistByIds(userIds []int64) (bool, error) {
	ret := _m.Called(userIds)

	var r0 bool
	if rf, ok := ret.Get(0).(func([]int64) bool); ok {
		r0 = rf(userIds)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]int64) error); ok {
		r1 = rf(userIds)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMessagesFromChat provides a mock function with given fields: chatId
func (_m *Storage) GetMessagesFromChat(chatId int64) ([]*storage.Message, error) {
	ret := _m.Called(chatId)

	var r0 []*storage.Message
	if rf, ok := ret.Get(0).(func(int64) []*storage.Message); ok {
		r0 = rf(chatId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*storage.Message)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(chatId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserChats provides a mock function with given fields: userId
func (_m *Storage) GetUserChats(userId int64) ([]*storage.Chat, error) {
	ret := _m.Called(userId)

	var r0 []*storage.Chat
	if rf, ok := ret.Get(0).(func(int64) []*storage.Chat); ok {
		r0 = rf(userId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*storage.Chat)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(userId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsChatExists provides a mock function with given fields: chatname
func (_m *Storage) IsChatExists(chatname string) (bool, error) {
	ret := _m.Called(chatname)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(chatname)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(chatname)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsUserExists provides a mock function with given fields: username
func (_m *Storage) IsUserExists(username string) (bool, error) {
	ret := _m.Called(username)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(username)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(username)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsUserInChat provides a mock function with given fields: userId, chatId
func (_m *Storage) IsUserInChat(userId int64, chatId int64) (bool, error) {
	ret := _m.Called(userId, chatId)

	var r0 bool
	if rf, ok := ret.Get(0).(func(int64, int64) bool); ok {
		r0 = rf(userId, chatId)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64, int64) error); ok {
		r1 = rf(userId, chatId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
