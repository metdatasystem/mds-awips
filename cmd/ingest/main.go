package main

import (
	"context"
	"log"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/metdatasystem/mds-awips/internal/ingest"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ingest",
	Short: "AWIPS ingest using the NWWS-OI",
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

		server, err := ingest.New()
		if err != nil {
			slog.Error("failed to create a new server", "error", err)
			return
		}

		go func() {
			err := server.Run()
			if err != nil {
				log.Fatalf("Fatal: %v", err)
			}
		}()

		<-ctx.Done()
		slog.Info("shutting down")
		server.Shutdown()
		slog.Info("successful shutdown")
	},
}

func init() {
	rootCmd.Flags().StringVar(&env, "env", "", "Specify the path of an env file to load")
}

var env string

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
