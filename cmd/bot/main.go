package main

import "analyze_narou/internal/app"

func main() {
	config := app.Config{
		"https://api.syosetu.com/novelapi/api/",
		"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",
	}

	app.Run(config)
}
