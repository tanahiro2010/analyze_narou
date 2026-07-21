package main

import "analyze_narou/internal/app"

func main() {
	config := app.Config{
		"https://api.syosetu.com/",
		"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",
		"https://discord.com/api/webhooks/1416959680076447764/2fBoIjV2dXhLxyX1IbLGROkb9excq8FRTT8wwov9Vv1Ui18vIHzfDnorR1qH25mGcI9p",
	}

	app.Run(config)
}
