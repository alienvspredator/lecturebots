package tgbot

import (
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

// BotApp is the telegram bot application
type BotApp struct {
	pool   *pgxpool.Pool
	redis  *redis.Client
	amqpCh *amqp.Channel
	logger *zap.Logger
	api    *tgbotapi.BotAPI
}

// NewApp creates new BotApp instance
func NewApp(
	pool *pgxpool.Pool,
	logger *zap.Logger,
	redisClient *redis.Client,
	amqpCh *amqp.Channel,
	botAPI *tgbotapi.BotAPI,
) (*BotApp, error) {
	if pool == nil {
		return nil, xerrors.New("expected pgxpool.Pool pointer, got nil")
	}
	if logger == nil {
		return nil, xerrors.New("expected zap.Logger pointer, got nil")
	}
	if redisClient == nil {
		return nil, xerrors.New("expected redis.Client pointer, got nil")
	}

	return &BotApp{pool, redisClient, amqpCh, logger, botAPI}, nil
}
