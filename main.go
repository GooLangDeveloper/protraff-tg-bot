package main

import (
	"context"
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
	botToken = os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN не установлен")
	}

	adminID := os.Getenv("ADMIN_CHAT_ID")
	if adminID == "" {
		log.Fatal("ADMIN_CHAT_ID не установлен")
	}
	if _, err := os.Sscanf(adminID, "%d", &adminChatID); err != nil {
		log.Fatalf("Некорректный ADMIN_CHAT_ID: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Бот запущен: @%s", bot.Self.UserName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go reminderWorker(ctx, bot)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-stop:
			log.Println("Получен сигнал остановки, завершаю работу...")
			cancel()
			bot.StopReceivingUpdates()
			return
		case update := <-updates:
			if update.Message != nil {
				handleMessage(bot, update.Message)
			} else if update.CallbackQuery != nil {
				handleCallback(bot, update.CallbackQuery)
			}
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	if message.Contact != nil {
		handleContact(bot, message)
		return
	}

	switch message.Command() {
	case "start":
		sendStart(bot, chatID)
	case "about":
		sendAbout(bot, chatID)
	case "faq":
		sendFAQMenu(bot, chatID)
	case "contact":
		requestContact(bot, chatID)
	case "help":
		sendHelp(bot, chatID)
	default:
		if message.Command() != "" {
			sendStart(bot, chatID)
		}
	}
}

func handleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	if callback.Message == nil {
		log.Println("Получен callback без Message, пропускаю")
		return
	}

	chatID := callback.Message.Chat.ID
	data := callback.Data

	if _, err := bot.Send(tgbotapi.NewCallback(callback.ID, "")); err != nil {
		log.Printf("Ошибка отправки callback ответа: %v", err)
	}

	switch data {
	case "request_contact":
		requestContact(bot, chatID)
	case "contact_manager":
		requestContact(bot, chatID)
	case "faq":
		sendFAQMenu(bot, chatID)
	case "about":
		sendAbout(bot, chatID)
	case "faq_1":
		sendFAQ1(bot, chatID)
	case "faq_2":
		sendFAQ2(bot, chatID)
	case "faq_3":
		sendFAQ3(bot, chatID)
	case "faq_4":
		sendFAQ4(bot, chatID)
	case "faq_5":
		sendFAQ5(bot, chatID)
	case "back_to_start":
		sendStart(bot, chatID)
	}
}

func sendStart(bot *tgbotapi.BotAPI, chatID int64) {
	text := `👋🏻 Добро пожаловать в Pro-traffic.

Чтобы оставить заявку на продвижение,
нажмите кнопку ниже — мы напишем вам в Telegram.

Если нужно задать вопрос или связаться с менеджером,
используйте соответствующий раздел.

Без звонков и навязывания.`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚀 Оставить заявку", "request_contact"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📞 Связаться с менеджером", "contact_manager"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ FAQ", "faq"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("ℹ️ О компании", "about"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки /start: %v", err)
	}
}

func sendAbout(bot *tgbotapi.BotAPI, chatID int64) {
	text := `Pro-traffic — это новая модель продвижения бизнеса в цифровом маркетинге.

Наша миссия — сделать маркетинг доступным
и экономически оправданным для малого и среднего бизнеса.

Мы сознательно убрали всё,
что в классических агентствах раздувает стоимость услуг:
посредников, project- и community-менеджеров,
отделы продаж и лишние уровни согласований.

Почему?
Потому что бизнес платит не за результат,
а за содержание большой внутренней структуры агентства.

В итоге:
— цена растёт
— специалистов в чатах становится больше
— а реальная работа всё равно делается несколькими людьми

Мы выбрали другой путь.

В проекте участвуют только те,
кто напрямую влияет на результат:
вы, ваш бизнес, ИИ и специалисты,
которые реально работают над продвижением.

Наша цель — доказать эффективность этой модели,
собрать своих клиентов
и выстроить долгосрочное сотрудничество,
а не масштабировать штат ради масштаба.`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "back_to_start"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки About: %v", err)
	}
}

func sendFAQMenu(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Выберите вопрос:"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("1. Как устроена работа", "faq_1"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("2. Почему нет менеджеров", "faq_2"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("3. Почему ИИ, а не дизайнер", "faq_3"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("4. Подойдёт ли формат", "faq_4"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("5. Что после заявки", "faq_5"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "back_to_start"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки FAQ меню: %v", err)
	}
}

