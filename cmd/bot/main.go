package main

import "analyze_narou/internal/app"

func main() {
	config := app.Config{
		"https://api.syosetu.com/novelapi/api/",
		"",
		""
	}

	app.Run(config)
}