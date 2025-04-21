-- +goose Up
CREATE TABLE IF NOT EXISTS game_ratings (
  game_id        UUID    PRIMARY KEY,
  ratings_count  BIGINT  NOT NULL DEFAULT 0,
  ratings_sum    BIGINT  NOT NULL DEFAULT 0,
  average_rating NUMERIC(4,2)
);

-- +goose Down
DROP TABLE IF EXISTS game_ratings;
