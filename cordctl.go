package main

import (
	bot "cordctl/bot"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		fmt.Println("DISCORD_BOT_TOKEN not found in environment, please set it in .env file")
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: cordctl <commands-directory>")
		os.Exit(1)
	}
	dir := os.Args[1]
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Failed to read directory: %v\n", err)
		os.Exit(1)
	}
	Commands := map[string]bot.Command{}
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".yaml" {
			continue
		}
		cmdName, cmd, err := bot.ParseYAMLCommand(filepath.Join(dir, file.Name()))
		if err != nil {
			fmt.Printf("Failed to parse %s: %v\n", file.Name(), err)
			continue
		}
		Commands[cmdName] = cmd
	}
	b := bot.Bot{
		Token:    token,
		Commands: Commands,
	}
	b.Run()
}
