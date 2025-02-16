package consts

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

const (
	MainMenuConst = `Главное меню:
				1. Добавить друга 🫂
				2. Удалить друга 👤
				3. Изменить имя 😶‍🌫️
				4. Список Друзей 📋`
	SendErrorConst = `Произошла ошибка!\nПоробуйте еще раз`
)
