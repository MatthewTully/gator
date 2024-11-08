-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
RETURNING *;

-- name: GetPostsForUser :many
SELECT p.*, f.name AS feed_name FROM users u
INNER JOIN feed_follows ff ON u.id = ff.user_id
INNER JOIN feeds f ON f.id = ff.feed_id
INNER JOIN posts p ON f.id = p.feed_id
WHERE u.id = $1
ORDER BY p.published_at DESC
LIMIT $2;