package main

import (
	"context"
	"fmt"
	"github.com/carterjs/words/internal/store"
	"github.com/carterjs/words/internal/words"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

var (
	GameDir = envOr("GAMES_DIR", filepath.Join("/tmp", "words-game"))
)

var (
	currentGameFile = filepath.Join(GameDir, "currentGame")
	currentGameID   = func() string {
		b, _ := os.ReadFile(currentGameFile)
		return string(b)
	}()
	gameID string
)

func envOr(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

type GameStore interface {
	SaveGame(ctx context.Context, game *words.Game) error
	GetGameByID(ctx context.Context, id string) (*words.Game, error)
}

func main() {
	gameStore := store.NewFS(GameDir)

	cmd := &cobra.Command{
		Use: "words",
		RunE: func(cmd *cobra.Command, args []string) error {
			game, err := gameStore.GetGameByID(context.Background(), gameID)
			if err != nil {
				return err
			}
			if game == nil {
				return fmt.Errorf("game not found: %s", gameID)
			}

			// print the board
			fmt.Println(game.Board.String())
			fmt.Println("---")

			// game status
			fmt.Printf("Round: %d\n", game.Round)
			fmt.Printf("Pool size: %d\n", game.LettersRemaining())
			fmt.Println("Scores:")
			for _, player := range game.Players {
				fmt.Printf("- %s: %d\n", player.Name, player.Score())
			}
			fmt.Printf("Current rurn: %s\n", game.Players[game.Turn].Name)
			printRack(game.Players[game.Turn])

			return nil
		},
	}
	cmd.AddCommand(createGameCommand(gameStore))
	cmd.AddCommand(playerInfoCommand(gameStore))
	cmd.AddCommand(playWordCommand(gameStore))
	cmd.AddCommand(undoCommand(gameStore))

	cmd.PersistentFlags().StringVarP(&gameID, "game", "g", currentGameID, "game id")

	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createGameCommand(store GameStore) *cobra.Command {
	cmd := &cobra.Command{
		Use: "new",
	}

	var players []string
	cmd.Flags().StringArrayVarP(&players, "player", "p", []string{}, "player names")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		game, err := words.NewGame(words.StandardConfig, players...)
		if err != nil {
			return err
		}

		err = store.SaveGame(context.Background(), game)
		if err != nil {
			return err
		}

		err = os.WriteFile(currentGameFile, []byte(game.ID), 0644)
		if err != nil {
			return err
		}

		fmt.Println(game.Board.String())
		fmt.Println("You've started a new game!")
		fmt.Println("First turn:", game.Players[game.Turn].Name)
		printRack(game.Players[game.Turn])

		return nil
	}

	return cmd
}

func playerInfoCommand(store GameStore) *cobra.Command {
	cmd := &cobra.Command{
		Use: "playerName",
	}

	var playerName string
	cmd.Flags().StringVarP(&playerName, "playerName", "p", "", "playerName name")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		game, err := store.GetGameByID(context.Background(), gameID)
		if err != nil {
			return err
		}
		if game == nil {
			return fmt.Errorf("game not found: %s", gameID)
		}

		if playerName == "" {
			playerName = game.Players[game.Turn].Name
		}

		player := game.GetPlayerByName(playerName)
		if player == nil {
			return fmt.Errorf("playerName not found: %s", player)
		}

		fmt.Printf("Name: %s\n", player.Name)
		fmt.Printf("Score: %d\n", player.Score())
		printRack(*player)

		return nil
	}

	return cmd
}

func printRack(player words.Player) {
	fmt.Print("Rack: ")
	var sb strings.Builder
	for i, letter := range player.Letters {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(string(letter))
	}

	fmt.Println(sb.String())
}

func playWordCommand(store GameStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "play <word>",
		Args: cobra.ExactArgs(1),
	}

	var x, y int
	var directionString string
	var blanks []int
	cmd.Flags().IntVarP(&x, "x", "x", 0, "x coordinate")
	cmd.Flags().IntVarP(&y, "y", "y", 0, "y coordinate")
	cmd.Flags().StringVarP(&directionString, "direction", "d", "horizontal", "word direction")
	cmd.Flags().IntSliceVarP(&blanks, "blanks", "b", []int{}, "blank letter indices")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		letters := strings.ToUpper(args[0])

		var direction words.Direction
		switch strings.ToLower(directionString) {
		case "h", "a", "horizontal", "across":
			direction = words.DirectionHorizontal
		case "v", "d", "vertical", "down":
			direction = words.DirectionVertical
		default:
			return fmt.Errorf("invalid direction: %s", args[2])
		}

		word := words.NewWord(x, y, direction, letters)
		if len(blanks) > 0 {
			for _, i := range blanks {
				word = word.WithBlank(i)
			}
		}

		game, err := store.GetGameByID(context.Background(), gameID)
		if err != nil {
			return err
		}
		if game == nil {
			return fmt.Errorf("game not found: %s", gameID)
		}

		player := game.Players[game.Turn].Name
		result, err := game.PlayWord(word)
		if err != nil {
			return err
		}

		err = store.SaveGame(context.Background(), game)
		if err != nil {
			return err
		}

		fmt.Println(game.Board.String())
		fmt.Println("---")

		fmt.Println(result.PointsExplanation)
		fmt.Printf("%s played %s for %d points!\n", player, string(word.Letters), result.Points)
		fmt.Println("Next up:", game.Players[game.Turn].Name)
		printRack(game.Players[game.Turn])

		return nil
	}

	return cmd
}

func undoCommand(store GameStore) *cobra.Command {
	cmd := &cobra.Command{
		Use: "undo",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		game, err := store.GetGameByID(context.Background(), gameID)
		if err != nil {
			return err
		}

		err = game.Undo()
		if err != nil {
			return err
		}

		err = store.SaveGame(context.Background(), game)
		if err != nil {
			return nil
		}

		fmt.Println(game.Board.String())

		fmt.Println("Removed the last turn")
		return nil
	}

	return cmd
}
