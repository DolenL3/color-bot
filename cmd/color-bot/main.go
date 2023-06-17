package main

import (
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"

	"colorbot/internal/services/bot"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(errors.Wrap(err, "error running app"))
		os.Exit(1)
	}
}

func run() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	var botToken string = os.Getenv("TELEGRAM_BOT_KEY")
	if botToken == "" {
		return errors.New("No telegram api key found in .env file")
	}

	tgBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return err
	}
	tgBot.Debug = true

	telegramBot := bot.New(tgBot)
	err = telegramBot.Start()
	if err != nil {
		return err
	}

	return nil
}