func sendFAQ1(bot *tgbotapi.BotAPI, chatID int64) {
	text := `Мы работаем по компактной и эффективной модели.

В проекте участвуют:
— ИИ для анализа ниши, конкурентов и офферов
— таргетолог как технический специалист
— маркетолог, отвечающий за стратегию и воронку
— ИИ-инструменты для создания креативов и тестирования гипотез

Без лишних ролей и посредников.`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад к FAQ", "faq"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки FAQ 1: %v", err)
	}
}

func sendFAQ2(bot *tgbotapi.BotAPI, chatID int64) {
	text := `Такие роли оправданы при масштабировании крупных команд.

Для малого и среднего бизнеса
они часто увеличивают стоимость,
не влияя напрямую на результат.

Возникает логичный вопрос: зачем?

Можно подключить ещё десяток людей.
Но ради какой цели?

На практике штат часто раздувается,
чтобы основатель агентства полностью делегировал работу команде.
Стоимость этого делегирования ложится на бизнес.

Мы выстроили процесс иначе —
с прямой и понятной коммуникацией
между бизнесом и специалистами,
которые реально работают над проектом.`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад к FAQ", "faq"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки FAQ 2: %v", err)
	}
}

func sendFAQ3(bot *tgbotapi.BotAPI, chatID int64) {
	text := `ИИ — это рациональный инструмент.

Хороший дизайнер — это специалист
с высокой стоимостью на рынке.
Спрос на дизайнеров, видеографов и мобилографов
растёт во всех нишах.

Вопрос простой:
переплачивать 100–150$ за то,
что креатив сделал человек,
или направить эти деньги в рекламный бюджет?

ИИ позволяет быстрее создавать креативы,
тестировать больше гипотез
и не закладывать стоимость штата в цену услуги.`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад к FAQ", "faq"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки FAQ 3: %v", err)
	}
}

func sendFAQ4(bot *tgbotapi.BotAPI, chatID int64) {
	text := `Формат подойдёт,
если у вас малый или средний бизнес
и нужен понятный запуск рекламы
без перегруженных процессов.`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад к FAQ", "faq"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки FAQ 4: %v", err)
	}
}

func sendFAQ5(bot *tgbotapi.BotAPI, chatID int64) {
	text := `После того как вы оставите контакт,
менеджер свяжется с вами в Telegram.

Мы:
— уточним задачу
— зададим несколько вопросов
— предложим дальнейшие шаги

Без звонков и навязывания.`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад к FAQ", "faq"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки FAQ 5: %v", err)
	}
}

func requestContact(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Нажмите кнопку ниже, чтобы поделиться контактом:"

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📱 Поделиться контактом"),
		),
	)
	keyboard.OneTimeKeyboard = true

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки запроса контакта: %v", err)
		return
	}

	mu.Lock()
	if _, exists := pendingReminders[chatID]; !exists {
		pendingReminders[chatID] = time.Now()
	}
	mu.Unlock()
}

func handleContact(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	mu.Lock()
	delete(pendingReminders, chatID)
	mu.Unlock()

	forward := tgbotapi.NewForward(adminChatID, chatID, message.MessageID)
	if _, err := bot.Send(forward); err != nil {
		log.Printf("Ошибка пересылки контакта админу: %v", err)
	}

	confirmText := "✅ Спасибо! Ваша заявка отправлена. Менеджер свяжется с вами в ближайшее время."
	msg := tgbotapi.NewMessage(chatID, confirmText)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки подтверждения: %v", err)
	}
}

func sendHelp(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Используйте /start для начала работы с ботом."
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки help: %v", err)
	}
}

func reminderWorker(ctx context.Context, bot *tgbotapi.BotAPI) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Остановка reminderWorker")
			return
		case <-ticker.C:
			mu.Lock()
			now := time.Now()
			for chatID, timestamp := range pendingReminders {
				if now.Sub(timestamp) >= 24*time.Hour {
					sendReminder(bot, chatID)
					delete(pendingReminders, chatID)
				}
			}
			mu.Unlock()
		}
	}
}

func sendReminder(bot *tgbotapi.BotAPI, chatID int64) {
	text := `Напоминаем, что вы можете задать вопрос по рекламе.

Если решите оставить заявку —
для вас действует разовая скидка 10%.

Промокод: protraff-2026
Просто укажите его менеджеру при общении.`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚀 Оставить заявку", "request_contact"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки напоминания chatID=%d: %v", chatID, err)
	}
}
