package botServe

import (
	"Yo/src/config"
	"Yo/src/consts"
	"Yo/src/models"
	"context"
	"errors"
	"log/slog"

	tgAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotServ struct {
	logger   *slog.Logger
	Token    string
	DataBase models.DataBase
	Bot      *tgAPI.BotAPI
	UpdChan  tgAPI.UpdatesChannel
	Ctx      context.Context
}

func NewBotServ(Token string, DataBase models.DataBase, logger *slog.Logger, ctx context.Context) (*BotServ, error) {
	const op = "Botserv:NewBotServ"

	config := config.InitConfig()
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

		ID := update.Message.Chat.ID
		msg := update.Message.Text

		if v, err := s.DataBase.GetState(ID, ctx); v == 0 || errors.As(err, "no rows") {
			name := update.Message.From.UserName
			s.DataBase.NewUser(ID, name, update.Message.From.UserName, ctx)
			slog.Info("New User!", "ChatId", ID, "Username", update.Message.From.UserName)
		}

		state, err := s.DataBase.GetState(ID, ctx)
		if err != nil {
			botMsg := tgAPI.NewMessage(ID, "Произошла ошибка!\nПоробуйте еще раз")
			s.Bot.Send(botMsg)
			slog.Error("Cant take state of user", "Operation", op, "ChatId", ID, "Error", err)
			continue
		}

		go func(chatId int64, msgIn string) {
			switch state {
			case consts.NothingState:
				switch msg {
				case "🤙Йоу🤙":
					s.yoForAll(chatId)
				case "1":
					s.updatingToWithCancel(chatId, consts.AddFriendState, "Пришли мне тэг друга!✍️")
				case "2":
					s.updatingToWithCancel(chatId, consts.DelFriendState, "Пришли мне тэг друга , что уже тебе не друг...")
				case "3":
					s.updatingToWithCancel(chatId, consts.UpdateNameState, "Пришли мне новое имя✍️")
				case "4":
					s.allFriends(chatId)
				case consts.MessageToAllPhraze:
					s.updatingToWithCancel(chatId, consts.MessageForAllState, `Пришли мне то , что ты хочешь отправить всем пользователям.`)
				case consts.TakeAllInfoFromBotPraze:
					s.sendDocument(chatId, config.GetLogUrl())
				default:
					botMsg := tgAPI.NewMessage(chatId, "Нет такой команды")
					s.Bot.Send(botMsg)
					s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctx)
				}
			case consts.StartState:
				s.startSwitch(chatId)
			case consts.AskNameState:
				s.askNameSwtich(chatId, msgIn)
			case consts.AddFriendState:
				s.addFriendSwitch(chatId, msgIn)
			case consts.DelFriendState:
				s.delFriendSwitch(chatId, msgIn)
			case consts.UpdateNameState:
				s.updateNameSwitch(chatId, msgIn)
			case consts.MessageForAllState:
				s.msgForAllSwitch(chatId, msgIn)
			}
		}(ID, msg)
	}
	return nil
}
