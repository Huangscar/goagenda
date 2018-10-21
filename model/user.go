package model

import (
	"strings"
)

type UserModel interface {
	AddUser(user User)
	GetUserByName(name string) User
	DeleteUser(name string)
}

type UserDB struct {
	Data []User
	Database
}

type User struct {
	Name      string `json:"user"`
	Email     string `json:"email"`
	Telephone string `json:"telephone"`
	Password  string `json:"password"`
	Salt      string `json:"salt"`
}

var userDB = UserDB{Database: Database{schema: "User"}}

func (m *UserDB) GetUserByName(name string) User {
	for _, item := range m.Data {
		if strings.ToLower(item.Name) == strings.ToLower(name) {
			return item
		}
	}
	return User{}
}

func (m *UserDB) AddUser(user User) {
	m.isDirty = true
	m.Data = append(m.Data, user)
}

func ReleaseUserModel() {
	userDB.releaseModel(&userDB.Data)
}

func (m *Manager) User() UserModel {
	if userDB.isInit == false {
		userDB.initModel(&userDB.Data)
	}
	return &userDB
}

func (m *UserDB) DeleteUser(name string) {
	m.isDirty = true
	index := -1
	for i := 0; i < len(m.Data); i++ {
		if m.Data[i].Name == name {
			index = i
			break
		}
	}
	if index != -1 {
		m.Data = append(m.Data[:index], m.Data[index+1:]...)
	}
}
