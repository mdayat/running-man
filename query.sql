-- name: CheckUserExistence :one
SELECT EXISTS(SELECT 1 FROM "user" WHERE id = $1);

-- name: CreateUser :exec
INSERT INTO "user" (id, first_name) VALUES ($1, $2);

-- name: GetRunningManLibraries :many
SELECT * FROM running_man_library;

-- name: GetRunningManVideosByYear :many
SELECT * FROM running_man_video WHERE running_man_library_year = $1;