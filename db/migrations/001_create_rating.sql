-- +goose Up
CREATE TABLE IF NOT EXISTS ratings (
  user_id     UUID       NOT NULL,
  game_id     UUID       NOT NULL,
  rating      SMALLINT   NOT NULL CHECK (rating BETWEEN 1 AND 10),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, game_id)
);

-- +goose Down
DROP TABLE IF EXISTS ratings;
