package notifybot

import (
	"context"
	"encoding/json"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

// Run ...
func (a *App) Run(ctx context.Context) error {
	q, err := a.amqpCh.QueueDeclare("lection", false, false, false, false, nil)
	if err != nil {
		return xerrors.Errorf("failed to declare AMQP queue: %v", err)
	}
	a.amqpCh.Qos(1, 0, false)

	msgCh, err := a.amqpCh.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return xerrors.Errorf("failed to consume messages", zap.Error(err))
	}

	go func() {
		<-ctx.Done()
		a.amqpCh.Close()
	}()

	for msg := range msgCh {
		msg.Ack(false)
		var tgmsg tgbotapi.Message
		if err := json.Unmarshal(msg.Body, &tgmsg); err != nil {
			a.logger.Error("Failed to unmarshall message", zap.Error(err))
			continue
		}

		a.api.Send(tgbotapi.NewMessage(
			a.subscriberID,
			fmt.Sprintf("Началась лекция %s %s", tgmsg.Chat.FirstName, tgmsg.Chat.LastName),
		))
	}

	return nil
}
