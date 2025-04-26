package entity

import "errors"

var (
	ErrGameNotFound = errors.New("game not found")
	ErrInvalidUUID  = errors.New("entered uuid is invalid")
)
