package postgres_storage

import (
	"context"
	"errors"

	"github.com/RozmiDan/gameReviewHubRating/internal/entity"
	"github.com/RozmiDan/gameReviewHubRating/pkg/postgres"
	"github.com/jackc/pgx"
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

	tx, err := r.pg.Pool.Begin(ctx)
	if err != nil {
		logger.Error("Begin tx failded", zap.Error(err))
	}
	defer tx.Rollback(ctx)

	var oldRating int32
	isNew := false

	err = tx.QueryRow(ctx,
		`SELECT rating FROM ratings WHERE user_id=$1 AND game_id=$2`,
		userID, gameID,
	).Scan(&oldRating)

	if err != nil {
		if err := pgx.ErrNoRows; err == pgx.ErrNoRows {
			isNew = true
		} else {
			logger.Error("failed to select old rating", zap.Error(err))
			return err
		}
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO ratings(user_id, game_id, rating)
        VALUES($1, $2, $3)
        ON CONFLICT (user_id, game_id)
        DO UPDATE SET rating = EXCLUDED.rating
    `, userID, gameID, rating)
	if err != nil {
		logger.Error("upsert ratings failed", zap.Error(err))
		return err
	}

	if isNew {
		_, err = tx.Exec(ctx, `
            INSERT INTO game_ratings(game_id, ratings_count, ratings_sum)
            VALUES($1, 1, $2)
            ON CONFLICT (game_id) DO UPDATE
              SET
                ratings_count = game_ratings.ratings_count + 1,
                ratings_sum   = game_ratings.ratings_sum + EXCLUDED.ratings_sum
        `, gameID, rating)
	} else {
		delta := rating - oldRating
		_, err = tx.Exec(ctx, `
            UPDATE game_ratings
            SET ratings_sum = ratings_sum + $1
            WHERE game_id = $2
        `, delta, gameID)
	}
	if err != nil {
		logger.Error("upsert game_ratings failed", zap.Error(err))
		return err
	}

	_, err = tx.Exec(ctx, `
        UPDATE game_ratings
        SET average_rating = ROUND(ratings_sum::numeric / ratings_count, 2)
        WHERE game_id = $1
    `, gameID)
	if err != nil {
		logger.Error("recalculate average failed", zap.Error(err))
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error("commit tx failed", zap.Error(err))
		return err
	}

	logger.Info("rating successfully updated",
		zap.String("game_id", gameID),
		zap.String("user_id", userID),
		zap.Int32("new_rating", rating),
		zap.Int32("old_rating", oldRating),
		zap.Bool("new_record", isNew),
	)

	return nil
}

func (r *RatingRepository) GetGameRatingRepo(ctx context.Context, gameID string) (entity.GameRating, error) {
	logger := r.logger.With(zap.String("func", "GetGameRatingRepo"))
	row := r.pg.Pool.QueryRow(ctx, `
      SELECT game_id, average_rating, ratings_count
      FROM game_ratings
      WHERE game_id = $1
    `, gameID)

	var gameRat entity.GameRating
	if err := row.Scan(&gameRat.GameId, &gameRat.AverageRating, &gameRat.RatingsCount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error("gameID not found", zap.Error(err))
			return entity.GameRating{}, entity.ErrGameNotFound
		}
		logger.Error("scan failed", zap.Error(err))
		return entity.GameRating{}, err
	}
	return gameRat, nil
}

func (r *RatingRepository) GetTopGamesRepo(ctx context.Context, limit, offset int32) ([]entity.GameRating, error) {
	return []entity.GameRating{}, nil
}
