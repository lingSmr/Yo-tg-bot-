package botServe

import (
	"Yo/src/consts"
	"Yo/src/models"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	tgAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	op = "server:ListenAndServe"
)

var OneMorTimeError = errors.New("One more time")

// 1, 2, 3 from main.menu
func (s *BotServ) updatingToWithCancel(chatId int64, state int, msgToUsr string) error {
	botMsg := tgAPI.NewMessage(chatId, msgToUsr)
	botMsg.ReplyMarkup = CancelKeyboard
	s.Bot.Send(botMsg)

	err := s.DataBase.UpdateState(chatId, state, s.Ctx)
	if err != nil {
		botMsg := tgAPI.NewMessage(chatId, consts.SendErrorConst)
		s.Bot.Send(botMsg)
		s.logger.Error("Cant update state of user", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	return nil
}

func (s *BotServ) startSwitch(chatId int64) error {
	botMsg := tgAPI.NewMessage(chatId, "Здравстуй дорогой пользователь!\nКак тебя зовут❔")
	s.Bot.Send(botMsg)
	err := s.DataBase.UpdateState(chatId, consts.AskNameState, s.Ctx)
	if err != nil {
		botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
		s.Bot.Send(botMsg)
		s.logger.Error("Cant update state of user", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	return nil
}

func (s *BotServ) askNameSwtich(chatId int64, msg string) error {
	if msg == "" {
		botMsg := tgAPI.NewMessage(chatId, "Ты прислал что то не то , попробуй еще раз")
		s.Bot.Send(botMsg)
		return OneMorTimeError
	}
	err := s.DataBase.UpdateName(chatId, msg, s.Ctx)
	if err != nil {
		botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
		s.Bot.Send(botMsg)
		s.logger.Error("Cant update name of user", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	botMsg := tgAPI.NewMessage(chatId, "Прекрасное имя "+msg+"❕\n"+
		"P.S. Ты всегда сможешь поменять его из главного меню\n")
	s.Bot.Send(botMsg)
	botMsg = tgAPI.NewMessage(chatId, `Теперь давай добавим парочку твоих друзей👐
	Что бы их добавить они должны пройти регистрацию в этом боте до стадии добавления друзей , а ты должен прислать мне тэг твоего друга✍️`)
	botMsg.ReplyMarkup = CancelKeyboard
	s.Bot.Send(botMsg)
	s.DataBase.UpdateState(chatId, consts.AddFriendState, s.Ctx)
	return nil
}

func (s *BotServ) addFriendSwitch(chatId int64, msg string) error {
	if msg == "Отмена" {
		s.returningToMainMenu(s.Bot, s.DataBase, chatId, s.Ctx)
		return nil
	}
	friendTag := strings.ReplaceAll(msg, "@", "")
	if friendTag == "" {
		botMsg := tgAPI.NewMessage(chatId, "ты прислал что то не то , попробуй еще раз.")
		s.Bot.Send(botMsg)
		return OneMorTimeError
	}
	err := s.DataBase.AddFriend(chatId, friendTag, s.Ctx)
	if err != nil {
		s.logger.Error("Cant add friend to user", "Operation", op, "ChatId", chatId, "Error", err)
		botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
		s.Bot.Send(botMsg)
		return err
	}
	botMsg := tgAPI.NewMessage(chatId, "Успех! Друг успешно добавлен🎉")
	s.Bot.Send(botMsg)
	s.returningToMainMenu(s.Bot, s.DataBase, chatId, s.Ctx)
	return nil
}

func (s *BotServ) delFriendSwitch(chatId int64, msg string) error {
	if msg == "Отмена" {
		s.returningToMainMenu(s.Bot, s.DataBase, chatId, s.Ctx)
		return nil
	}
	friendTag := strings.ReplaceAll(msg, "@", "")
	err := s.DataBase.DelFriend(chatId, friendTag, s.Ctx)
	if err != nil {
		botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
		s.Bot.Send(botMsg)
		s.logger.Error("Cant del friend to user", "Operation", op, "ChatId", chatId, "Error", err)
		return nil
	}
	botMsg := tgAPI.NewMessage(chatId, "Друг успешно удален😞")
	s.Bot.Send(botMsg)
	s.returningToMainMenu(s.Bot, s.DataBase, chatId, s.Ctx)
	return nil
}

func (s *BotServ) updateNameSwitch(chatId int64, msg string) error {
	if msg == "Отмена" {
		s.returningToMainMenu(s.Bot, s.DataBase, chatId, s.Ctx)
		return nil
	}
	if msg == "" {
		botMsg := tgAPI.NewMessage(chatId, "Ты прислал что то не то , попробуй еще раз")
		s.Bot.Send(botMsg)
		return OneMorTimeError
	}
	err := s.DataBase.UpdateName(chatId, msg, s.Ctx)
	if err != nil {
		s.logger.Error("Cant update name of user", "Operation", op, "ChatId", chatId, "Error", err)
		botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
		s.Bot.Send(botMsg)
		return err
	}
	botMsg := tgAPI.NewMessage(int64(chatId), "Теперь тебя зовут так: "+msg+" 🤨")
	s.Bot.Send(botMsg)
	s.returningToMainMenu(s.Bot, s.DataBase, chatId, s.Ctx)
	return nil
}

func (s *BotServ) msgForAllSwitch(chatId int64, msg string) error {
	if msg == "Отмена" {
		s.returningToMainMenu(s.Bot, s.DataBase, chatId, s.Ctx)
		return nil
	}
	users, err := s.DataBase.GetAllUsers(s.Ctx)
	if err != nil {
		s.logger.Error("Cant send message to all", "Operation", op, "ChatId", chatId, "Error", err)
		botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
		s.Bot.Send(botMsg)
		return nil
	}
	delete(users, chatId)
	go func() {
		currentChatId := chatId
		msgIn := msg
		for user := range users {
			botMsg := tgAPI.NewMessage(int64(user), msgIn)
			s.Bot.Send(botMsg)
		}
		botMsg := tgAPI.NewMessage(currentChatId, "Сообщение успешно отправленно!")
		s.Bot.Send(botMsg)
	}()
	s.returningToMainMenu(s.Bot, s.DataBase, chatId, s.Ctx)
	return nil
}

func (s *BotServ) sendDocument(chatId int64, url string) error {
	currentChatId := chatId
	file, err := os.Open(url)
	if err != nil {
		s.logger.Error("Cant open .log file to send it", "ChatId", chatId)
		botMsg := tgAPI.NewMessage(currentChatId, ".log не был отправлен")
		botMsg.ReplyMarkup = consts.MainMenuConst
		s.Bot.Send(botMsg)
		s.returningToMainMenu(s.Bot, s.DataBase, chatId, s.Ctx)
		return err
	}

	docMsg := tgAPI.NewDocument(chatId, tgAPI.FileReader{
		Name:   "logs.log",
		Reader: file,
	})
	_, err = s.Bot.Send(docMsg)
	if err != nil {
		s.logger.Error("Cant send logs to user", "ChatId", chatId)
		botMsg := tgAPI.NewMessage(currentChatId, ".log не был отправлен")
		botMsg.ReplyMarkup = consts.MainMenuConst
		s.Bot.Send(botMsg)
		s.returningToMainMenu(s.Bot, s.DataBase, chatId, s.Ctx)
		return err
	}
	s.logger.Info("Send log to user", "ChatId", chatId)
	s.returningToMainMenu(s.Bot, s.DataBase, chatId, s.Ctx)
	return nil
}

func (s *BotServ) yoForAll(chatId int64) error {
	friendMap, err := s.DataBase.GetFriends(chatId, s.Ctx)
	if err != nil || friendMap == nil {
		botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
		s.Bot.Send(botMsg)
		s.logger.Error("Cant take users friends", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}

	go func() {
		currentChatId := chatId
		user, err := s.DataBase.GetData(chatId, s.Ctx)
		if err != nil {
			botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
			s.Bot.Send(botMsg)
			s.logger.Error("Cant take data about user", "Operation", op, "ChatId", chatId, "Error", err)
			s.returningToMainMenu(s.Bot, s.DataBase, currentChatId, s.Ctx)
		}
		wg := sync.WaitGroup{}
		for friendChatId := range friendMap {
			wg.Add(1)
			go func() {
				currentFrId := friendChatId
				s.sendYo(*s.Bot, currentFrId, user.Name, user.Tag)
				wg.Done()
			}()
		}
		wg.Wait()
		botMsg := tgAPI.NewMessage(currentChatId, "Йоу!")
		s.Bot.Send(botMsg)
		s.returningToMainMenu(s.Bot, s.DataBase, currentChatId, s.Ctx)
		s.logger.Info("User Send YO!", "ChatId", currentChatId, "User tag", user.Tag, "User msg", user.Name)
	}()
	return nil
}

func (s *BotServ) allFriends(chatId int64) error {
	strBuilder := strings.Builder{}
	strBuilder.Write([]byte("Список твоих друзей 📋:\n"))
	friendMap, err := s.DataBase.GetFriends(chatId, s.Ctx)
	if err != nil || friendMap == nil {
		botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
		s.Bot.Send(botMsg)
		s.logger.Error("Cant take friends of user", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	go func() {
		currentChatId := chatId
		for friendChatId := range friendMap {
			user, err := s.DataBase.GetData(friendChatId, s.Ctx)
			if err != nil {
				botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
				s.Bot.Send(botMsg)
				s.logger.Error("Cant take data about friend", "Operation", op, "ChatId", chatId, "Error", err)
				continue
			}
			str := fmt.Sprintf("- %s (@%s)\n", user.Name, user.Tag)
			strBuilder.Write([]byte(str))
		}
		botMsg := tgAPI.NewMessage(currentChatId, strBuilder.String())
		s.Bot.Send(botMsg)

		s.returningToMainMenu(s.Bot, s.DataBase, currentChatId, s.Ctx)
	}()
	return nil
}

// отправляет йоу выбранному пользователю. ЧатАйди - кому , имя и тэг - от кого.
func (s *BotServ) sendYo(bot tgAPI.BotAPI, chatId int64, Name, tag string) {
	msg := fmt.Sprintf("%s(@%s) - 🤙🤙🤙", Name, tag)
	botMsg := tgAPI.NewMessage(int64(chatId), msg)
	bot.Send(botMsg)
}

// Возвращает пользователя в главное меню
func (s *BotServ) returningToMainMenu(bot *tgAPI.BotAPI, dataBase models.DataBase, chatId int64, ctx context.Context) error {
	err := dataBase.UpdateState(chatId, consts.NothingState, ctx)
	if err != nil {
		return err
	}
	botMsg := tgAPI.NewMessage(int64(chatId), consts.MainMenuConst)
	botMsg.ReplyMarkup = NothingStateKeyboard
	bot.Send(botMsg)
	return nil
}
