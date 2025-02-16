package botServe

import tgAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"

var NothingStateKeyboard = tgAPI.NewReplyKeyboard(
	tgAPI.NewKeyboardButtonRow(
		tgAPI.NewKeyboardButton("🤙Йоу🤙"),
	),
	tgAPI.NewKeyboardButtonRow(
		tgAPI.NewKeyboardButton("1"),
		tgAPI.NewKeyboardButton("2"),
		tgAPI.NewKeyboardButton("3"),
		tgAPI.NewKeyboardButton("4"),
	),
)

var CancelKeyboard = tgAPI.NewReplyKeyboard(
	tgAPI.NewKeyboardButtonRow(
		tgAPI.NewKeyboardButton("Отмена"),
	),
)
