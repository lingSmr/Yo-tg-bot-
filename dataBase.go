package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

// TODO: доделать интерфейс
type DataBaseImpl interface {
	NewUser(chatId int, UserName string) (id int, err error)
	GetState(chatId int) (int, error)
	UpdateState(chatId, state int) error
	GetFriends(chatId int) ([]int, error)
}

type DataBase struct {
	Pool *pgxpool.Pool
}

func (d *DataBase) NewUser(chatId int, userName, tag string) (id int, err error) {
	row := d.Pool.QueryRow(context.Background(), `INSERT INTO Users (ChatId, Name ,Tag ) VALUES ($1, $2 , $3);`, chatId, userName, tag)
	err = row.Scan()
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (d *DataBase) GetState(chatId int) (int, error) {
	var state int
	row := d.Pool.QueryRow(context.Background(), `SELECT State FROM Users WHERE chatId = $1;`, chatId)
	err := row.Scan(
		&state,
	)
	if err != nil {
		return 0, err
	}
	return state, nil
}

func (d *DataBase) GetData(chatId int) (*User, error) {
	var user User
	row := d.Pool.QueryRow(context.Background(), `SELECT Name, Tag FROM Users WHERE ChatId = $1;`, chatId)
	err := row.Scan(
		&user.Name,
		&user.Tag,
	)
	if err != nil {
		return &User{}, err
	}
	return &user, nil
}

func (d *DataBase) GetChatIdByTag(tag string) (int, error) {
	var chatId int
	row := d.Pool.QueryRow(context.Background(), `SELECT ChatId FROM Users WHERE Tag = $1;`, tag)
	err := row.Scan(&chatId)
	if err != nil {
		return 0, err
	}
	return chatId, nil
}

func (d *DataBase) AddFriend(chatId int, Tag string) (int, error) {
	var ID int8
	friendChatId, err := d.GetChatIdByTag(Tag)
	if err != nil {
		return 0, err
	}

	row := d.Pool.QueryRow(context.Background(), `INSERT INTO Friends (ChatId, ScndChatId) VALUES ($1, $2) RETURNING Id;`,
		chatId, friendChatId)
	err = row.Scan(&ID)
	if err != nil && ID != 0 {
		return 0, err
	}

	return int(ID), nil
}

// Возвращает 1 если успешно удаленно , 0 если неуспешно
func (d *DataBase) DelFriend(chatId int, Tag string) (int, error) {
	friendChatId, err := d.GetChatIdByTag(Tag)
	if err != nil {
		return 0, err
	}

	d.Pool.QueryRow(context.Background(),
		`DELETE FROM Friends
		WHERE (chatId = $1 AND ScndChatId = $2)
   		OR (chatId = $2 AND ScndChatId = $1);`,
		chatId, friendChatId)

	return 1, nil
}

func (d *DataBase) GetFriends(chatId int) (map[int]interface{}, error) {
	friends := make(map[int]interface{})
	rows, err := d.Pool.Query(context.Background(),
		`SELECT ScndChatId FROM Friends WHERE chatId = $1
		UNION
		SELECT chatId FROM Friends WHERE ScndChatId = $1;`, chatId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var FriendchatId int
		err := rows.Scan(&FriendchatId)
		if err != nil {
			return nil, err
		}
		friends[FriendchatId] = nil
	}
	return friends, nil
}

func (d *DataBase) UpdateState(chatId, state int) error {
	_, err := d.Pool.Exec(context.Background(), `UPDATE Users SET State = $1 WHERE chatId = $2;`, state, chatId)
	if err != nil {
		return err
	}
	return nil
}

func (d *DataBase) UpdateName(chatId int, name string) error {
	_, err := d.Pool.Exec(context.Background(), `UPDATE Users SET Name = $1 WHERE chatId = $2;`, name, chatId)
	if err != nil {
		return err
	}
	return nil
}

func NewDatabase(Addr string) (*DataBase, error) {
	const op = "dataBase:NewDataBase"
	pool, err := pgxpool.Connect(context.Background(), Addr)
	if err != nil {
		log.Fatalf("%s:%s", op, err)
	}

	return &DataBase{Pool: pool}, nil
}
