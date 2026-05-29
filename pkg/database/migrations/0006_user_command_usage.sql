-- +goose Up

CREATE TABLE UserCommandUsage (
    ID          INTEGER  PRIMARY KEY,
    UserID      INTEGER  NOT NULL,
    CommandName TEXT     NOT NULL,
    UsageCount  INTEGER  NOT NULL DEFAULT 0,
    LastUsedAt  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (UserID, CommandName),
    FOREIGN KEY (UserID) REFERENCES User(ID) ON DELETE CASCADE
);

-- ON DELETE CASCADE on UserID means /user forget-me (which deletes User rows)
-- and Guild departures (which cascade User on Guild deletion) clean up the
-- per-user usage rows automatically.

-- +goose Down

DROP TABLE UserCommandUsage;
