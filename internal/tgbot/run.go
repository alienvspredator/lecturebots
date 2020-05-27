package tgbot

import (
	"context"
	"encoding/json"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/streadway/amqp"
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
		return xerrors.Errorf("failed to SAdd triggers: %v", err)
	}

	q, err := b.amqpCh.QueueDeclare("lection", false, false, false, false, nil)
	if err != nil {
		return xerrors.Errorf("failed to declare AMQP queue: %v", err)
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
				break
			}

			if strings.Contains(strings.ToLower(update.Message.Text), iter.Val()) {
				body, err := json.Marshal(update.Message)
				if err != nil {
					b.logger.Error("Failed to marshall message", zap.Error(err))
					break
				}

				if err := b.amqpCh.Publish("", q.Name, false, false, amqp.Publishing{
					DeliveryMode: amqp.Persistent,
					ContentType:  "application/json",
					Body:         body,
				}); err != nil {
					b.logger.Error("Failed to publish message", zap.Error(err))
					break
				}

				break
			}
		}
	}

	return nil
}
