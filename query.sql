-- name: GetRunningManLibraries :many
SELECT * FROM running_man_library;

-- name: GetRunningManVideosByYear :many
SELECT * FROM running_man_video WHERE running_man_library_year = $1;