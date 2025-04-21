package postgres_storage

import (
	"context"

	"github.com/RozmiDan/gameReviewHubRating/internal/entity"
	"github.com/RozmiDan/gameReviewHubRating/pkg/postgres"
	"go.uber.org/zap"
)

type RatingRepository struct {
	pg     *postgres.Postgres
	logger *zap.Logger
}

func New(pg *postgres.Postgres, logger *zap.Logger) *RatingRepository {
	logger = logger.With(zap.String("layer", "RatingRepository"))
	return &RatingRepository{pg, logger}
}

func (r *RatingRepository) SubmitRatingRepo(ctx context.Context, userID string, gameID string, rating int32) error {
	logger := r.logger.With(zap.String("func", "SubmitRatingRepo"))

	query, args, err := r.pg.Builder.Insert("ratings").
		Columns("user_id", "game_id", "rating").
		Values(userID, gameID, rating).
		ToSql()

	if err != nil {
		logger.Error("ошибка при создании запроса:", zap.Error(err))
		return err
	}

	cmnd, err := r.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		logger.Error("Some error", zap.Error(err))
		return err
	}
	logger.Info(cmnd.String())

	logger.Info("Rating successfuly updated")

	return nil
}

func (r *RatingRepository) GetGameRatingRepo(ctx context.Context, gameID string) (entity.GameRating, error) {
	return entity.GameRating{}, nil
}

func (r *RatingRepository) GetTopGamesRepo(ctx context.Context, limit, offset int32) ([]entity.GameRating, error) {
	return []entity.GameRating{}, nil
}
