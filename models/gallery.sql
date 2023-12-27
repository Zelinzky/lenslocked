-- name: create
INSERT INTO galleries (title, user_id)
VALUES (:title, :user_id)
RETURNING id;

-- name: by_id
SELECT title, user_id
FROM galleries
WHERE id = :id;

-- name: by_user_id
SELECT id, title
FROM galleries
WHERE user_id = $1;

-- name: update
UPDATE galleries
SET title = :title
WHERE id = :id;

-- name: delete
DELETE
FROM galleries
WHERE id = $1;