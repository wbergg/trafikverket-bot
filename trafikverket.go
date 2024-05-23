package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/wbergg/bordershop-bot/tele"
	"github.com/wbergg/trafikverket-bot/apipoller"
	"github.com/wbergg/trafikverket-bot/config"
	"github.com/wbergg/trafikverket-bot/db"
)

func main() {
	// Enable bool debug flag
	debugTelegram := flag.Bool("debug", false, "Turns on debug for telegram")
	debugStdout := flag.Bool("stdout", false, "Turns on stdout rather than sending to telegram")
	telegramTest := flag.Bool("telegram-test", false, "Sends a test message to specified telegram channel")
	configFile := flag.String("config-file", "./config/config.json", "Absolute path for config-file")
	flag.Parse()

	// Load config
	config, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Error(err)
		panic("Could not load config, check config/config.json")
	}

	channel, err := strconv.ParseInt(config.Telegram.TgChannel, 10, 64)
	if err != nil {
		log.Error(err)
		panic("Could not convert Telegram channel to int64")
	}

	// Initiate telegram
	tg := tele.New(config.Telegram.TgAPIKey, channel, *debugTelegram, *debugStdout)
	tg.Init(*debugTelegram)

	// Setup db
	d := db.Open()

	// Check if DB is set up, if not, set it up (first time only)
	if d.Setup == 0 {
		fmt.Println("Looks like it's the first time - Populating DB...")
		apipoller.DbSetup(&d)
		fmt.Println("DB population sucess! Please rerun the program!")
		os.Exit(0)
	}

	// Program start
	if *telegramTest {
		tg.SendM("DEBUG: trafikverket-bot test message")
		// End program after sending message
		os.Exit(0)
	} else {
		// Poll and diff data from categories
		data := apipoller.GetData()
		apipoller.UpdateData(tg, &d, data)
	}
}
