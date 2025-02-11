# Yo! Telegram Bot🤙

## Описание проекта

Yo! Telegram Bot — это бот, вдохновленный приложением "Bro!" из сериала "Кремниевая долина".  
Основная цель бота — позволить пользователям в любой момент отправить своим друзьям смачное **ЙОУ!** 🚀

---

## Технологии💻

В проекте использованы следующие технологии:
- **Golang** — основной язык разработки.
- **telegram-bot-api** — библиотека для взаимодействия с Telegram Bot API.
- **pgx v4** — драйвер для работы с PostgreSQL.
- **migrate** — инструмент для управления миграциями базы данных.

---

## Как работает бот❓

### Регистрация пользователя
![{16AC04AA-ACBF-4789-A564-5493CD20761A}](https://github.com/user-attachments/assets/a2539f27-1e5f-499c-a77d-725fe322aa25)

1. Когда пользователь впервые отправляет сообщение боту, происходит проверка, зарегистрирован ли он.
2. Если пользователь новый, выполняется `INSERT`-запрос в таблицу `Users`, где сохраняются:
   - `chatId` (primary key) — уникальный идентификатор чата.
   - `tag` — тег пользователя.
   - `state` — текущее состояние пользователя.

### Состояния пользователя⏳
![{BCE5DD94-5250-4220-8EA4-DD9D979883CA}](https://github.com/user-attachments/assets/cadb4662-69c2-4676-9d27-114be027ba0c)

Бот работает на основе состояний (`state`). В зависимости от значения `state`, бот выполняет определенные действия:
- **State 0**: Незарегестрированный Пользователь.
- **State 1**: Стадия регистрации.
- **State 2**: Ожидание команды.
- ...

### Основные функции💬
![{8A5D47E2-96CA-4334-9FE3-7BFD3EB7F7FA}](https://github.com/user-attachments/assets/8c0b8918-b9c4-4196-bc39-676c10e251e7)

1. **Отправка "ЙОУ!"**:
   - ![{B98B3F43-2EF6-4F51-BD27-DBD3654533F3}](https://github.com/user-attachments/assets/8640c76d-3077-44b5-b2a9-d634cb2fb4ae)

   - Пользователь нажимает на кнопку , после чего бот берет список всех его друзей и отправляет всем им Йоу.

2. **Добавление друзей**:
   - ![34ке3](https://github.com/user-attachments/assets/87e70f4d-b698-484b-929d-ee310b34afc1)

   - Пользователь может добавлять друзей, указывая их `tag`.

3. **Удаление друзей**:
   - ![пцукиукццукпу](https://github.com/user-attachments/assets/84ff3717-81b7-48f6-83f4-297fece0a06b)
   - Пользователь может удалять друзей из своего списка, указывая их `tag`.

4. **Изменение имени*:
   - ![{B3F86D55-42F6-4EB4-B508-D9F5E6D287CF}](https://github.com/user-attachments/assets/f102040d-a835-4eff-b19d-e452142b7208)
     
   - Пользатель может изменить свое имя из главного меню, нужно просот прислать новое имя.

5. **Отображения друзей**:
   - ![у234к34к](https://github.com/user-attachments/assets/81278287-8624-48cb-92f8-f2fcde45016f)

   - Пользователь может посмотреть всех своих друзей, от кого и кому будет отправляться Йоу.
---

## Установка и запуск🚀

### Требования
- Установленный **Go** (версия 1.20 или выше).
- База данных **PostgreSQL**.
- Токен Telegram Bot API.

### Шаги для запуска
1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/your-username/yo-telegram-bot.git
   cd yo-telegram-bot
   ```
2. Соберите исполняемый файл заменив токен и адресс бд в main.go:
   ```bash
   go build .
4. Запустите исполняемый файл:
   ```bash
   server.exe

---

## Структура базы данных📋

### Таблица `Users`
| Поле     | Тип      | Описание                     |
|----------|----------|------------------------------|
| `chatId` | BIGINT  | Уникальный идентификатор чата (primary key). |
| `tag`    | TEXT     | Тег пользователя.            |
| `name`   | TEXT     | Имя пользователя.            |
| `state`  | INTEGER  | Текущее состояние пользователя. |

### Пример
![345ка34е](https://github.com/user-attachments/assets/fd15e300-246d-46ff-9ab5-6d20166c79c0)

### Таблица `Friends`
| Поле         | Тип      | Описание                     |
|--------------|----------|------------------------------|
| `Id`         | SERIAL   | Уникальный идентификатор записи (primary key). |
| `chatId`     | BIGINT  | Идентификатор первого пользователя. |
| `ScndChatId` | BIGINT  | Идентификатор второго пользователя. |

### Пример
![{15B6F206-2728-435E-ABF9-8F7C59EE5771}](https://github.com/user-attachments/assets/ebc7be84-3dfa-4d36-b505-5e3507165c00)

---

## Пример Использования👣
1. Пользователь запускает бота
2. Вводит свое имя, которое будет отображаться другим пользователям
3. Отправляет тэг друга. Например: @Example
4. Из главного меню отправляет всем своих друзья Йоу , т.к. друг всего один, Йоу будет отправлен только ему
---

## Лицензия
Этот проект распространяется под лицензией MIT.
