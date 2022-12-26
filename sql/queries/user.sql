-- name: ListUsers :many
SELECT * FROM homepage_schema.user;

-- name: AddUser :one
INSERT INTO homepage_schema.user(
        name,
        age,
        city,
        phone
) VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: DeleteUser :one
DELETE FROM homepage_schema.user
	WHERE id = $1
                AND name = $2
	RETURNING *;

