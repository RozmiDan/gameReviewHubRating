package entity

import "errors"

var (
	ErrGameNotFound = errors.New("game not found")
	ErrInvalidUUID  = errors.New("entered uuid is invalid")
)

type GameRating struct {
	GameId        string
	AverageRating float64
	RatingsCount  int64
}

// RatingMessage — та же структура, что и в main_service
type RatingMessage struct {
	GameID string `json:"game_id"`
	UserID string `json:"user_id"`
	Rating int32  `json:"rating"`
}
