-- foreign keys must be disabled to drop tables with foreign keys
PRAGMA foreign_keys = OFF;

DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

-- foreign keys must be enabled with every database connection
PRAGMA foreign_keys = ON;

CREATE TABLE users
(
    -- columns for identification
    user_id         INTEGER PRIMARY KEY AUTOINCREMENT,
    email           TEXT UNIQUE NOT NULL,
    timezone        TEXT        NOT NULL DEFAULT 'UTC',

    -- columns for authentication
    is_active       INTEGER     NOT NULL DEFAULT 0,
    hashed_password TEXT        NOT NULL,

    -- columns for authorization
    clan            TEXT UNIQUE NOT NULL,
    role            TEXT        NOT NULL DEFAULT 'user',

    -- columns for auditing
    created_at      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login      TIMESTAMP   NOT NULL
);

CREATE TABLE roles
(
    user_id INTEGER NOT NULL,
    role    TEXT    NOT NULL,
    PRIMARY KEY (user_id, role),
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);

CREATE TABLE sessions
(
    sess_id    TEXT      NOT NULL,
    user_id    INTEGER   NOT NULL,
    expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- columns for auditing
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (sess_id),
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);
