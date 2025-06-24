package main

import (
	"context"

	"github.com/sbilibin2017/cs2/internal/apps"
	"github.com/sbilibin2017/cs2/internal/flags"
)

func main() {
	err := run()
	if err != nil {
		panic(err)
	}

}

func run() error {
	config := flags.ParseAppFlags()

	app, err := apps.NewApp(config)
	if err != nil {
		return err
	}

	err = app.Run(context.Background())
	if err != nil {
		return err
	}

	return nil
}
