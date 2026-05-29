-- +goose Up

CREATE TABLE UserRating (
    ID         INTEGER  PRIMARY KEY,
    UserID     INTEGER  NOT NULL,
    RatingName TEXT     NOT NULL,
    Value      INTEGER  NOT NULL,
    UpdatedAt  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (UserID, RatingName),
    FOREIGN KEY (UserID) REFERENCES User(ID) ON DELETE CASCADE
);

-- ON DELETE CASCADE on UserID means /user forget-me (which deletes User rows)
-- and Guild departures (which cascade User on Guild deletion) both clean up
-- ratings automatically — no extra application logic required.

-- +goose Down

DROP TABLE UserRating;
