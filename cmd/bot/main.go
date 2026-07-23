package main

import (
	"analyze_narou/configs"
	"analyze_narou/internal/app"
	"analyze_narou/internal/client/narou"
	"analyze_narou/internal/utils"
)

func main() {
	utils.LoadDotEnv(".env")

	config := configs.CONFIG
	mode := narou.RankingModeDaily

	app.Run(config, mode)
}
