package postgres_storage

import (
	"context"
	"errors"

	"github.com/RozmiDan/gameReviewHubRating/internal/entity"
	"github.com/RozmiDan/gameReviewHubRating/pkg/postgres"
	"github.com/jackc/pgx/v5"
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
			logger.Info("gameID not found")
			return entity.GameRating{}, entity.ErrGameNotFound
		}
		logger.Error("scan failed", zap.Error(err))
		return entity.GameRating{}, err
	}

	logger.Info("game successfuly found",
		zap.String("game_id", gameID),
	)

	return gameRat, nil
}

func (r *RatingRepository) GetTopGamesRepo(ctx context.Context, limit, offset int32) ([]entity.GameRating, error) {
	logger := r.logger.With(zap.String("func", "GetTopGamesRepo"))

	rows, err := r.pg.Pool.Query(ctx, `
      SELECT game_id, average_rating, ratings_count
      FROM game_ratings
      ORDER BY average_rating DESC
      LIMIT $1 OFFSET $2
    `, limit, offset*limit)

	if err != nil {
		logger.Error("query failed", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var out []entity.GameRating
	for rows.Next() {
		var gr entity.GameRating
		if err := rows.Scan(&gr.GameId, &gr.AverageRating, &gr.RatingsCount); err != nil {
			logger.Error("scan failed", zap.Error(err))
			return nil, err
		}
		out = append(out, gr)
	}

	logger.Info("games successfuly found", zap.Int("count", len(out)))

	return out, nil
}
