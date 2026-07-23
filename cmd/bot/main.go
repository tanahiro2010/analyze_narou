package main

import (
	"analyze_narou/configs"
	"analyze_narou/internal/app"
	"analyze_narou/internal/utils"
	"fmt"
	"os"
)

func main() {
	utils.LoadDotEnv(".env")

	config := configs.Load()
	mode, err := utils.RankingModeFromArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	app.Run(config, mode)
}
