--  Copyright (c) 2024 Michael D Henderson. All rights reserved.

-- --------------------------------------------------------------------------
-- CreateNewReport creates a new report.
--
-- name: CreateNewReport :one
INSERT INTO reports (clan, year, month, unit, hash, lines)
VALUES (:clan, :year, :month, :unit, :hash, :lines)
RETURNING id;

-- --------------------------------------------------------------------------
-- DeleteReportByHash deletes a report by its hash value.
--
-- name: DeleteReportByHash :exec
DELETE
FROM reports
WHERE clan = :clan
  AND hash = :hash;

-- --------------------------------------------------------------------------
-- DeleteReportByName returns a report by its name (year, month and unit).
--
-- name: DeleteReportByName :exec
DELETE
FROM reports
WHERE clan = :clan
  AND year = :year
  AND month = :month
  AND unit = :unit;

-- --------------------------------------------------------------------------
-- GetReportByHash returns a report by its hash value.
--
-- name: GetReportByHash :one
SELECT id, clan, year, month, unit
FROM reports
WHERE clan = :clan
  AND hash = :hash;

-- --------------------------------------------------------------------------
-- GetReportsByTurn returns a report by its turn number (year and month).
--
-- name: GetReportsByTurn :many
SELECT id, clan, year, month, unit, hash
FROM reports
WHERE clan = :clan
  AND year = :year
  AND month = :month;