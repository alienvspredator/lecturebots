package tgbot

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

// Run starts bot loop
func (b *BotApp) Run(ctx context.Context) error {
	b.logger.Info("Starting main loop")

	triggers := make([]interface{}, 0)
	rows, err := b.pool.Query(ctx, `SELECT "word" FROM "triggers" WHERE bot = $1;`, "reader")
	if err != nil {
		return xerrors.Errorf("cannot read triggers from SQL database: %v", err)
	}

	for rows.Next() {
		var trigger string
		rows.Scan(&trigger)
		triggers = append(triggers, strings.ToLower(trigger))
	}

	_, err = b.redis.SAdd(ctx, "reader_triggers", triggers...).Result()
	if err != nil {
		b.logger.Info("Failed to SAdd triggers", zap.Error(err))
	}

	updates, err := b.api.GetUpdatesChan(tgbotapi.NewUpdate(0))
	go func() {
		<-ctx.Done()
		b.api.StopReceivingUpdates()
	}()
	for update := range updates {
		if update.Message == nil {
			continue
		}

		iter := b.redis.SScan(ctx, "reader_triggers", 0, "", 0).Iterator()
		for iter.Next(ctx) {
			if err := iter.Err(); err != nil {
				b.logger.Error("Cannot scan value", zap.Error(err))
			}

			if strings.Contains(strings.ToLower(update.Message.Text), iter.Val()) {
				b.api.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					strings.ToUpper(update.Message.Text)),
				)

				break
			}
		}
	}

	return nil
}
