package main

import "context"

//обращение к таблице friends происходит посредством связи ChatId
//ChatId выступает в роли primary key , так как у каждого пользователя уникален
type User struct {
	ChatId int64
	Name   string
	Tag    string
	State  int
}

type DataBase interface {
	NewUser(chatId int64, userName, tag string, ctx context.Context) error
	UpdateName(chatId int64, name string, ctx context.Context) error
	GetState(chatId int64, ctx context.Context) (int, error)
	UpdateState(chatId int64, state int, ctx context.Context) error
	GetChatIdByTag(tag string, ctx context.Context) (int64, error)
	GetData(chatId int64, ctx context.Context) (*User, error)
	AddFriend(chatId int64, Tag string, ctx context.Context) error
	DelFriend(chatId int64, Tag string, ctx context.Context) error
	GetFriends(chatId int64, ctx context.Context) (map[int64]interface{}, error)
	GetAllUsers(ctx context.Context) (map[int64]interface{}, error)
}
