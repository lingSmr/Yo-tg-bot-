package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresDb struct {
	Pool *pgxpool.Pool
}

func (d PostgresDb) NewUser(chatId int64, userName, tag string, ctx context.Context) error {
	row := d.Pool.QueryRow(context.Background(), `INSERT INTO Users (ChatId, Name ,Tag ) VALUES ($1, $2 , $3);`, chatId, userName, tag)
	err := row.Scan()
	if err != nil {
		return err
	}
	return nil
}

func (d PostgresDb) GetState(chatId int64, ctx context.Context) (int, error) {
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

func (d PostgresDb) GetData(chatId int64, ctx context.Context) (*User, error) {
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

func (d PostgresDb) GetChatIdByTag(tag string, ctx context.Context) (int64, error) {
	var chatId int64
	row := d.Pool.QueryRow(context.Background(), `SELECT ChatId FROM Users WHERE Tag = $1;`, tag)
	err := row.Scan(&chatId)
	if err != nil {
		return 0, err
	}
	return chatId, nil
}

func (d PostgresDb) AddFriend(chatId int64, Tag string, ctx context.Context) error {
	var ID int8
	friendChatId, err := d.GetChatIdByTag(Tag, ctx)
	if err != nil {
		return err
	}

	row := d.Pool.QueryRow(context.Background(),
		`INSERT INTO Friends (ChatId, ScndChatId) VALUES ($1, $2) RETURNING Id;`,

		chatId, friendChatId)
	err = row.Scan(&ID)
	if err != nil || ID == 0 {
		return err
	}
	return nil
}

func (d PostgresDb) DelFriend(chatId int64, Tag string, ctx context.Context) error {
	friendChatId, err := d.GetChatIdByTag(Tag, ctx)
	if err != nil {
		return err
	}

	d.Pool.QueryRow(context.Background(),
		`DELETE FROM Friends
		WHERE (chatId = $1 AND ScndChatId = $2)
		OR (chatId = $2 AND ScndChatId = $1);`,
		chatId, friendChatId)

	return nil
}

func (d PostgresDb) GetAllUsers(ctx context.Context) (map[int64]interface{}, error) {
	chatIds := make(map[int64]interface{})
	rows, err := d.Pool.Query(context.Background(), `SELECT chatId FROM Users`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		chatIds[id] = nil
	}
	return chatIds, nil
}

func (d PostgresDb) GetFriends(chatId int64, ctx context.Context) (map[int64]interface{}, error) {
	friends := make(map[int64]interface{})
	rows, err := d.Pool.Query(context.Background(),
		`SELECT ScndChatId FROM Friends WHERE chatId = $1
		UNION
		SELECT chatId FROM Friends WHERE ScndChatId = $1;`, chatId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var FriendchatId int64
		err := rows.Scan(&FriendchatId)
		if err != nil {
			return nil, err
		}
		friends[FriendchatId] = nil
	}
	return friends, nil
}

func (d PostgresDb) UpdateState(chatId int64, state int, ctx context.Context) error {
	_, err := d.Pool.Exec(context.Background(), `UPDATE Users SET State = $1 WHERE chatId = $2;`, state, chatId)
	if err != nil {
		return err
	}
	return nil
}

func (d PostgresDb) UpdateName(chatId int64, name string, ctx context.Context) error {
	_, err := d.Pool.Exec(context.Background(), `UPDATE Users SET Name = $1 WHERE chatId = $2;`, name, chatId)
	if err != nil {
		return err
	}
	return nil
}

func NewPostgresDb(Addr string, ctx context.Context) (PostgresDb, error) {
	const op = "PostgresDb:NewPostgresDb"
	pool, err := pgxpool.Connect(ctx, Addr)
	if err != nil {
		log.Fatalf("%s:%s", op, err)
	}

	return PostgresDb{Pool: pool}, nil
}
