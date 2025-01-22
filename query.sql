-- name: CheckUserExistence :one
SELECT EXISTS(SELECT 1 FROM "user" WHERE id = $1);

-- name: CheckVideoOwnership :one
SELECT EXISTS(SELECT 1 FROM collection WHERE user_id = $1 AND running_man_video_episode = $2);

-- name: GetEpisodesFromUserVideoCollection :many
SELECT running_man_video_episode FROM collection WHERE user_id = $1;

-- name: CreateUser :exec
INSERT INTO "user" (id, first_name) VALUES ($1, $2);

-- name: GetRunningManYears :many
SELECT year FROM running_man_library ORDER BY year ASC;

-- name: GetRunningManEpisodesByYear :many
SELECT episode FROM running_man_video WHERE running_man_library_year = $1 ORDER BY episode ASC;

-- name: GetRunningManVideoPrice :one
SELECT price FROM running_man_video WHERE episode = $1;

-- name: GetRunningManVideoAndLibraryByEpisode :one
SELECT
  v.id AS running_man_video_id,
  l.id AS running_man_library_id,
  l.year
FROM running_man_video v JOIN running_man_library l ON v.running_man_library_year = l.year
WHERE v.episode = $1;

-- name: CheckInvoiceExpiration :one
SELECT EXISTS(SELECT 1 FROM invoice WHERE user_id = $1 AND running_man_video_episode = $2 AND expired_at > NOW());

-- name: CreateInvoice :exec
INSERT INTO invoice (id, user_id, running_man_video_episode, amount, expired_at) VALUES ($1, $2, $3, $4, $5);