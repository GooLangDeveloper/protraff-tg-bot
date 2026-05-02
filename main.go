package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	botToken    string
	adminChatID int64

	pendingReminders = make(map[int64]time.Time)
	mu               sync.Mutex
)

func main() {
	// === ENV ===
	botToken = os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN не установлен")
	}

	adminID := os.Getenv("ADMIN_CHAT_ID")
	if adminID == "" {
		log.Fatal("ADMIN_CHAT_ID не установлен")
	}

	if _, err := fmt.Sscanf(adminID, "%d", &adminChatID); err != nil {
		log.Fatalf("Некорректный ADMIN_CHAT_ID: %v", err)
	}

	// === BOT INIT ===
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go reminderWorker(ctx, bot)

	log.Println("Bot started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Bot shutting down")
			return

		case update := <-updates:
			if update.Message != nil {
				handleMessage(bot, update.Message)
			}

			if update.CallbackQuery != nil {
				handleCallback(bot, update.CallbackQuery)
			}
		}
	}
}

// ================= HANDLERS =================

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	// Контакт
	if msg.Contact != nil {
		mu.Lock()
		delete(pendingReminders, chatID)
		mu.Unlock()

		forward := tgbotapi.NewForward(adminChatID, chatID, msg.MessageID)
		bot.Send(forward)

		confirm := tgbotapi.NewMessage(chatID, "Спасибо. Менеджер свяжется с вами в Telegram.")
		confirm.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		bot.Send(confirm)
		return
	}

	switch msg.Text {
	case "/start":
		sendStart(bot, chatID)
	case "/about":
		sendAbout(bot, chatID)
	case "/faq":
		sendFAQMenu(bot, chatID)
	case "/contact":
		requestContact(bot, chatID)
	}
}

func handleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery) {
	if cb.Message == nil {
		return
	}

	chatID := cb.Message.Chat.ID

	switch cb.Data {
	case "start":
		sendStart(bot, chatID)

	case "about":
		sendAbout(bot, chatID)

	case "faq":
		sendFAQMenu(bot, chatID)

	case "contact":
		requestContact(bot, chatID)

	case "faq_1":
		sendText(bot, chatID, faq1())

	case "faq_2":
		sendText(bot, chatID, faq2())

	case "faq_3":
		sendText(bot, chatID, faq3())

	case "faq_4":
		sendText(bot, chatID, faq4())

	case "faq_5":
		sendText(bot, chatID, faq5())

	case "back":
		sendStart(bot, chatID)

	case "leave_contact":
		requestContact(bot, chatID)
	}

	bot.Request(tgbotapi.NewCallback(cb.ID, ""))
}

// ================= UI =================

func sendStart(bot *tgbotapi.BotAPI, chatID int64) {
	text := `👋🏻 Добро пожаловать в Pro-traffic.

Чтобы оставить заявку на продвижение,
нажмите кнопку ниже — мы напишем вам в Telegram.

Если нужно задать вопрос или связаться с менеджером,
используйте соответствующий раздел.

Без звонков и навязывания.`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = mainMenu()
	bot.Send(msg)
}

func sendAbout(bot *tgbotapi.BotAPI, chatID int64) {
	text := `Pro-traffic — это новая модель продвижения бизнеса в цифровом маркетинге.

Наша миссия — сделать маркетинг доступным
и экономически оправданным для малого и среднего бизнеса.

Мы убрали всё, что раздувает стоимость услуг:
посредников, лишние роли и уровни согласований.

В проекте участвуют только те,
кто напрямую влияет на результат:
вы, ваш бизнес, ИИ и специалисты,
которые реально работают над продвижением.`

	sendText(bot, chatID, text)
}

func sendFAQMenu(bot *tgbotapi.BotAPI, chatID int64) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Как устроена работа", "faq_1"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Почему нет менеджеров", "faq_2"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Почему ИИ, а не дизайнер", "faq_3"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подойдёт ли формат", "faq_4"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Что после заявки", "faq_5"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "back"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Частые вопросы:")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func requestContact(bot *tgbotapi.BotAPI, chatID int64) {
	mu.Lock()
	pendingReminders[chatID] = time.Now()
	mu.Unlock()

	msg := tgbotapi.NewMessage(chatID, "Нажмите кнопку ниже, чтобы отправить контакт.")
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📞 Отправить контакт"),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = true
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func sendText(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = backMenu()
	bot.Send(msg)
}

// ================= MENUS =================

func mainMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚀 Оставить заявку", "leave_contact"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ FAQ", "faq"),
			tgbotapi.NewInlineKeyboardButtonData("ℹ️ О компании", "about"),
		),
	)
}

func backMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "back"),
		),
	)
}

// ================= FAQ TEXTS =================

func faq1() string {
	return `Мы работаем по компактной и эффективной модели.

В проекте участвуют:
— ИИ для анализа ниши, конкурентов и офферов
— таргетолог как технический специалист
— маркетолог, отвечающий за стратегию и воронку
— ИИ-инструменты для создания и тестирования креативов

Без лишних ролей и посредников.`
}

func faq2() string {
	return `Такие роли оправданы при масштабировании крупных команд.

Для малого и среднего бизнеса
они часто увеличивают стоимость,
не влияя напрямую на результат.

Мы выстроили процесс
с прямой и понятной коммуникацией
между бизнесом и специалистами.`
}

func faq3() string {
	return `ИИ — это рациональный инструмент.

Он позволяет быстрее создавать креативы,
тестировать больше гипотез
и направлять бюджет в рекламу,
а не в содержание штата.`
}

func faq4() string {
	return `Формат подойдёт,
если у вас малый или средний бизнес
и нужен понятный запуск рекламы
без перегруженных процессов.`
}

func faq5() string {
	return `После того как вы оставите контакт,
менеджер свяжется с вами в Telegram.

Мы уточним задачу
и предложим дальнейшие шаги.

Без звонков и навязывания.`
}

// ================= REMINDER =================

func reminderWorker(ctx context.Context, bot *tgbotapi.BotAPI) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			now := time.Now()

			mu.Lock()
			for chatID, ts := range pendingReminders {
				if now.Sub(ts) >= 24*time.Hour {
					text := `Напоминаем, что вы можете задать вопрос по рекламе.

Если решите оставить заявку —
для вас действует разовая скидка 10%.

Промокод: protraff-2026
Просто укажите его менеджеру при общении.`

					msg := tgbotapi.NewMessage(chatID, text)
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("🚀 Оставить заявку", "leave_contact"),
						),
					)

					bot.Send(msg)
					delete(pendingReminders, chatID)
				}
			}
			mu.Unlock()
		}
	}
}
