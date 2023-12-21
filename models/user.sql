-- name: create
INSERT INTO users (email, password_hash)
VALUES (:email, :password_hash)
RETURNING id;

-- name: authenticate
SELECT *
FROM users
WHERE email = $1;

-- name: updatePass
UPDATE users
SET password_hash = $2
WHERE id = $1;