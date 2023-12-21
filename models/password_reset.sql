-- name: getUserID
SELECT id
FROM users
WHERE email = $1;

-- name: create
INSERT INTO password_resets (user_id, token_hash, expires_at)
VALUES (:user_id, :token_hash, :expires_at)
ON CONFLICT (user_id) DO UPDATE SET token_hash = :token_hash,
                                    expires_at = :expires_at
RETURNING id;

-- name: consume
SELECT password_resets.id         "pw_reset.id",
       password_resets.expires_at "pw_reset.expires_at",
       users.id                   "user.id",
       users.email                "user.email",
       users.password_hash        "user.password_hash"
FROM password_resets
         JOIN users ON users.id = password_resets.user_id
WHERE password_resets.token_hash = $1;

-- name: delete
DELETE
FROM password_resets
WHERE id = $1;