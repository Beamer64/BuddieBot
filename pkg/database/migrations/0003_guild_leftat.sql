-- +goose Up

-- LeftAt is NULL while the bot is in the guild, and a timestamp once the bot
-- is removed. We mark rather than delete so per-guild data (User rows, future
-- balances) survives a kick or a transient outage and is restored on rejoin.
ALTER TABLE Guild ADD COLUMN LeftAt DATETIME;

-- +goose Down

ALTER TABLE Guild DROP COLUMN LeftAt;
