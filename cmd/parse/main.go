package main

import (
	"context"
	"log"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/metdatasystem/mds-awips/internal/parse"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "parse",
	Short: "The AWIPS server listens for pending products and processes them",
	Run: func(cmd *cobra.Command, args []string) {
		// Listen for shutdown signals
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		if env != "" {
			err := godotenv.Load(env)
			if err != nil {
				slog.Error("failed loading env", "error", err)
				return
			}
		}

		config := parse.Config{
			MinLog: minlog,
		}

		server, err := parse.New(config)
		if err != nil {
			slog.Error("failed to create a new server", "error", err)
			return
		}

		go server.Start()

		<-ctx.Done()
		slog.Info("shutting down")
		server.Shutdown()
		slog.Info("successful shutdown")
	},
}

func init() {
	rootCmd.Flags().StringVar(&env, "env", "", "Specify the path of an env file to load")
	rootCmd.Flags().IntVar(&minlog, "minlog", 0, "The minimum logging level to use")
}

var env string
var minlog int

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
