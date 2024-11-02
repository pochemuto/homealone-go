-- name: GetPreference :one
SELECT user_id, notify_on_update FROM preference WHERE user_id = $1 LIMIT 1;

-- name: SavePreference :exec
INSERT INTO preference (user_id, notify_on_update)
VALUES ($1, $2)
ON CONFLICT (user_id)
DO UPDATE SET notify_on_update = EXCLUDED.notify_on_update;