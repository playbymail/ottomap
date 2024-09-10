-- foreign keys must be disabled to drop tables with foreign keys
PRAGMA foreign_keys = OFF;

DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS server;
DROP TABLE IF EXISTS user_roles;
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

    -- columns for auditing
    created_at      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login      TIMESTAMP   NOT NULL
);

CREATE TABLE roles
(
    role_id     INTEGER PRIMARY KEY AUTOINCREMENT,
    role        TEXT UNIQUE NOT NULL,
    description TEXT        NOT NULL,

    -- columns for auditing
    created_at  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_roles
(
    user_id    INTEGER   NOT NULL,
    role_id    INTEGER   NOT NULL,

    -- columns for auditing
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles (role_id) ON DELETE CASCADE
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

CREATE TABLE server
(
    assets_path    TEXT NOT NULL,
    templates_path TEXT NOT NULL,
    salt           TEXT NOT NULL
);

-- insert roles
INSERT INTO roles (role, description)
VALUES ('admin', 'Administrator');
INSERT INTO roles (role, description)
VALUES ('operator', 'Operator');
INSERT INTO roles (role, description)
VALUES ('user', 'User');
INSERT INTO roles (role, description)
VALUES ('deleted', 'Deleted User');

-- create admin user
INSERT INTO users (user_id, email, is_active, hashed_password, clan, last_login)
VALUES (1, 'admin@ottomap', 0, '*', '0000', CURRENT_TIMESTAMP);

INSERT INTO user_roles (user_id, role_id)
SELECT user_id, role_id
FROM users,
     roles
WHERE users.email = 'admin@ottomap'
  AND roles.role = 'admin';

-- -- create clan 0138 user
-- INSERT INTO users (email, is_active, hashed_password, clan, last_login)
-- VALUES ('0138@ottomap', 0, '*', '0138', CURRENT_TIMESTAMP);
--
-- INSERT INTO user_roles (user_id, role_id)
-- SELECT user_id, role_id
-- FROM users,
--      roles
-- WHERE users.email = '0138@ottomap'
--   AND roles.role = 'user';
--
