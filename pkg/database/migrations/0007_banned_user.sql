-- +goose Up

CREATE TABLE BannedUser (
    ID             INTEGER  PRIMARY KEY,
    Discord_UserID TEXT     NOT NULL UNIQUE,
    Reason         TEXT,
    BannedAt       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    BannedBy       TEXT
);

-- Bans intentionally have NO foreign key to User. Two reasons:
--   1. /user forget-me deletes User rows but bans should outlive that — the
--      ban is the maintainer's enforcement, not the user's data.
--   2. A user can be pre-emptively banned before they've ever used the bot
--      (no User row exists yet to FK to).
-- Discord_UserID is UNIQUE so a single row covers a user globally; bans are
-- not scoped per guild.

-- +goose Down

DROP TABLE BannedUser;
