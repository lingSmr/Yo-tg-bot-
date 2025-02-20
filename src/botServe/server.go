package botServe

import (
	"Yo/src/config"
	"Yo/src/consts"
	"Yo/src/models"
	"context"
	"errors"
	"log/slog"
	"time"

	tgAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
)

type BotServ struct {
	logger   *slog.Logger
	Token    string
	DataBase models.DataBase
	Bot      *tgAPI.BotAPI
	UpdChan  tgAPI.UpdatesChannel
	Ctx      *context.Context
}

func NewBotServ(Token string, DataBase models.DataBase, logger *slog.Logger) (*BotServ, error) {
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

		if v, err := s.DataBase.GetState(ID, ctx); v == 0 || errors.Is(err, pgx.ErrNoRows) {
			name := update.Message.From.UserName
			err := s.DataBase.NewUser(ID, name, update.Message.From.UserName, ctx)
			if err != nil {
				s.sendErr(ID)
				continue
			}
			slog.Info("New User!", "ChatId", ID, "Username", update.Message.From.UserName)
		}

		state, err := s.DataBase.GetState(ID, ctx)
		if err != nil {
			botMsg := tgAPI.NewMessage(ID, "Произошла ошибка!\nПоробуйте еще раз")
			s.Bot.Send(botMsg)
			slog.Error("Cant take state of user", "Operation", op, "ChatId", ID, "Error", err)
			continue
		}

		go func(chatId int64, msgIn string, upd tgAPI.Update) {
			ctxForSwitch, cancel := context.WithTimeout(context.Background(), time.Second*40)
			defer cancel()
			switch state {
			case consts.NothingState:
				switch msg {
				case "🤙Йоу🤙":
					err := s.yoForAll(chatId, ctxForSwitch)
					if err != nil {
						s.sendErr(chatId)
					}
				case "":
					err := s.sendPhotoToFriends(chatId, upd, ctxForSwitch)
					if err != nil {
						s.sendErr(chatId)
					}
				case "1":
					err := s.updatingToWithCancel(chatId, consts.AddFriendState, "Пришли мне тэг друга!✍️", ctxForSwitch)
					if err != nil {
						s.sendErr(chatId)
					}
				case "2":
					err := s.updatingToWithCancel(chatId, consts.DelFriendState, "Пришли мне тэг друга , что уже тебе не друг...", ctxForSwitch)
					if err != nil {
						s.sendErr(chatId)
					}
				case "3":
					err := s.updatingToWithCancel(chatId, consts.UpdateNameState, "Пришли мне новое имя✍️", ctxForSwitch)
					if err != nil {
						s.sendErr(chatId)
					}
				case "4":
					err := s.allFriends(chatId, ctxForSwitch)
					if err != nil {
						s.sendErr(chatId)
					}
				case consts.MessageToAllPhraze:
					err := s.updatingToWithCancel(chatId, consts.MessageForAllState,
						`Пришли мне то , что ты хочешь отправить всем пользователям.`, ctxForSwitch)
					if err != nil {
						s.sendErr(chatId)
					}
				case consts.TakeAllInfoFromBotPraze:
					err := s.sendDocument(chatId, config.GetLogUrl(), ctxForSwitch)
					if err != nil {
						s.sendErr(chatId)
					}
				default:
					botMsg := tgAPI.NewMessage(chatId, "Нет такой команды")
					s.Bot.Send(botMsg)
					err := s.returningToMainMenu(s.Bot, s.DataBase, chatId, ctxForSwitch)
					if err != nil {
						s.sendErr(chatId)
						s.logger.Error("Cant return user to main menu", "Operation", op, "ChatId", chatId, "Error", err)
						return
					}
				}
			case consts.StartState:
				err := s.startSwitch(chatId, ctxForSwitch)
				if err != nil {
					s.sendErr(chatId)
				}
			case consts.AskNameState:
				err := s.askNameSwtich(chatId, msgIn, ctxForSwitch)
				if err != nil {
					s.sendErr(chatId)
				}
			case consts.AddFriendState:
				err := s.addFriendSwitch(chatId, msgIn, ctxForSwitch)
				if err != nil {
					s.sendErr(chatId)
				}
			case consts.DelFriendState:
				err := s.delFriendSwitch(chatId, msgIn, ctxForSwitch)
				if err != nil {
					s.sendErr(chatId)
				}
			case consts.UpdateNameState:
				err := s.updateNameSwitch(chatId, msgIn, ctxForSwitch)
				if err != nil {
					s.sendErr(chatId)
				}
			case consts.MessageForAllState:
				err := s.msgForAllSwitch(chatId, msgIn, ctxForSwitch)
				if err != nil {
					s.sendErr(chatId)
				}
			}
		}(ID, msg, update)
	}
	return nil
}

func (s *BotServ) sendErr(chatId int64) {
	botMsg := tgAPI.NewMessage(chatId, "Произошла ошибка!\nПоробуйте еще раз")
	s.Bot.Send(botMsg)
}
