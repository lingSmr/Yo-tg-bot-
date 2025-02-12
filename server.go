package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	tgAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Server struct {
	Token    string
	DataBase *DataBase
}

func NewServer(Token string, DataBase *DataBase) (*Server, error) {
	return &Server{Token: Token, DataBase: DataBase}, nil
}

func (s *Server) ListAndServe() {
	const op = "server:ListAndServe"

	bot, err := tgAPI.NewBotAPI(s.Token)
	if err != nil {
		log.Fatalf("%s : %s", op, err)
	}

	u := tgAPI.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Printf("%s : %s", op, "Bot Started!")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatId := update.Message.Chat.ID
		msg := update.Message.Text

		if v, _ := s.DataBase.GetState(int(chatId)); v == 0 {
			name := update.Message.From.UserName
			s.DataBase.NewUser(int(chatId), name, update.Message.From.UserName)
			log.Printf("New User! ChatId : %v . Username : %s", chatId, update.Message.From.UserName)
		}
		if err != nil {
			botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
			bot.Send(botMsg)
			log.Print(err)
			continue
		}

		state, err := s.DataBase.GetState(int(chatId))
		if err != nil {
			botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
			bot.Send(botMsg)
			log.Print(err)
			continue
		}
		switch state {
		case NothingState:
			switch msg {
			case "🤙Йоу🤙":
				friendMap, err := s.DataBase.GetFriends(int(chatId))
				if err != nil || friendMap == nil {
					botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
					bot.Send(botMsg)
					log.Print(err)
					continue
				}

				go func() {
					currentChatId := chatId
					user, err := s.DataBase.GetData(int(chatId))
					if err != nil {
						botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
						bot.Send(botMsg)
						log.Print(err)
						returningToMainMenu(bot, s.DataBase, int(currentChatId))
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
					returningToMainMenu(bot, s.DataBase, int(currentChatId))
				}()
			case "1":
				botMsg := tgAPI.NewMessage(chatId, "Пришли мне тэг друга!✍️")
				botMsg.ReplyMarkup = CancelKeyboard
				bot.Send(botMsg)
				s.DataBase.UpdateState(int(chatId), AddFriendState)
				continue
			case "2":
				botMsg := tgAPI.NewMessage(chatId, "Пришли мне тэг друга , что уже тебе не друг...")
				botMsg.ReplyMarkup = CancelKeyboard
				bot.Send(botMsg)
				s.DataBase.UpdateState(int(chatId), DelFriendState)
				continue
			case "3":
				botMsg := tgAPI.NewMessage(chatId, "Пришли мне новое имя✍️")
				botMsg.ReplyMarkup = CancelKeyboard
				bot.Send(botMsg)
				s.DataBase.UpdateState(int(chatId), UpdateNameState)
				continue
			case "4":
				strBuilder := strings.Builder{}
				strBuilder.Write([]byte("Список твоих друзей 📋:\n"))
				friendMap, err := s.DataBase.GetFriends(int(chatId))
				if err != nil || friendMap == nil {
					botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
					bot.Send(botMsg)
					log.Print(err)
					continue
				}
				go func() {
					currentChatId := chatId
					for friendChatId := range friendMap {
						user, err := s.DataBase.GetData(int(friendChatId))
						if err != nil {
							botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
							bot.Send(botMsg)
							log.Print(err)
							continue
						}
						str := fmt.Sprintf("- %s (@%s)\n", user.Name, user.Tag)
						strBuilder.Write([]byte(str))
					}
					botMsg := tgAPI.NewMessage(currentChatId, strBuilder.String())
					bot.Send(botMsg)

					returningToMainMenu(bot, s.DataBase, int(currentChatId))
				}()
			case MessageToAllPhraze:
				botMsg := tgAPI.NewMessage(chatId,
					`Пришли мне то , что ты хочешь отправить всем пользователям.`)
				botMsg.ReplyMarkup = CancelKeyboard
				bot.Send(botMsg)
				s.DataBase.UpdateState(int(chatId), MessageForAllState)
				continue
			}

			// botMsg := tgAPI.NewMessage(chatId, MainMenuConst)
			// botMsg.ReplyMarkup = NothingStateKeyboard
			// bot.Send(botMsg)
		case StartState:
			botMsg := tgAPI.NewMessage(chatId, "Здравстуй дорогой пользователь!\nКак тебя зовут❔")
			bot.Send(botMsg)
			err := s.DataBase.UpdateState(int(chatId), AskNameState)
			if err != nil {
				continue
			}
		case AskNameState:
			if update.Message.Text == "" {
				botMsg := tgAPI.NewMessage(chatId, "Ты прислать что то не то , попробуй еще раз")
				bot.Send(botMsg)
			}
			err := s.DataBase.UpdateName(int(chatId), msg)
			if err != nil {
				botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
				bot.Send(botMsg)
				log.Print(err)
				continue
			}
			botMsg := tgAPI.NewMessage(chatId, "Прекрасное имя "+msg+"❕\n"+
				"P.S. Ты всегда сможешь поменять его из главного меню\n")
			bot.Send(botMsg)
			botMsg = tgAPI.NewMessage(chatId, `Теперь давай добавим парочку твоих друзей👐
			Что бы их добавить они должны пройти регистрацию в этом боте до стадии добавления друзей , а ты должен прислать мне тэг твоего друга✍️`)
			botMsg.ReplyMarkup = CancelKeyboard
			bot.Send(botMsg)
			s.DataBase.UpdateState(int(chatId), AddFriendState)
		case AddFriendState:
			if msg == "Отмена" {
				returningToMainMenu(bot, s.DataBase, int(chatId))
				continue
			}
			friendTag := strings.ReplaceAll(msg, "@", "")
			if friendTag == "" {
				botMsg := tgAPI.NewMessage(chatId, "ты прислал что то не то , попробуй еще раз.")
				bot.Send(botMsg)
				continue
			}
			ok, err := s.DataBase.AddFriend(int(chatId), friendTag)
			if err != nil || ok == 0 {
				botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
				bot.Send(botMsg)
				continue
			}
			botMsg := tgAPI.NewMessage(chatId, "Успех! Друг успешно добавлен🎉")
			bot.Send(botMsg)
			returningToMainMenu(bot, s.DataBase, int(chatId))
		case DelFriendState:
			if msg == "Отмена" {
				returningToMainMenu(bot, s.DataBase, int(chatId))
				continue
			}
			friendTag := strings.ReplaceAll(msg, "@", "")
			ok, err := s.DataBase.DelFriend(int(chatId), friendTag)
			if err != nil || ok == 0 {
				botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
				bot.Send(botMsg)
				log.Printf("%s:%s", "del friend", err)
				continue
			}
			botMsg := tgAPI.NewMessage(chatId, "Друг успешно удален😞")
			bot.Send(botMsg)
			returningToMainMenu(bot, s.DataBase, int(chatId))
		case UpdateNameState:
			if msg == "Отмена" {
				returningToMainMenu(bot, s.DataBase, int(chatId))
				continue
			}
			err := s.DataBase.UpdateName(int(chatId), msg)
			if err != nil {
				botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
				bot.Send(botMsg)
				continue
			}
			botMsg := tgAPI.NewMessage(int64(chatId), "Теперь тебя зовут так: "+msg+" 🤨")
			bot.Send(botMsg)
			returningToMainMenu(bot, s.DataBase, int(chatId))
		case MessageForAllState:
			if msg == "Отмена" {
				returningToMainMenu(bot, s.DataBase, int(chatId))
				continue
			}
			users, err := s.DataBase.GetAllUsers()
			if err != nil {
				botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
				bot.Send(botMsg)
				continue
			}
			delete(users, int(chatId))
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
			returningToMainMenu(bot, s.DataBase, int(chatId))
		}
	}
}

// отправляет йоу выбранному пользователю. ЧатАйди - кому , имя и тэг - от кого.
func sendYo(bot tgAPI.BotAPI, chatId int, Name, tag string) {
	msg := fmt.Sprintf("%s(@%s) - 🤙🤙🤙", Name, tag)
	botMsg := tgAPI.NewMessage(int64(chatId), msg)
	bot.Send(botMsg)
}

func returningToMainMenu(bot *tgAPI.BotAPI, dataBase *DataBase, chatId int) error {
	err := dataBase.UpdateState(chatId, NothingState)
	if err != nil {
		return err
	}
	botMsg := tgAPI.NewMessage(int64(chatId), MainMenuConst)
	botMsg.ReplyMarkup = NothingStateKeyboard
	bot.Send(botMsg)
	return nil
}
