-- name: CheckUserExistence :one
SELECT EXISTS(SELECT 1 FROM "user" WHERE id = $1);

-- name: CheckUserVideo :one
SELECT EXISTS(SELECT 1 FROM collection WHERE user_id = $1 AND running_man_video_episode = $2);

-- name: CreateUser :exec
INSERT INTO "user" (id, first_name) VALUES ($1, $2);

-- name: GetRunningManLibraries :many
SELECT year FROM running_man_library ORDER BY year ASC;

-- name: GetRunningManVideosByYear :many
SELECT episode FROM running_man_video WHERE running_man_library_year = $1 ORDER BY episode ASC;

-- name: GetRunningManVideoPrice :one
SELECT price FROM running_man_video WHERE episode = $1;