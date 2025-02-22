# Yo! Telegram Bot🤙

## Описание проекта

Yo! Telegram Bot — это бот, вдохновленный приложением "Bro!" из сериала "Кремниевая долина".  
Основная цель бота — дать возможность пользователям в любой момент отправить своим друзьям  **ЙОУ!** 🚀

---

## Технологии💻

В проекте использованы следующие технологии:
- **Golang v1.23.5** — основной язык разработки.
- **telegram-bot-api v5** — библиотека для взаимодействия с Telegram Bot API.
- **pgx v5** — драйвер для работы с PostgreSQL.
- **migrate v4.18.2** — инструмент для управления миграциями базы данных.
- **godotenv v1.5** - библиотека для управления конфигурацией.
- **Docker** - инструмент контейнеризации.

---

## Как работает бот❓

### Регистрация пользователя

1. Когда пользователь впервые отправляет сообщение боту, происходит проверка, зарегистрирован ли он.
```golang
if v, err := s.DataBase.GetState(ID, ctx); v == 0 || errors.Is(err, pgx.ErrNoRows) {
	name := update.Message.From.UserName
	err := s.DataBase.NewUser(ID, name, name, ctx)
	if err != nil {
   	   s.sendErr(ID)
	   continue
	}
	slog.Info("New User!", "ChatId", ID, "Username", update.Message.From.UserName)
}
```
2. Если пользователь новый выполняется **транзакция**, которая создает новую строчку в базе данных.
```golang
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
```
3. Далее идет switch case , определяющий состояние пользователя , и в зависимости от онного выполняет определенные команды.
```golang
go func(chatId int64, msgIn string, upd tgAPI.Update, state int) {
	ctxForSwitch, cancel := context.WithTimeout(context.Background(), time.Second*40)
	defer cancel()
	switch state {
	case consts.NothingState:
	case consts.StartState:
	case consts.AskNameState:
	case consts.AddFriendState:
	case consts.DelFriendState:
	case consts.UpdateNameState:
	case consts.MessageForAllState:
        }
}(ID, msg, update, st)
```
4. Новый пользователь , с состоянием **1 - StartState** , попадает в метод startSwitch.
```golang
func (s *BotServ) startSwitch(chatId int64, ctx context.Context) error {
	const op = "botMethod:startSwitch"
	botMsg := tgAPI.NewMessage(chatId, "Здравстуй дорогой пользователь!\nКак тебя зовут❔")
	s.Bot.Send(botMsg)
	err := s.DataBase.UpdateState(chatId, consts.AskNameState, ctx)
	if err != nil {
		botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
		s.Bot.Send(botMsg)
		s.logger.Error("Cant update state of user", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	return nil
}
```


### Состояния пользователя⏳


Бот работает на основе состояний (`state`). В зависимости от значения `state`, бот выполняет определенные действия.
```golang
const (
	StartState         = iota + 1 //1
	NothingState                  //2
	AskNameState                  //3
	EditNameState                 //4
	AddFriendState                //5
	DelFriendState                //6
	UpdateNameState               //7
	MessageForAllState            //8
)
```

### Основные функции💬
   ![Без имени-1](https://github.com/user-attachments/assets/3885811a-55cb-4210-b04e-caccd06d42c4)


1. **Отправка "ЙОУ!"**:
   - Пользователь нажимает на кнопку , после чего бот берет список всех его друзей и отправляет всем им Йоу.
   ![Йоу!](https://github.com/user-attachments/assets/2b4d52ca-c65d-4b70-aded-5af0e662f3a9)
   - Так же пользователь может отправить Йоу! прикрепив фотографию , для этого нужно отправить фотографию находясь в главном меню.
   ![Йоу! с фото](https://github.com/user-attachments/assets/d44a385d-1ce9-43bb-9090-e3a5c2114897)



2. **Добавление друзей**:
   - Пользователь может добавлять друзей, указывая их **tag**.
   ![Добавление в друзья](https://github.com/user-attachments/assets/246a6220-51e9-4b6e-96c5-7c189de343bd)
   - И не может добавлять уже добавленных.
   ![Уже в друзьях](https://github.com/user-attachments/assets/2ab19e5f-9e52-4b36-b917-00fb19c509e2)



3. **Удаление друзей**:
   - Пользователь может удалять друзей из своего списка, указывая их **tag**.
   ![Удаление друга](https://github.com/user-attachments/assets/8a9ed899-fcd3-44c9-bf1a-2cd3fd8040aa)
   - И, точно так же как и с добавлением друзей, не может удалить несуществующего друга.
   ![нет друга](https://github.com/user-attachments/assets/02d30d1b-9cac-44d7-aee8-e4444359b551)



4. **Изменение имени**:
   - Пользатель может изменить свое имя из главного меню, нужно просто прислать новое имя.
   ![меняю имя](https://github.com/user-attachments/assets/1a0a54d9-b6f7-4975-8da5-578be2dbdfc9)

     

5. **Отображения друзей**:
   - Пользователь может посмотреть всех своих друзей, от кого и кому будет отправляться Йоу.
   ![список](https://github.com/user-attachments/assets/984813cf-0896-4a7d-bc69-0b05cb2fb6f3)



---

## Структура базы данных📋

### Таблица `Users`
| Поле     | Тип      | Описание                     |
|----------|----------|------------------------------|
| `chatId` | BIGINT  | Уникальный идентификатор чата (primary key). |
| `tag`    | TEXT     | Тег пользователя.            |
| `name`   | TEXT     | Имя пользователя.            |
| `state`  | INTEGER  | Текущее состояние пользователя. |


### Таблица `Friends`
| Поле         | Тип      | Описание                     |
|--------------|----------|------------------------------|
| `Id`         | SERIAL   | Уникальный идентификатор записи (primary key). |
| `chatId`     | BIGINT  | Идентификатор первого пользователя. |
| `ScndChatId` | BIGINT  | Идентификатор второго пользователя. |

---

## Установка и запуск🚀

### Требования
- Установленный **Go** (версия 1.20 или выше).
- Установленные выше перечисленные инструменты.
- Токен Telegram Bot API.

### Шаги для запуска
1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/lingSmr/Yo-tg-bot-.git
   cd yo-telegram-bot
   ```
2. Установите необходимые зависимости:
   ```bash
   go mod download
3. Поднимите docker контейнер
   ```bash
   docker compose up -d
4. Перейдите в src/cmd, настройте **.env** файл и соберите программу
   ```bash
   cd src/cmd
   nano .env
   go build -o ""Yo
   ./Yo
   

---

## Лицензия
Этот проект распространяется под лицензией MIT.
