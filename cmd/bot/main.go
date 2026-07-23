package main

import (
	"analyze_narou/configs"
	"analyze_narou/internal/app"
	"analyze_narou/internal/client/narou"
	"analyze_narou/internal/utils"
	"fmt"
	"os"
	"strings"
)

func main() {
	utils.LoadDotEnv(".env")

	config := configs.Load()
	mode, err := rankingModeFromArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	app.Run(config, mode)
}

func rankingModeFromArgs(args []string) (narou.RankingMode, error) {
	if len(args) == 0 {
		return narou.RankingModeDaily, nil
	}
	if len(args) > 1 {
		return "", fmt.Errorf("too many arguments; use exactly one of daily, weekly, quarterly, or yearly")
	}

	switch strings.ToLower(args[0]) {
	case "daily", "d":
		return narou.RankingModeDaily, nil
	case "weekly", "w":
		return narou.RankingModeWeekly, nil
	case "quarterly", "quarter", "q":
		return narou.RankingModeQuarterly, nil
	case "yearly", "annual", "y":
		return narou.RankingModeYearly, nil
	default:
		return "", fmt.Errorf("invalid ranking mode %q; use daily, weekly, quarterly, or yearly", args[0])
	}
}
