package tgbot

import "context"

// Run starts bot loop
func (b *BotApp) Run(ctx context.Context) error {
	b.logger.Info("Starting main loop")
	return nil
}
