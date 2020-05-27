package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alienvspredator/tgbot/internal/notifybot"
	"github.com/alienvspredator/tgbot/internal/tgbot"
	"github.com/alienvspredator/tgbot/pkg/flagsetup"
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

var (
	flagToken           string
	flagVerb            bool
	flagDSN             string
	flagRedisAddr       string
	flagRabbitMQURL     string
	flagSubscriberToken string
	flagSubscriberID    int64

	requiredFlags = []string{"token", "dsn", "redis-addr", "rabbitmq-url", "subtoken"}

	gitCommit string
)

func init() {
	flag.StringVar(&flagToken, "token", "", "Telegram token")
	flag.BoolVar(&flagVerb, "v", false, "Verbose mode")
	flag.StringVar(
		&flagDSN,
		"dsn",
		"",
		`DSN string. Example:
	-dsn "user=username password=pass host=localhost port=5432 dbname=name sslmode=prefer"

See details:
	https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING`,
	)

	flag.StringVar(&flagRedisAddr, "redis-addr", "", "Redis address")
	flag.StringVar(&flagRabbitMQURL, "rabbitmq-url", "", "RabbitMQ URL")
	flag.StringVar(&flagSubscriberToken, "subtoken", "", "Subscriber bot token")
	flag.Int64Var(&flagSubscriberID, "subscriber-id", 0, "Subscriber ID")
}

func newLogger(debug bool) (*zap.Logger, error) {
	if debug {
		return zap.NewDevelopment(zap.AddCaller())
	}

	return zap.NewProduction()
}

func outputInfo() {
	fmt.Printf("%s\nCommit: %s\n", logo, gitCommit)
}

func main() {
	outputInfo()

	flag.Parse()
	if err := flagsetup.CheckRequired(requiredFlags); err != nil {
		log.Fatalln(err)
	}

	logger, err := newLogger(flagVerb)
	if err != nil {
		log.Fatalf("Cannot create logger: %v\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)

	pool, err := pgxpool.Connect(ctx, flagDSN)
	if err != nil {
		logger.Fatal("Cannot connect to database", zap.Error(err))
	}
	defer pool.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: flagRedisAddr,
	})
	defer redisClient.Close()

	publishBot, err := tgbotapi.NewBotAPI(flagToken)
	if err != nil {
		logger.Fatal("Cannot create bot api instance", zap.Error(err))
	}

	rabbitMQ, err := amqp.Dial(flagRabbitMQURL)
	if err != nil {
		logger.Fatal("Cannot dial RabbitMQ", zap.Error(err))
	}
	defer rabbitMQ.Close()

	amqpCh, err := rabbitMQ.Channel()
	if err != nil {
		logger.Fatal("Cannot open AMQP channel", zap.Error(err))
	}
	defer amqpCh.Close()

	publishBotApp, err := tgbot.NewApp(
		pool,
		logger.Named("TG_BOT"),
		redisClient,
		amqpCh,
		publishBot,
	)
	if err != nil {
		logger.Fatal("Failed to create publisher app instance", zap.Error(err))
	}

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT)
		s := <-sigc
		logger.Info("Got OS signal", zap.Stringer("Signal", s))
		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := publishBotApp.Run(ctx); err != nil {
			logger.Error("Bot exited with error", zap.Error(err))
		}

		logger.Info("Bot goroutine stopped")
	}()

	subscriberBot, err := tgbotapi.NewBotAPI(flagSubscriberToken)
	if err != nil {
		logger.Fatal("Failed to create subscriber bot", zap.Error(err))
	}

	subscrCh, err := rabbitMQ.Channel()
	if err != nil {
		logger.Fatal("Cannot open AMQP channel", zap.Error(err))
	}
	defer subscrCh.Close()
	notifyApp := notifybot.NewApp(logger, subscriberBot, subscrCh, flagSubscriberID)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := notifyApp.Run(ctx); err != nil {
			logger.Error("Notification bot exited with error", zap.Error(err))
		}

		logger.Info("Notification bot goroutine stopped")
	}()

	wg.Wait()
	logger.Info("Application stopped")
}
