package main

import (
	"github.com/RozmiDan/gameReviewHubRating/internal/app"
	"github.com/RozmiDan/gameReviewHubRating/internal/config"
)

func main() {
	cfg := config.MustLoad()
	app.Run(cfg)
}
