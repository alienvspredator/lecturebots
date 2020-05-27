package tgbot

import (
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

// BotApp is the telegram bot application
type BotApp struct {
	pool   *pgxpool.Pool
	redis  *redis.Client
	logger *zap.Logger
	api    *tgbotapi.BotAPI
}

// NewApp creates new BotApp instance
func NewApp(
	pool *pgxpool.Pool,
	logger *zap.Logger,
	client *redis.Client,
	botAPI *tgbotapi.BotAPI,
) (*BotApp, error) {
	if pool == nil {
		return nil, xerrors.New("expected pgxpool.Pool pointer, got nil")
	}
	if logger == nil {
		return nil, xerrors.New("expected zap.Logger pointer, got nil")
	}
	if client == nil {
		return nil, xerrors.New("expected redis.Client pointer, got nil")
	}

	return &BotApp{pool, client, logger, botAPI}, nil
}
