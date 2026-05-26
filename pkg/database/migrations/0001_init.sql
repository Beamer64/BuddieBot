-- +goose Up

CREATE TABLE Guild (
    ID                          INTEGER  PRIMARY KEY,
    Discord_GuildID             TEXT     NOT NULL UNIQUE,
    AudioEnabled                BOOLEAN  NOT NULL DEFAULT 0,
    PrefixOverride              TEXT,
    Discord_EventNotifChannelID TEXT,
    JoinedAt                    DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE User (
    ID             INTEGER  PRIMARY KEY,
    Discord_UserID TEXT     NOT NULL,
    GuildID        INTEGER  NOT NULL,
    Dosh           INTEGER  NOT NULL DEFAULT 0,
    IsDayOne       BOOLEAN  NOT NULL DEFAULT 0,
    CreatedAt      DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (Discord_UserID, GuildID),
    FOREIGN KEY (GuildID) REFERENCES Guild(ID) ON DELETE CASCADE
);

-- +goose Down

DROP TABLE User;
DROP TABLE Guild;
