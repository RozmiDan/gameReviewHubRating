package usecase

import (
	"context"
	"errors"
	"math"

	"github.com/RozmiDan/gameReviewHubRating/internal/entity"
	"go.uber.org/zap"
)

type RatingRepository interface {
	SubmitRatingRepo(ctx context.Context, userID string, gameID string, rating int32) error
	GetGameRatingRepo(ctx context.Context, gameID string) (entity.GameRating, error)
	GetTopGamesRepo(ctx context.Context, limit, offset int32) ([]entity.GameRating, error)
}

type ratingService struct {
	repo   RatingRepository
	logger *zap.Logger
	//producer kafka.Producer

}

func NewRatingService(repository RatingRepository, logger *zap.Logger) *ratingService {
	logger = logger.With(zap.String("layer", "ratingService"))
	return &ratingService{repo: repository, logger: logger}
}

func (s *ratingService) SubmitRating(ctx context.Context, userID string, gameID string, rating int32) error {
	logger := s.logger.With(zap.String("func", "SubmitRating"))

	if err := s.repo.SubmitRatingRepo(ctx, userID, gameID, rating); err != nil {
		logger.Error("some error", zap.Error(err))
		return err
	}

	logger.Info("rating successfuly updated")

	return nil
}

func (s *ratingService) GetGameRating(ctx context.Context, gameID string) (entity.GameRating, error) {
	logger := s.logger.With(zap.String("func", "GetGameRating"))

	game, err := s.repo.GetGameRatingRepo(ctx, gameID)

	game.AverageRating = math.Round(game.AverageRating*100) / 100

	if err != nil {
		if errors.Is(err, entity.ErrGameNotFound) {
			logger.Info("game not found", zap.Error(err))
			return entity.GameRating{}, err
		}
		logger.Info("some error", zap.Error(err))
		return entity.GameRating{}, err
	}

	logger.Info("game successfuly found")

	return game, nil
}

func (s *ratingService) GetTopGames(ctx context.Context, limit, offset int32) ([]entity.GameRating, error) {
	logger := s.logger.With(zap.String("func", "GetTopGames"))

	list, err := s.repo.GetTopGamesRepo(ctx, limit, offset)

	if err != nil {
		logger.Error("some error", zap.Error(err))
		return []entity.GameRating{}, err
	}

	logger.Info("games successfuly found")

	return list, nil
}
