# trafikverket-bot
Trafikverket Bot

![trafikverket-logo](https://bransch.trafikverket.se/contentassets/8a2d09a88ebb431b9892aa9ff7e80b0f/skyltar_film_ny_logo-_1920x1080.jpg)

## About The Project

A simple Telegram bot that collects information from the Trafikverket Open API about deviations on the Swedish railroad network.

Why?
Why not.

### Prerequisites

You need an APIkey from Trafikverket Open API which can be obtained by creating an account at: https://data.trafikverket.se/oauth2/Account/register

## Getting Started

To run this program you need at least Go version 1.22. You also need to create a json-config file in config/config.json that should look like this:

```
{
    "Telegram": {
		"tgAPIkey": "xxx",
		"tgChannel": "xxx"
	},
	"TrafikverketAPIKey": "xxx"
}
```

## Usage

To keep track on updates it's recommended to use for example crontab to run the program once every fifth minute.

```
go run trafikverket.go
```


### DEBUG mode
```
  -config-file string
        Absolute path for config-file (default "./config/config.json")
  -debug
        Turns on debug for telegram
  -stdout
        Turns on stdout rather than sending to telegram
  -telegram-test
        Sends a test message to configured telegram channel
```

## Roadmap

- [x] Add support for other file location for config-file
- [ ] Fix update function posting updates to ongoing events
- [ ] Fix deletion when ongoing event are cleared
- [ ] Improve station lookup function
- [ ] Change to html when sent to Telegram for support of embedded images
- [ ] Fix affected stations print
