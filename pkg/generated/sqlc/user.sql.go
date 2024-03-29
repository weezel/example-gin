// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1
// source: user.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addUser = `-- name: AddUser :one
INSERT INTO homepage_schema.user(
        name,
        age,
        city,
        phone
) VALUES ($1, $2, $3, $4)
RETURNING id
`

type AddUserParams struct {
	Name  string      `json:"name"`
	Age   int32       `json:"age"`
	City  pgtype.Text `json:"city"`
	Phone pgtype.Text `json:"phone"`
}

func (q *Queries) AddUser(ctx context.Context, arg AddUserParams) (int32, error) {
	row := q.db.QueryRow(ctx, addUser,
		arg.Name,
		arg.Age,
		arg.City,
		arg.Phone,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const deleteUser = `-- name: DeleteUser :one
DELETE FROM homepage_schema.user
	WHERE name = $1
	RETURNING id, name, age, city, phone
`

func (q *Queries) DeleteUser(ctx context.Context, name string) (*HomepageSchemaUser, error) {
	row := q.db.QueryRow(ctx, deleteUser, name)
	var i HomepageSchemaUser
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Age,
		&i.City,
		&i.Phone,
	)
	return &i, err
}

const getUser = `-- name: GetUser :one
SELECT id, name, age, city, phone FROM homepage_schema.user
        WHERE name = $1
`

func (q *Queries) GetUser(ctx context.Context, name string) (*HomepageSchemaUser, error) {
	row := q.db.QueryRow(ctx, getUser, name)
	var i HomepageSchemaUser
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Age,
		&i.City,
		&i.Phone,
	)
	return &i, err
}

const listUsers = `-- name: ListUsers :many
SELECT id, name, age, city, phone FROM homepage_schema.user
`

func (q *Queries) ListUsers(ctx context.Context) ([]*HomepageSchemaUser, error) {
	rows, err := q.db.Query(ctx, listUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*HomepageSchemaUser
	for rows.Next() {
		var i HomepageSchemaUser
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Age,
			&i.City,
			&i.Phone,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
