package main

import (
	"log"
	"os"
	"strconv"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/getsentry/sentry-go"

	"swisstronik/app"
	"swisstronik/cmd/swisstronikd/cmd"
)

func main() {
	sentryDsn := os.Getenv("SENTRY_DSN")
	sentryRate := os.Getenv("SENTRY_RATE")

	var rate float64
	if sentryDsn != "" {
		if sentryRate == "" {
			rate = 1.0
		} else {
			parsedRate, err := strconv.ParseFloat(sentryRate, 64)
			if err != nil {
				log.Fatalf("Cannot parse sentry trace rate: %s", err)
			}
			rate = parsedRate
		}

		err := sentry.Init(sentry.ClientOptions{
			Dsn:              sentryDsn,
			TracesSampleRate: rate,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
	}

	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
