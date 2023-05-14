package main

import (
	"log"
	"os"
	"time"

	"context"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	app "github.com/piatoss3612/presentation-helper-bot/internal/app/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/config"
	"github.com/piatoss3612/presentation-helper-bot/internal/service/study"
	store "github.com/piatoss3612/presentation-helper-bot/internal/store/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/tools"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger

func main() {
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	sugar = logger.Sugar()

	defer func() {
		if r := recover(); r != nil {
			sugar.Info("Panic recovered", "error", r)
		}
	}()

	mustSetTimezone(os.Getenv("TIME_ZONE"))

	run()
}

func run() {
	cfg := mustLoadConfig(os.Getenv("CONFIG_FILE"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient := mustConnectMongoDB(ctx, cfg.MongoDB.URI)
	defer func() {
		_ = mongoClient.Disconnect(context.Background())
		sugar.Info("Disconnected from MongoDB!")
	}()

	sugar.Info("Connected to MongoDB!")

	cache := mustInitStudyCache(ctx, cfg.Redis.Addr, 1*time.Minute)

	sugar.Info("Study cache is ready!")

	svc := mustInitStudyService(
		ctx, store.NewTx(mongoClient, cfg.MongoDB.DBName),
		cfg.Discord.GuildID, cfg.Discord.ManagerID, cfg.Discord.NoticeChannelID, cfg.Discord.ReflectionChannelID,
	)

	sugar.Info("Study service is ready!")

	sess := mustOpenDiscordSession(cfg.Discord.BotToken)

	bot := app.New(sess, svc, cache, sugar).Setup()

	stop, err := bot.Run()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = bot.Close()
		sugar.Info("Disconnected from Discord!")
	}()

	sugar.Info("Connected to Discord!")

	<-stop
}

func mustSetTimezone(tz string) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		sugar.Fatal(err)
	}

	time.Local = loc
}

func mustLoadConfig(path string) *config.StudyConfig {
	cfg, err := config.NewStudyConfig(path)
	if err != nil {
		sugar.Fatal(err)
	}

	return cfg
}

func mustConnectMongoDB(ctx context.Context, uri string) *mongo.Client {
	mongoClient, err := tools.ConnectMongoDB(ctx, uri)
	if err != nil {
		sugar.Fatal(err)
	}

	return mongoClient
}

func mustInitStudyCache(ctx context.Context, addr string, ttl time.Duration) store.Cache {
	cache, err := tools.ConnectRedisCache(ctx, addr, ttl)
	if err != nil {
		sugar.Fatal(err)
	}

	return store.NewCache(cache)
}

func mustInitStudyService(ctx context.Context, tx store.Tx, guildID, managerID, noticeChID, reflectionChID string) study.Service {
	svc, err := study.NewService(ctx, tx, guildID, managerID, noticeChID, reflectionChID)
	if err != nil {
		sugar.Fatal(err)
	}

	return svc
}

func mustOpenDiscordSession(token string) *discordgo.Session {
	sess, err := discordgo.New("Bot " + token)
	if err != nil {
		sugar.Fatal(err)
	}

	return sess
}
