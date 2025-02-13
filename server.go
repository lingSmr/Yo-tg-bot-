package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	tgAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Server struct {
	logger   *slog.Logger
	Token    string
	DataBase *DataBase
}

func NewServer(Token string, DataBase *DataBase, logger *slog.Logger) *Server {
	return &Server{Token: Token, DataBase: DataBase, logger: logger}
}

func (s *Server) ListAndServe(ctx context.Context) error {
	const op = "server:ListAndServe"

	bot, err := tgAPI.NewBotAPI(s.Token)
	if err != nil {
		s.logger.Error("Cant make botApi", "Operation", op, "Error", err)
		return err
	}

	u := tgAPI.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	s.logger.Info("Bot started!!!", "Operation", op)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatId := update.Message.Chat.ID
		msg := update.Message.Text

		if v, _ := s.DataBase.GetState(chatId, ctx); v == 0 {
			name := update.Message.From.UserName
			s.DataBase.NewUser(chatId, name, update.Message.From.UserName, ctx)
			s.logger.Info("New User!", "ChatId", chatId, "Username", update.Message.From.UserName)
		}
		if err != nil {
			botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
			bot.Send(botMsg)
			s.logger.Error("Cant take state of user", "Operation", op, "ChatId", chatId, "Error", err)
			continue
		}

		state, err := s.DataBase.GetState(chatId, ctx)
		if err != nil {
			botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
			bot.Send(botMsg)
			s.logger.Error("Cant take state of user", "Operation", op, "ChatId", chatId, "Error", err)
			continue
		}
		switch state {
		case NothingState:
			switch msg {
			case "🤙Йоу🤙":
				friendMap, err := s.DataBase.GetFriends(chatId, ctx)
				if err != nil || friendMap == nil {
					botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
					bot.Send(botMsg)
					s.logger.Error("Cant take users friends", "Operation", op, "ChatId", chatId, "Error", err)
					continue
				}

				go func() {
					currentChatId := chatId
					user, err := s.DataBase.GetData(chatId, ctx)
					if err != nil {
						botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
						bot.Send(botMsg)
						s.logger.Error("Cant take data about user", "Operation", op, "ChatId", chatId, "Error", err)
						returningToMainMenu(bot, s.DataBase, currentChatId, ctx)
					}
					wg := sync.WaitGroup{}
					for friendChatId := range friendMap {
						wg.Add(1)
						go func() {
							currentFrId := friendChatId
							sendYo(*bot, currentFrId, user.Name, user.Tag)
							wg.Done()
						}()
					}
					wg.Wait()
					botMsg := tgAPI.NewMessage(currentChatId, "Йоу!")
					bot.Send(botMsg)
					returningToMainMenu(bot, s.DataBase, currentChatId, ctx)
					s.logger.Info("User Send YO!", "ChatId", currentChatId, "User tag", user.Tag, "User msg", user.Name)
				}()
			case "1":
				botMsg := tgAPI.NewMessage(chatId, "Пришли мне тэг друга!✍️")
				botMsg.ReplyMarkup = CancelKeyboard
				bot.Send(botMsg)
				s.DataBase.UpdateState(chatId, AddFriendState, ctx)
				continue
			case "2":
				botMsg := tgAPI.NewMessage(chatId, "Пришли мне тэг друга , что уже тебе не друг...")
				botMsg.ReplyMarkup = CancelKeyboard
				bot.Send(botMsg)
				s.DataBase.UpdateState(chatId, DelFriendState, ctx)
				continue
			case "3":
				botMsg := tgAPI.NewMessage(chatId, "Пришли мне новое имя✍️")
				botMsg.ReplyMarkup = CancelKeyboard
				bot.Send(botMsg)
				s.DataBase.UpdateState(chatId, UpdateNameState, ctx)
				continue
			case "4":
				strBuilder := strings.Builder{}
				strBuilder.Write([]byte("Список твоих друзей 📋:\n"))
				friendMap, err := s.DataBase.GetFriends(chatId, ctx)
				if err != nil || friendMap == nil {
					botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
					bot.Send(botMsg)
					s.logger.Error("Cant take friends of user", "Operation", op, "ChatId", chatId, "Error", err)
					continue
				}
				go func() {
					currentChatId := chatId
					for friendChatId := range friendMap {
						user, err := s.DataBase.GetData(friendChatId, ctx)
						if err != nil {
							botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
							bot.Send(botMsg)
							s.logger.Error("Cant take data about friend", "Operation", op, "ChatId", chatId, "Error", err)
							continue
						}
						str := fmt.Sprintf("- %s (@%s)\n", user.Name, user.Tag)
						strBuilder.Write([]byte(str))
					}
					botMsg := tgAPI.NewMessage(currentChatId, strBuilder.String())
					bot.Send(botMsg)

					returningToMainMenu(bot, s.DataBase, currentChatId, ctx)
				}()
			case MessageToAllPhraze:
				botMsg := tgAPI.NewMessage(chatId,
					`Пришли мне то , что ты хочешь отправить всем пользователям.`)
				botMsg.ReplyMarkup = CancelKeyboard
				bot.Send(botMsg)
				s.DataBase.UpdateState(chatId, MessageForAllState, ctx)
				continue
			default:
				returningToMainMenu(bot, s.DataBase, chatId, ctx)
			}
		case StartState:
			botMsg := tgAPI.NewMessage(chatId, "Здравстуй дорогой пользователь!\nКак тебя зовут❔")
			bot.Send(botMsg)
			err := s.DataBase.UpdateState(chatId, AskNameState, ctx)
			if err != nil {
				botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
				bot.Send(botMsg)
				s.logger.Error("Cant update state of user", "Operation", op, "ChatId", chatId, "Error", err)
				continue
			}
		case AskNameState:
			if msg == "" {
				botMsg := tgAPI.NewMessage(chatId, "Ты прислал что то не то , попробуй еще раз")
				bot.Send(botMsg)
				continue
			}
			err := s.DataBase.UpdateName(chatId, msg, ctx)
			if err != nil {
				botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
				bot.Send(botMsg)
				s.logger.Error("Cant update name of user", "Operation", op, "ChatId", chatId, "Error", err)
				continue
			}
			botMsg := tgAPI.NewMessage(chatId, "Прекрасное имя "+msg+"❕\n"+
				"P.S. Ты всегда сможешь поменять его из главного меню\n")
			bot.Send(botMsg)
			botMsg = tgAPI.NewMessage(chatId, `Теперь давай добавим парочку твоих друзей👐
			Что бы их добавить они должны пройти регистрацию в этом боте до стадии добавления друзей , а ты должен прислать мне тэг твоего друга✍️`)
			botMsg.ReplyMarkup = CancelKeyboard
			bot.Send(botMsg)
			s.DataBase.UpdateState(chatId, AddFriendState, ctx)
		case AddFriendState:
			if msg == "Отмена" {
				returningToMainMenu(bot, s.DataBase, chatId, ctx)
				continue
			}
			friendTag := strings.ReplaceAll(msg, "@", "")
			if friendTag == "" {
				botMsg := tgAPI.NewMessage(chatId, "ты прислал что то не то , попробуй еще раз.")
				bot.Send(botMsg)
				continue
			}
			err := s.DataBase.AddFriend(chatId, friendTag, ctx)
			if err != nil {
				s.logger.Error("Cant add friend to user", "Operation", op, "ChatId", chatId, "Error", err)
				botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
				bot.Send(botMsg)
				continue
			}
			botMsg := tgAPI.NewMessage(chatId, "Успех! Друг успешно добавлен🎉")
			bot.Send(botMsg)
			returningToMainMenu(bot, s.DataBase, chatId, ctx)
		case DelFriendState:
			if msg == "Отмена" {
				returningToMainMenu(bot, s.DataBase, chatId, ctx)
				continue
			}
			friendTag := strings.ReplaceAll(msg, "@", "")
			err := s.DataBase.DelFriend(chatId, friendTag, ctx)
			if err != nil {
				botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
				bot.Send(botMsg)
				s.logger.Error("Cant del friend to user", "Operation", op, "ChatId", chatId, "Error", err)
				continue
			}
			botMsg := tgAPI.NewMessage(chatId, "Друг успешно удален😞")
			bot.Send(botMsg)
			returningToMainMenu(bot, s.DataBase, chatId, ctx)
		case UpdateNameState:
			if msg == "Отмена" {
				returningToMainMenu(bot, s.DataBase, chatId, ctx)
				continue
			}
			if msg == "" {
				botMsg := tgAPI.NewMessage(chatId, "Ты прислал что то не то , попробуй еще раз")
				bot.Send(botMsg)
				continue
			}
			err := s.DataBase.UpdateName(chatId, msg, ctx)
			if err != nil {
				s.logger.Error("Cant update name of user", "Operation", op, "ChatId", chatId, "Error", err)
				botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
				bot.Send(botMsg)
				continue
			}
			botMsg := tgAPI.NewMessage(int64(chatId), "Теперь тебя зовут так: "+msg+" 🤨")
			bot.Send(botMsg)
			returningToMainMenu(bot, s.DataBase, chatId, ctx)
		case MessageForAllState:
			if msg == "Отмена" {
				returningToMainMenu(bot, s.DataBase, chatId, ctx)
				continue
			}
			users, err := s.DataBase.GetAllUsers(ctx)
			if err != nil {
				s.logger.Error("Cant send message to all", "Operation", op, "ChatId", chatId, "Error", err)
				botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
				bot.Send(botMsg)
				continue
			}
			delete(users, chatId)
			go func() {
				currentChatId := chatId
				msgIn := msg
				for user := range users {
					botMsg := tgAPI.NewMessage(int64(user), msgIn)
					bot.Send(botMsg)
				}
				botMsg := tgAPI.NewMessage(currentChatId, "Сообщение успешно отправленно!")
				bot.Send(botMsg)
			}()
			returningToMainMenu(bot, s.DataBase, chatId, ctx)
		}
	}
	return nil
}

// отправляет йоу выбранному пользователю. ЧатАйди - кому , имя и тэг - от кого.
func sendYo(bot tgAPI.BotAPI, chatId int64, Name, tag string) {
	msg := fmt.Sprintf("%s(@%s) - 🤙🤙🤙", Name, tag)
	botMsg := tgAPI.NewMessage(int64(chatId), msg)
	bot.Send(botMsg)
}

// Возвращает пользователя в главное меню
func returningToMainMenu(bot *tgAPI.BotAPI, dataBase *DataBase, chatId int64, ctx context.Context) error {
	err := dataBase.UpdateState(chatId, NothingState, ctx)
	if err != nil {
		return err
	}
	botMsg := tgAPI.NewMessage(int64(chatId), MainMenuConst)
	botMsg.ReplyMarkup = NothingStateKeyboard
	bot.Send(botMsg)
	return nil
}
