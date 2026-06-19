package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	traqwsbot "github.com/traPtitech/traq-ws-bot"

	"slack-traq-transfer/config"
	"slack-traq-transfer/handler"
)

func main() {
	godotenv.Load()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config.json: %v", err)
	}
	if cfg == nil {
		log.Fatalf("config.json is required but not found")
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	TRAQ_ACCESS_TOKEN := os.Getenv("TRAQ_BOT_ACCESS_TOKEN")
	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		AccessToken: TRAQ_ACCESS_TOKEN,
	})
	if err != nil {
		log.Fatal(err)
	}

	SLACK_ACCESS_TOKEN := os.Getenv("SLACK_TOKEN")
	SLACK_APP_TOKEN := os.Getenv("SLACK_WEBSOCKET_TOKEN")
	api := slack.New(
		SLACK_ACCESS_TOKEN,
		slack.OptionAppLevelToken(SLACK_APP_TOKEN),
		slack.OptionDebug(false),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)
	socket := socketmode.New(
		api,
		socketmode.OptionDebug(false),
		socketmode.OptionLog(log.New(os.Stdout, "socket-mode: ", log.Lshortfile|log.LstdFlags)),
	)
	authTestResp, authTestErr := api.AuthTest()
	if authTestErr != nil {
		log.Fatalf("SLACK_BOT_TOKEN is invalid: %v\n", authTestErr)
	}

	FORM_TOKEN := os.Getenv("FORM_TOKEN")

	appHandler := handler.NewHandler(api, bot, cfg, FORM_TOKEN)
	appHandler.SlackUserID = authTestResp.UserID
	appHandler.SlackBotID = authTestResp.BotID

	go appHandler.RunSlackSocketLoop(socket)
	appHandler.SetupTraqHandlers()

	go func() {
		if err := bot.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		socket.Run()
	}()

	e := echo.New()

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	for _, profile := range cfg.Profiles {
		e.POST("/"+profile.Endpoint, appHandler.HandleFormSubmit(profile.SlackChannelID))
	}

	e.Logger.Fatal(e.Start(":8080"))
}
