package postgres

import (
	"Yo/src/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	UserExistsErr           = errors.New("User already exists")
	FriendshipExistsErr     = errors.New("FriendShip already exists")
	FriendshipDontExistsErr = errors.New("FriendShip doesn`t exists")
)

type PostgresDb struct {
	Pool *pgxpool.Pool
}

func (d PostgresDb) NewUser(chatId int64, userName, tag string, ctx context.Context) error {
	tx, err := d.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var userExists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE chatid = $1);`,
		chatId).Scan(&userExists)
	if err != nil {
		return err
	}
	if userExists {
		return UserExistsErr
	}
	err = d.Pool.QueryRow(context.Background(), `INSERT INTO Users (ChatId, Name ,Tag ) VALUES ($1, $2 , $3);`,
		chatId, userName, tag).Scan()
	if err != nil {
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (d PostgresDb) GetState(chatId int64, ctx context.Context) (int, error) {
	tx, err := d.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var state int
	err = d.Pool.QueryRow(ctx, `SELECT State FROM Users WHERE chatId = $1;`, chatId).Scan(&state)
	if err != nil {
		return 0, err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}
	return state, nil
}

func (d PostgresDb) GetData(chatId int64, ctx context.Context) (*models.User, error) {
	tx, err := d.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var user models.User
	err = d.Pool.QueryRow(context.Background(), `SELECT Name, Tag FROM Users WHERE ChatId = $1;`, chatId).Scan(
		&user.Name,
		&user.Tag,
	)
	if err != nil {
		return nil, err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d PostgresDb) GetChatIdByTag(tag string, ctx context.Context) (int64, error) {
	tx, err := d.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var chatId int64
	err = d.Pool.QueryRow(context.Background(), `SELECT ChatId FROM Users WHERE Tag = $1;`, tag).Scan(
		&chatId,
	)
	if err != nil {
		return 0, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}
	return chatId, nil
}

func (d PostgresDb) AddFriend(chatId int64, Tag string, ctx context.Context) error {

	friendID, err := d.GetChatIdByTag(Tag, ctx)
	if err != nil {
		return err
	}
	tx, err := d.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var friendshipExists bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM friends WHERE (chatId = $1 AND scndchatid = $2) OR (chatId = $2 AND ScndChatId = $1));
	`, chatId, friendID).Scan(&friendshipExists)
	if err != nil {
		return err
	}
	if friendshipExists {
		return FriendshipExistsErr
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO friends (chatId, scndchatid) VALUES ($1, $2);
	`, chatId, friendID)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (d PostgresDb) DelFriend(chatId int64, Tag string, ctx context.Context) error {
	friendID, err := d.GetChatIdByTag(Tag, ctx)
	if err != nil {
		return err
	}
	tx, err := d.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var friendshipExists bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM friends WHERE (chatId = $1 AND scndchatid = $2) OR (chatId = $2 AND ScndChatId = $1));
	`, chatId, friendID).Scan(&friendshipExists)
	if err != nil {
		return err
	}
	if !friendshipExists {
		return FriendshipDontExistsErr
	}

	_, err = tx.Exec(ctx,
		`DELETE FROM Friends
		WHERE (chatId = $1 AND ScndChatId = $2)
		OR (chatId = $2 AND ScndChatId = $1);`,
		chatId, friendID)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil

}

func (d PostgresDb) GetAllUsers(ctx context.Context) (map[int64]interface{}, error) {
	tx, err := d.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	chatIds := make(map[int64]interface{})
	rows, err := d.Pool.Query(ctx, `SELECT chatId FROM Users;`)
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

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return chatIds, nil
}

func (d PostgresDb) GetFriends(chatId int64, ctx context.Context) (map[int64]interface{}, error) {
	tx, err := d.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	friends := make(map[int64]interface{})
	rows, err := d.Pool.Query(ctx,
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

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return friends, nil
}

func (d PostgresDb) UpdateState(chatId int64, state int, ctx context.Context) error {
	tx, err := d.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = d.Pool.Exec(ctx, `SELECT State FROM Users WHERE chatId = $1`, chatId)
	if err != nil {
		return err
	}

	_, err = d.Pool.Exec(ctx, `UPDATE Users SET State = $1 WHERE chatId = $2;`, state, chatId)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (d PostgresDb) UpdateName(chatId int64, name string, ctx context.Context) error {
	tx, err := d.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = d.Pool.Exec(ctx, `SELECT name FROM users WHERE chatid = $1;`, chatId)
	if err != nil {
		return err
	}
	_, err = d.Pool.Exec(ctx, `UPDATE Users SET Name = $1 WHERE chatId = $2;`, name, chatId)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func NewPostgresDb(Addr string, ctx context.Context) (*PostgresDb, error) {
	pool, err := pgxpool.New(ctx, Addr)
	if err != nil {
		return nil, err
	}

	return &PostgresDb{Pool: pool}, nil
}
