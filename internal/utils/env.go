package utils

import (
	"analyze_narou/internal/client/narou"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func LoadDotEnv(path string) {
	fmt.Println("Loading .env file")
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func RankingModeFromArgs(args []string) (narou.RankingMode, error) {
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
