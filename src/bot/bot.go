package bot

import (
    "log"
    "go_binance_bot/src/handlers"
    
    "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
    API *tgbotapi.BotAPI
}

func NewBot(token string) (*Bot, error) {
    api, err := tgbotapi.NewBotAPI(token)
    if err != nil {
        return nil, err
    }

    log.Printf("Authorized on account %s", api.Self.UserName)
    return &Bot{API: api}, nil
}

func (b *Bot) Start() {
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates := b.API.GetUpdatesChan(u)

    for update := range updates {
        handlers.HandleMessage(update, b.API)
    }
}
