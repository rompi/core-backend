package postgres

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ScanOne scans a single row from the result into a struct.
// It uses pgx's CollectOneRow with RowToStructByName.
func ScanOne[T any](rows pgx.Rows) (*T, error) {
	result, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[T])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNoRows
		}
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	return result, nil
}

// ScanAll scans all rows from the result into a slice of structs.
// It uses pgx's CollectRows with RowToStructByName.
func ScanAll[T any](rows pgx.Rows) ([]T, error) {
	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	return results, nil
}

// ScanMap scans a single row into a map[string]any.
func ScanMap(rows pgx.Rows) (map[string]any, error) {
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
		}
		return nil, ErrNoRows
	}

	descriptions := rows.FieldDescriptions()
	values, err := rows.Values()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}

	result := make(map[string]any, len(descriptions))
	for i, desc := range descriptions {
		result[desc.Name] = values[i]
	}

	return result, nil
}

// ScanAllMaps scans all rows into a slice of maps.
func ScanAllMaps(rows pgx.Rows) ([]map[string]any, error) {
	var results []map[string]any

	descriptions := rows.FieldDescriptions()

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
		}

		row := make(map[string]any, len(descriptions))
		for i, desc := range descriptions {
			row[desc.Name] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}

	return results, nil
}

// RowsAffected returns the number of rows affected by an Exec command.
func RowsAffected(tag pgconn.CommandTag) int64 {
	return tag.RowsAffected()
}
