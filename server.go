package main

import (
	"Yo/configs"
	"context"
	"errors"
	tgAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
)

type BotServ struct {
	logger   *slog.Logger
	Token    string
	DataBase DataBase
	Bot      *tgAPI.BotAPI
	UpdChan  tgAPI.UpdatesChannel
	Ctx      context.Context
}

func NewBotServ(Token string, DataBase DataBase, logger *slog.Logger, ctx context.Context) (*BotServ, error) {
	const op = "Botserv:NewBotServ"

	config := configs.InitConfig()
	defer config.LogFile.Close()
	slog.SetDefault(config.Logger)

	bot, err := tgAPI.NewBotAPI(config.Token)
	if err != nil {
		slog.Error("Cant make botApi", "Operation", op, "Error", err)
		return nil, err
	}

	u := tgAPI.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	slog.Info("Bot inited", "Operation", op)

	return &BotServ{
		Token:    Token,
		DataBase: DataBase,
		logger:   logger,
		Bot:      bot,
		UpdChan:  updates,
		Ctx:      ctx,
	}, nil
}

func (s *BotServ) ListAndServe(ctx context.Context) error {
	const op = "Botserv:ListAndServe"
	slog.SetDefault(s.logger)
	slog.Info("Bot started!!!", "Operation", op)

	for update := range s.UpdChan {
		if update.Message == nil {
			continue
		}

		chatId := update.Message.Chat.ID
		msg := update.Message.Text

		if v, err := s.DataBase.GetState(chatId, ctx); v == 0 || errors.As(err, "no rows") {
			name := update.Message.From.UserName
			s.DataBase.NewUser(chatId, name, update.Message.From.UserName, ctx)
			slog.Info("New User!", "ChatId", chatId, "Username", update.Message.From.UserName)
		}

		state, err := s.DataBase.GetState(chatId, ctx)
		if err != nil {
			botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
			s.Bot.Send(botMsg)
			slog.Error("Cant take state of user", "Operation", op, "ChatId", chatId, "Error", err)
			continue
		}

		switch state {
		case NothingState:
			switch msg {
			case "🤙Йоу🤙":
				go func(chId int64) { s.yoForAll(chatId) }(chatId)
				continue
			case "1":
				s.updatingToWithCancel(chatId, AddFriendState, "Пришли мне тэг друга!✍️")
				continue
			case "2":
				s.updatingToWithCancel(chatId, DelFriendState, "Пришли мне тэг друга , что уже тебе не друг...")
				continue
			case "3":
				s.updatingToWithCancel(chatId, UpdateNameState, "Пришли мне новое имя✍️")
				continue
			case "4":
				go func(chId int64) { s.allFriends(chatId) }(chatId)
				continue
			case MessageToAllPhraze:
				s.updatingToWithCancel(chatId, MessageForAllState, `Пришли мне то , что ты хочешь отправить всем пользователям.`)
				continue
			case TakeAllInfoFromBotPraze:
				go func(chId int64) { s.sendDocument(chatId, configs.GetLogUrl()) }(chatId)
				continue
			default:
				botMsg := tgAPI.NewMessage(chatId, "Нет такой команды")
				s.Bot.Send(botMsg)
				s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
			}
		case StartState:
			go func(chId int64) { s.startSwitch(chatId) }(chatId)
			continue
		case AskNameState:
			go func(chId int64) { s.askNameSwtich(chatId, msg) }(chatId)
			continue
		case AddFriendState:
			go func(chId int64) { s.addFriendSwitch(chatId, msg) }(chatId)
			continue
		case DelFriendState:
			go func(chId int64) { s.delFriendSwitch(chatId, msg) }(chatId)
			continue
		case UpdateNameState:
			go func(chId int64) { s.updateNameSwitch(chatId, msg) }(chatId)
			continue
		case MessageForAllState:
			go func(chId int64) { s.msgForAllSwitch(chatId, msg) }(chatId)
			continue
		}
	}
	return nil
}
