-- +goose Up

CREATE TABLE ApiURL (
    ID          INTEGER  PRIMARY KEY,
    ApiName     TEXT     NOT NULL UNIQUE,
    ApiURL      TEXT     NOT NULL,
    Description TEXT,
    IsActive    BOOLEAN  NOT NULL DEFAULT 1,
    CreatedAt   DATETIME DEFAULT CURRENT_TIMESTAMP,
    UpdatedAt   DATETIME
);

-- Seed the eight current API URLs. INSERT OR IGNORE leaves existing rows
-- alone, so admin edits made via direct SQL after first deploy survive
-- re-running this migration on a fresh DB.
--
-- REPLACE the 'REPLACE_FROM_CONFIG' values with the actual URLs from
-- the apiURLs: block of your config.yaml secret before deploying.
INSERT OR IGNORE INTO ApiURL (ApiName, ApiURL, Description) VALUES
    ('steam',       'http://api.steampowered.com/ISteamApps/GetAppList/v0002/?key=STEAMKEY&format=json', 'Random Steam game picker (/pick steam)'),
    ('affirmation', 'https://www.affirmations.dev/', 'Daily affirmation (/daily affirmation)'),
    ('advice',      'https://api.adviceslip.com/advice', 'Daily advice slip (/daily advice)'),
    ('doggo',       'https://api.thedogapi.com/v1/breeds/', 'Random dog breed images (/animals doggo)'),
    ('ninjaKatz',   'https://randomuser.me/api/', 'API-Ninjas cats endpoint (/animals katz)'),
    ('fakePerson',  'https://api.api-ninjas.com/v1/cats?name=', 'Random fake person generator (/generate fake-person)'),
    ('xkcd',        'https://c.xkcd.com/random/comic/', 'Latest XKCD comic (/get xkcd)'),
    ('landsat',     'https://science.nasa.gov/specials/your-name-in-landsat/', 'Landsat headless-Chrome target (/generate landsat)');

-- +goose Down

DROP TABLE ApiURL;
