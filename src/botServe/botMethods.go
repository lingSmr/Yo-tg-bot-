package botServe

import (
	"Yo/src/consts"
	"Yo/src/models"
	"Yo/src/postgres"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	tgAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
)

func (s *BotServ) updatingToWithCancel(chatId int64, state int, msgToUsr string, ctx context.Context) error {
	const op = "botMethod:updatingToWithCancel"
	botMsg := tgAPI.NewMessage(chatId, msgToUsr)
	botMsg.ReplyMarkup = CancelKeyboard
	s.Bot.Send(botMsg)

	err := s.DataBase.UpdateState(chatId, state, ctx)
	if err != nil {
		botMsg := tgAPI.NewMessage(chatId, consts.SendErrorConst)
		s.Bot.Send(botMsg)
		s.logger.Error("Cant update state of user", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	return nil
}

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

func (s *BotServ) askNameSwtich(chatId int64, msg string, ctx context.Context) error {
	const op = "botMethod:askNameSwitch"
	if msg == "" {
		botMsg := tgAPI.NewMessage(chatId, "Ты прислал что то не то , попробуй еще раз")
		s.Bot.Send(botMsg)
		return nil
	}
	err := s.DataBase.UpdateName(chatId, msg, ctx)
	if err != nil {
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
	err = s.DataBase.UpdateState(chatId, consts.AddFriendState, ctx)
	if err != nil {
		s.logger.Error("Cant update state of user", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	return nil
}

func (s *BotServ) addFriendSwitch(chatId int64, msg string, ctx context.Context) error {
	const op = "botMethod:addFriendSwitch"
	if msg == "Отмена" {
		err := s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
		if err != nil {
			s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
			return err
		}
		return nil
	}
	friendTag := strings.ReplaceAll(msg, "@", "")
	if friendTag == "" {
		botMsg := tgAPI.NewMessage(chatId, "ты прислал что то не то , попробуй еще раз.")
		s.Bot.Send(botMsg)
		return nil
	}
	err := s.DataBase.AddFriend(chatId, friendTag, ctx)
	if errors.Is(err, postgres.FriendshipExistsErr) {
		botMsg := tgAPI.NewMessage(chatId, "Вы уже дружите!")
		s.Bot.Send(botMsg)
		return nil
	} else if errors.Is(err, pgx.ErrNoRows) {
		botMsg := tgAPI.NewMessage(chatId, "Такого пользователя не найдено")
		s.Bot.Send(botMsg)
		return nil
	} else if err != nil {
		s.logger.Error("Cant add friend to user", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}

	botMsg := tgAPI.NewMessage(chatId, "Успех! Друг успешно добавлен🎉")
	s.Bot.Send(botMsg)
	err = s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
	if err != nil {
		s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	return nil
}

func (s *BotServ) delFriendSwitch(chatId int64, msg string, ctx context.Context) error {
	const op = "botMethod:delFriendSwitch"
	if msg == "Отмена" {
		err := s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
		if err != nil {
			s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
			return err
		}
		return nil
	}
	if msg == "" {
		botMsg := tgAPI.NewMessage(chatId, "Ты прислал что то не то , попробуй еще раз")
		s.Bot.Send(botMsg)
		return nil
	}
	friendTag := strings.ReplaceAll(msg, "@", "")
	err := s.DataBase.DelFriend(chatId, friendTag, ctx)
	if errors.Is(err, postgres.FriendshipDontExistsErr) {
		botMsg := tgAPI.NewMessage(chatId, "Вы еще не дружите , что бы сорриться🍅")
		s.Bot.Send(botMsg)
		return nil
	} else if errors.Is(err, pgx.ErrNoRows) {
		botMsg := tgAPI.NewMessage(chatId, "Такого пользователя не найдено")
		s.Bot.Send(botMsg)
		return nil
	} else if err != nil {
		s.logger.Error("Cant del friend to user", "Operation", op, "ChatId", chatId, "Error", err)
		return nil
	}

	botMsg := tgAPI.NewMessage(chatId, "Друг успешно удален😞")
	s.Bot.Send(botMsg)
	err = s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
	if err != nil {
		s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	return nil
}

func (s *BotServ) updateNameSwitch(chatId int64, msg string, ctx context.Context) error {
	const op = "botMethod:updateNameSwtich"
	if msg == "Отмена" {
		err := s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
		if err != nil {
			s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
			return err
		}
		return nil
	}
	if msg == "" {
		botMsg := tgAPI.NewMessage(chatId, "Ты прислал что то не то , попробуй еще раз")
		s.Bot.Send(botMsg)
		return nil
	}
	err := s.DataBase.UpdateName(chatId, msg, ctx)
	if err != nil {
		s.logger.Error("Cant update name of user", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	botMsg := tgAPI.NewMessage(int64(chatId), "Теперь тебя зовут так: "+msg+" 🤨")
	s.Bot.Send(botMsg)
	err = s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
	if err != nil {
		s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	return nil
}

func (s *BotServ) msgForAllSwitch(chatId int64, msg string, ctx context.Context) error {
	const op = "botMethod:msgForAllSwitch"
	if msg == "Отмена" {
		err := s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
		if err != nil {
			s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
			return err
		}
		return nil
	}
	users, err := s.DataBase.GetAllUsers(ctx)
	if err != nil {
		s.logger.Error("Cant send message to all", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	delete(users, chatId)

	msgIn := fmt.Sprintf("❗️АДМИН ВЕЩАЕТ❗️ : %s", msg)
	wg := sync.WaitGroup{}
	for user := range users {
		wg.Add(1)
		go func(frID int64) {
			botMsg := tgAPI.NewMessage(frID, msgIn)
			s.Bot.Send(botMsg)
			wg.Done()
		}(user)
	}
	wg.Wait()
	botMsg := tgAPI.NewMessage(chatId, "Сообщение успешно отправленно!")
	s.Bot.Send(botMsg)
	err = s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
	if err != nil {
		s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	s.logger.Info("User send msg for all", "ChatId", chatId)
	return nil
}

func (s *BotServ) sendDocument(chatId int64, url string, ctx context.Context) error {
	const op = "botMethod:sendDocument"
	file, err := os.Open(url)
	if err != nil {
		s.logger.Error("Cant open .log file to send it", "ChatId", chatId)
		err := s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
		if err != nil {
			s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
			return err
		}
		return err
	}

	docMsg := tgAPI.NewDocument(chatId, tgAPI.FileReader{
		Name:   "logs.txt",
		Reader: file,
	})
	_, err = s.Bot.Send(docMsg)
	if err != nil {
		s.logger.Error("Cant send logs to user", "ChatId", chatId)
		err := s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
		if err != nil {
			s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
			return err
		}
		return err
	}
	s.logger.Info("Send log to user", "Operation", op, "ChatId", chatId)
	err = s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
	if err != nil {
		s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	return nil
}

func (s *BotServ) yoForAll(chatId int64, ctx context.Context) error {
	const op = "botMethod:yoForAll"
	friendMap, err := s.DataBase.GetFriends(chatId, ctx)
	if err != nil || friendMap == nil {
		s.logger.Error("Cant take users friends", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	user, err := s.DataBase.GetData(chatId, ctx)
	if err != nil || user == nil {
		s.logger.Error("Cant take data about user", "Operation", op, "ChatId", chatId, "Error", err)
		err := s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
		if err != nil {
			s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
			return err
		}
		return err
	}
	wg := sync.WaitGroup{}
	for friendChatId := range friendMap {
		wg.Add(1)
		go func(frID int64) {
			err := s.sendYo(frID, user.Name, user.Tag, nil)
			if err != nil {
				return
			}
			wg.Done()
		}(friendChatId)
	}
	wg.Wait()
	botMsg := tgAPI.NewMessage(chatId, "Йоу!")
	s.Bot.Send(botMsg)
	err = s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
	if err != nil {
		s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	s.logger.Info("User Send YO!", "ChatId", chatId, "User tag", user.Tag, "User msg", user.Name)
	return nil
}

func (s *BotServ) allFriends(chatId int64, ctx context.Context) error {
	const op = "botMethod:allFriends"
	strBuilder := strings.Builder{}
	strBuilder.Write([]byte("Список твоих друзей 📋:\n"))
	friendMap, err := s.DataBase.GetFriends(chatId, ctx)
	if err != nil || friendMap == nil {
		s.logger.Error("Cant take friends of user", "Operation", op, "ChatId", chatId, "Error", err)
		err = s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
		if err != nil {
			s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
			return err
		}
		return err
	}

	for friendChatId := range friendMap {
		user, err := s.DataBase.GetData(friendChatId, ctx)
		if err != nil {
			s.logger.Error("Cant take data about one of friend", "Operation", op, "ChatId", chatId, "Error", err)
			continue
		}
		str := fmt.Sprintf("- %s (@%s)\n", user.Name, user.Tag)
		strBuilder.Write([]byte(str))
	}
	botMsg := tgAPI.NewMessage(chatId, strBuilder.String())
	s.Bot.Send(botMsg)

	err = s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
	if err != nil {
		s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}

	return nil
}

func (s *BotServ) sendPhotoToFriends(chatId int64, upd tgAPI.Update, ctx context.Context) error {
	const op = "botMethod:SendPhotoToFriends"
	if upd.Message.Photo == nil {
		botMsg := tgAPI.NewMessage(chatId, "В сообщении не найдено фото\nПопробуйте еще раз")
		s.Bot.Send(botMsg)
		s.logger.Error("No photo in message", "Operation", op, "ChatId", chatId)
		return nil
	}
	user, err := s.DataBase.GetData(chatId, ctx)
	if err != nil {
		s.logger.Error("Cant take data about user", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	fileId := upd.Message.Photo[len(upd.Message.Photo)-1].FileID
	file, err := s.Bot.GetFile(tgAPI.FileConfig{FileID: fileId})
	if err != nil {
		s.logger.Error("Cant take info about photo", "Operation", op, "ChatId", chatId)
		return err
	}
	fileURL := file.Link(s.Token)
	resp, err := http.Get(fileURL)
	if err != nil {
		s.logger.Error("Cant take resp from server", "Operation", op, "ChatId", chatId)
		return err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Cant read photo from resp", "Operation", op, "ChatId", chatId)
		return err
	}
	friends, err := s.DataBase.GetFriends(chatId, ctx)
	if err != nil {
		s.logger.Error("Cant take data about friends", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	wg := sync.WaitGroup{}
	for friend := range friends {
		wg.Add(1)
		go func(ID int64, photoData []byte) {
			err := s.sendYo(ID, user.Name, user.Tag, photoData)
			if err != nil {
				s.logger.Error("Cant send yo to user", "Operation", op, "ChatId", chatId, "Error", err)
				return
			}
			wg.Done()
		}(friend, data)
	}
	wg.Wait()
	botMsg := tgAPI.NewMessage(chatId, "Йоу!\nP.S. С фото")
	s.Bot.Send(botMsg)
	err = s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
	if err != nil {
		s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
		return err
	}
	s.logger.Info("User Send YO with photo!", "ChatId", chatId, "User tag", user.Tag, "User msg", user.Name)
	return nil

}

func (s *BotServ) sendYo(chatId int64, Name, tag string, photo []byte) error {
	if photo != nil {
		msg := fmt.Sprintf("%s(@%s) - 🤙🤙🤙", Name, tag)
		botMsg := tgAPI.NewPhoto(chatId, tgAPI.FileBytes{Name: "photo-popa.png", Bytes: photo})
		botMsg.Caption = msg
		s.Bot.Send(botMsg)
		return nil
	}
	msg := fmt.Sprintf("%s(@%s) - 🤙🤙🤙", Name, tag)
	botMsg := tgAPI.NewMessage(chatId, msg)
	s.Bot.Send(botMsg)
	return nil
}

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
