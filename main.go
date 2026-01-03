package main

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	BOT_TOKEN     = "8549089105:AAGFrBrus-N4a4cLU1QeRRnvyUjhV3Up21U"
	ADMIN_CHAT_ID = 433873179 // <-- chat ID админа
)

type Reminder struct {
	CreatedAt time.Time
}

var pendingReminders = make(map[int64]Reminder)

func main() {
	bot, err := tgbotapi.NewBotAPI(BOT_TOKEN)
	if err != nil {
		log.Fatal(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	go reminderWorker(bot)

	for update := range updates {

		if update.Message != nil {
			handleMessage(bot, update.Message)
		}

		if update.CallbackQuery != nil {
			handleCallback(bot, update.CallbackQuery)
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {

	chatID := msg.Chat.ID

	// Контакт отправлен
	if msg.Contact != nil {
		delete(pendingReminders, chatID)

		forward := tgbotapi.NewForward(ADMIN_CHAT_ID, chatID, msg.MessageID)
		bot.Send(forward)

		confirm := tgbotapi.NewMessage(chatID, "Спасибо. Мы свяжемся с вами в ближайшее время.")
		bot.Send(confirm)
		return
	}

	switch msg.Text {

	case "/start":
		sendStart(bot, chatID)

	case "/help":
		sendHelp(bot, chatID)

	case "/faq":
		sendFAQ(bot, chatID)

	case "/about":
		sendAbout(bot, chatID)

	case "/contact":
		sendContact(bot, chatID)
	}
}

func handleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery) {

	chatID := cb.Message.Chat.ID

	switch cb.Data {

	case "start":
		sendStart(bot, chatID)

	case "help":
		sendHelp(bot, chatID)

	case "faq":
		sendFAQ(bot, chatID)

	case "about":
		sendAbout(bot, chatID)

	case "contact":
		sendContact(bot, chatID)

	case "leave_contact":
		sendContactRequest(bot, chatID)
		pendingReminders[chatID] = Reminder{CreatedAt: time.Now()}
	}

	bot.Request(tgbotapi.NewCallback(cb.ID, ""))
}

func sendStart(bot *tgbotapi.BotAPI, chatID int64) {

	text := `Вы в Pro-traffic.

Маркетинг нового поколения на базе ИИ
для малого и среднего бизнеса.

Выберите, с чего хотите начать.`

	msg := tgbotapi.NewMessage(chatID, text)

	msg.ReplyMarkup = mainMenu()

	bot.Send(msg)
}

func sendHelp(bot *tgbotapi.BotAPI, chatID int64) {

	text := `Pro-traffic — маркетинг нового поколения на базе ИИ.

Мы работаем с малым и средним бизнесом,
используя ИИ для оптимизации рекламы
и ускорения запуска кампаний.`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = mainMenu()

	bot.Send(msg)
}

func sendFAQ(bot *tgbotapi.BotAPI, chatID int64) {

	text := `Частые вопросы:

— С какими бизнесами вы работаете?
С малым и средним бизнесом.

— Какие каналы используете?
Facebook / Instagram, Яндекс Директ, Telegram Ads.

— Подойдёт ли небольшой бюджет?
Да, под такие задачи и выстроен формат.`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = mainMenu()

	bot.Send(msg)
}

func sendAbout(bot *tgbotapi.BotAPI, chatID int64) {

	text := `Pro-traffic — команда маркетинга нового поколения на базе ИИ.

Наша цель — упростить запуск рекламы
и сделать маркетинг доступным для бизнеса.`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = mainMenu()

	bot.Send(msg)
}

func sendContact(bot *tgbotapi.BotAPI, chatID int64) {

	text := `Оставьте контакт — мы напишем вам первыми.
Без звонков и без спама.`

	msg := tgbotapi.NewMessage(chatID, text)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📲 Оставить контакт", "leave_contact"),
		),
	)

	msg.ReplyMarkup = keyboard

	bot.Send(msg)
}

func sendContactRequest(bot *tgbotapi.BotAPI, chatID int64) {

	msg := tgbotapi.NewMessage(chatID, "Нажмите кнопку ниже, чтобы отправить контакт.")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📞 Отправить контакт"),
		),
	)

	keyboard.OneTimeKeyboard = true
	keyboard.ResizeKeyboard = true

	msg.ReplyMarkup = keyboard

	bot.Send(msg)
}

func mainMenu() tgbotapi.InlineKeyboardMarkup {

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚀 Начать", "start"),
			tgbotapi.NewInlineKeyboardButtonData("ℹ️ Help", "help"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ FAQ", "faq"),
			tgbotapi.NewInlineKeyboardButtonData("📞 Контакты", "contact"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏢 About", "about"),
		),
	)
}

func reminderWorker(bot *tgbotapi.BotAPI) {

	ticker := time.NewTicker(10 * time.Minute)

	for range ticker.C {

		now := time.Now()

		for chatID, r := range pendingReminders {

			if now.Sub(r.CreatedAt) >= 24*time.Hour {

				text := `Если вопрос по рекламе всё ещё актуален,
вы можете оставить контакт — мы напишем вам первыми.`

				msg := tgbotapi.NewMessage(chatID, text)

				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("📲 Оставить контакт", "leave_contact"),
					),
				)

				msg.ReplyMarkup = keyboard
				bot.Send(msg)

				delete(pendingReminders, chatID)
			}
		}
	}
}
