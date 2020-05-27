package notifybot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

// App ...
type App struct {
	logger       *zap.Logger
	api          *tgbotapi.BotAPI
	amqpCh       *amqp.Channel
	subscriberID int64
}

// NewApp ...
func NewApp(logger *zap.Logger, api *tgbotapi.BotAPI, ch *amqp.Channel, subscriberID int64) *App {
	return &App{
		logger,
		api,
		ch,
		subscriberID,
	}
}
