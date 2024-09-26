--  Copyright (c) 2024 Michael D Henderson. All rights reserved.

-- SetServerAssetsPath sets the path to the assets directory for the server.
--
-- name: SetServerAssetsPath :exec
UPDATE server
SET assets_path = :path;

-- SetServerTemplatesPath sets the path to the templates directory for the server.
--
-- name: SetServerTemplatesPath :exec
UPDATE server
SET templates_path = :path;

-- SetServerSalt sets the salt for the server.
--
-- name: SetServerSalt :exec
UPDATE server
SET salt = :salt;
