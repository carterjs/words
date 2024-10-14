package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/carterjs/words/internal/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	viper.SetDefault("port", "8080")
	viper.SetDefault("dataDir", "/tmp/word-game")
	viper.SetDefault("publicDir", "public")
	viper.MustBindEnv("port", "PORT")
	viper.MustBindEnv("dataDir", "DATA_DIR")
	viper.MustBindEnv("publicDir", "PUBLIC_DIR")

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	cmd := &cobra.Command{
		Use: "api",
		RunE: func(cmd *cobra.Command, args []string) error {
			port := viper.GetString("port")
			store := store.NewFS(viper.GetString("dataDir"))

			server := NewServer(store, viper.GetString("publicDir"))
			logger.Info("starting server", "port", port, "dataDir", viper.GetString("dataDir"), "publicDir", viper.GetString("publicDir"))
			return http.ListenAndServe(":"+port, server.Handler())
		},
	}

	if err := cmd.Execute(); err != nil {
		logger.Error("error executing command", "error", err)
		os.Exit(1)
	}

}
