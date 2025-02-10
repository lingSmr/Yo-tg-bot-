package main

//обращение к таблице friends происходит посредством связи ChatId
//ChatId выступает в роли primary key , так как у каждого пользователя уникален
type UsersTable struct {
	ChatId int
	Tag    string
	Name   string
	State  int
}

//одна связь одна дружба , нет необходимости создавать обратную одной из связей
type Friends struct {
	ID         int
	ChatId     int
	ScndChatId int
}
