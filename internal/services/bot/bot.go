package bot

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

// Bot is a service with bot's logic.
type Bot struct {
	bot *tgbotapi.BotAPI
}

// New returns Bot service.
func New(bot *tgbotapi.BotAPI) *Bot {
	return &Bot{bot: bot}
}

// Start starts Bot service.
func (b *Bot) Start() error {
	log.Printf("Arthorized on account %s", b.bot.Self.UserName)

	updates := b.initUpdatesChannel()
	err := b.handleUpdates(updates)

	if err != nil {
		return err
	}

	return nil
}

// handleUpdates handles all updates of Bot.
func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) error {
	for update := range updates {
		if update.Message != nil { // If we got a message
			fmt.Println("##########################################################################")
			err := b.handleMessage(update.Message)
			if err != nil {
				return errors.Wrap(err, "handling message")
			}
		}
	}
	return nil
}

// initUpdatesChannel starts long polling.
func (b *Bot) initUpdatesChannel() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return b.bot.GetUpdatesChan(u)
}
