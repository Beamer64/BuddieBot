-- +goose Up

-- Aggregate, privacy-safe command counts — no user or guild IDs, just how
-- often each top-level command has been invoked bot-wide.
CREATE TABLE CommandUsage (
    ID          INTEGER  PRIMARY KEY,
    CommandName TEXT     NOT NULL UNIQUE,
    UsageCount  INTEGER  NOT NULL DEFAULT 0,
    LastUsedAt  DATETIME
);

-- +goose Down

DROP TABLE CommandUsage;
