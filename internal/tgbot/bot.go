package tgbot

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

// BotApp is the telegram bot application
type BotApp struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewApp creates new BotApp instance
func NewApp(pool *pgxpool.Pool, logger *zap.Logger) (*BotApp, error) {
	if pool == nil {
		return nil, xerrors.New("expected pgxpool.Pool pointer, got nil")
	}
	if logger == nil {
		return nil, xerrors.New("expected zap.Logger pointer, got nil")
	}

	return &BotApp{
		pool,
		logger,
	}, nil
}
