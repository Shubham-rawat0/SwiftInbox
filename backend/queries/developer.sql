-- name: CreateDeveloper :one
INSERT INTO developer (
    id ,
    name,
    email
) VALUES ($1,$2,$3)
RETURNING id, name,email;