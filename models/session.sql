-- name: create
INSERT INTO sessions (user_id, token_hash)
VALUES (:user_id, :token_hash)
ON CONFLICT (user_id) DO UPDATE
    SET token_hash = :token_hash
RETURNING id;

-- name: user
SELECT u.id, u.email, u.password_hash
FROM sessions s
         JOIN users u ON u.id = s.user_id
WHERE s.token_hash = $1;

-- name: delete
DELETE
FROM sessions
WHERE token_hash = $1;