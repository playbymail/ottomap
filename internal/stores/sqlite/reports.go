// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sqlite

import (
	"database/sql"
	"errors"
	"strings"
)

// CreateNewReport creates a new report.
// Returns the id of the new report.
// Accepts empty input for lines.
func (s *Store) CreateNewReport(clan, year, month int, unit, hash string, lines []byte) (int, error) {
	if !(0 < clan && clan <= 1000) {
		return 0, ErrInvalidClanId
	} else if !(899 <= year && year <= 1234) {
		return 0, ErrInvalidYear
	} else if !(1 <= month && month <= 12) {
		return 0, ErrInvalidMonth
	} else if unit == "" {
		return 0, ErrInvalidUnit
	} else if hash == "" {
		return 0, ErrInvalidHash
	}
	id, err := s.q.CreateNewReport(s.ctx, CreateNewReportParams{
		Clan:  int64(clan),
		Year:  int64(year),
		Month: int64(month),
		Unit:  unit,
		Hash:  hash,
		Lines: string(lines),
	})
	if err != nil {
		// ugh. this is so fragile. we have to inspect the error string to figure out which constraint failed.
		if strings.HasPrefix(err.Error(), "constraint failed: UNIQUE constraint failed: reports.clan, reports.hash (") {
			return 0, ErrDuplicateHash
		} else if strings.HasPrefix(err.Error(), "constraint failed: UNIQUE constraint failed: reports.clan, reports.year, reports.month, reports.unit (") {
			return 0, ErrDuplicateReportName
		}
		return 0, err
	}
	return int(id), nil
}

// DeleteReportByHash deletes a report by its hash.
// Returns nil if no report is found.
func (s *Store) DeleteReportByHash(clan int, hash string) error {
	if !(0 < clan && clan <= 1000) {
		return ErrInvalidClanId
	}
	if err := s.q.DeleteReportByHash(s.ctx, DeleteReportByHashParams{
		Clan: int64(clan),
		Hash: hash,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	return nil
}

// DeleteReportByName deletes a report by its name (year, month and unit).
// Returns nil if no report is found.
func (s *Store) DeleteReportByName(clan, year, month int, unit string) error {
	if !(0 < clan && clan <= 1000) {
		return ErrInvalidClanId
	}
	if err := s.q.DeleteReportByName(s.ctx, DeleteReportByNameParams{
		Clan:  int64(clan),
		Year:  int64(year),
		Month: int64(month),
		Unit:  unit,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	return nil
}

// GetReportByHash returns a report by its hash.
// If no report is found, nil is return with no error.
func (s *Store) GetReportByHash(clan int, hash string) (*Report_t, error) {
	if !(0 < clan && clan <= 1000) {
		return nil, ErrInvalidClanId
	}
	row, err := s.q.GetReportByHash(s.ctx, GetReportByHashParams{
		Clan: int64(clan),
		Hash: hash,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// report not found is not an error
			return nil, nil
		}
		return nil, err
	}
	return &Report_t{
		ID:    int(row.ID),
		Clan:  int(row.Clan),
		Year:  int(row.Year),
		Month: int(row.Month),
		Unit:  row.Unit,
		Hash:  hash,
	}, nil
}

// GetReportsByTurn returns a list of reports for the requested clan, year, and month.
// If no reports are found, an empty list is returned.
func (s *Store) GetReportsByTurn(clan, year, month int) ([]*Report_t, error) {
	if !(0 < clan && clan <= 1000) {
		return nil, ErrInvalidClanId
	}
	rows, err := s.q.GetReportsByTurn(s.ctx, GetReportsByTurnParams{
		Clan:  int64(clan),
		Year:  int64(year),
		Month: int64(month),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// no reports found is not an error
			return nil, nil
		}
		return nil, err
	}
	var list []*Report_t
	for _, row := range rows {
		list = append(list, &Report_t{
			ID:    int(row.ID),
			Clan:  int(row.Clan),
			Year:  int(row.Year),
			Month: int(row.Month),
			Unit:  row.Unit,
			Hash:  row.Hash,
		})

	}
	return list, nil
}
