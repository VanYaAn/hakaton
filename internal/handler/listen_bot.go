package handler

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

// runBot — основная функция, запускающая бота с обработкой обновлений
func RunBot(token string, ctx context.Context) error {
	api, err := maxbot.New(token)
	if err != nil {
		return fmt.Errorf("не удалось зарегистрировать бота: %w", err)
	}
	// Проверим, что бот авторизован
	// info, err := api.Bots.GetBot(ctx)
	// if err != nil {
	// 	return fmt.Errorf("не удалось получить информацию о боте: %w", err)
	// }

	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // на случай, если runBot завершится раньше

	// Горутина для обработки SIGINT / SIGTERM (Ctrl+C)
	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
		<-exit
		fmt.Println("\n🛑 Получен сигнал завершения. Останавливаем бота...")
		cancel()
	}()
	for upd := range api.GetUpdates(ctx) {
		switch upd := upd.(type) {
		case *schemes.MessageCreatedUpdate:
			switch upd.GetCommand() {
			case "/start":
				_, err = api.Messages.Send(ctx, maxbot.NewMessage().SetChat(upd.Message.Recipient.ChatId).SetText("Привет, Дорогой друг! Я бот-долбаеб. Используй команду /help чтобы ахуеть"))
				if err != nil {
					fmt.Printf("❌ Не отправилось: %v\n", err)
				}
			case "/help":
				_, err = api.Messages.Send(ctx,
					maxbot.NewMessage().SetChat(
						upd.Message.Recipient.ChatId).SetText(
						"Список команд:\n /create_meeting - создать встречу\n"))
				if err != nil {
					fmt.Printf("❌ Не отправилось: %v\n", err)
				}
			case "/create_meeting":
				_, err = api.Messages.Send(ctx, maxbot.NewMessage().SetChat(upd.Message.Recipient.ChatId).SetText("Введите название встречи или слово Отмена"))
				if err != nil {
					fmt.Printf("❌ Не отправилось: %v\n", err)
				}
			Out:
				for upd := range api.GetUpdates(ctx) {
					switch upd := upd.(type) {
					case *schemes.MessageCreatedUpdate:
						switch upd.Message.Body.Text {
						case "Отмена":
							break Out
						default:
							_, err = api.Messages.Send(ctx, maxbot.NewMessage().SetChat(upd.Message.Recipient.ChatId).SetText("Введите время встречи в формате 25:72"))
							if err != nil {
								fmt.Printf("❌ Не отправилось: %v\n", err)
							}
						Out2:
							for upd := range api.GetUpdates(ctx) {
								switch upd := upd.(type) {
								case *schemes.MessageCreatedUpdate:
									switch upd.Message.Body.Text {
									default:
										_, err = api.Messages.Send(ctx, maxbot.NewMessage().SetChat(upd.Message.Recipient.ChatId).SetText("Встреча сохранена идите нахуй"))
										if err != nil {
											fmt.Printf("❌ Не отправилось: %v\n", err)
										}
										break Out2
									}
								}
							}
							break Out
						}
					}
				}
			default:
				fmt.Printf("📦 Получено обновление типа %T — пропускаем\n", upd)
			}
		default:
			fmt.Printf("📦 Получено обновление типа %T — пропускаем\n", upd)
		}
	}

	fmt.Println("✅ Бот остановлен.")
	return nil
}
