package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/d7561985/pb-ab/cmd/postgres"
	"github.com/d7561985/pb-ab/cmd/postgres/migrate"
	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli/v2" // imports as package "cli"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, os.Interrupt)

		<-ch

		log.Println("stop application")
		cancel()
	}()

	pgCli := postgres.New()

	app := &cli.App{
		Name:     pgCli.Name,
		Usage:    pgCli.Usage,
		Action:   pgCli.Action,
		Flags:    pgCli.Flags,
		Commands: []*cli.Command{migrate.New().Cli()},
	}

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
