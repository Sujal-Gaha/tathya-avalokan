package proxy

import (
	"database/sql"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"tathya-avalokan/backend/internal/guard"
	"tathya-avalokan/backend/internal/models"
)

var (
	returningWordRe = regexp.MustCompile(`(?i)\bRETURNING\b`)
	singleQuoteRe   = regexp.MustCompile(`(?s)'(?:''|[^'])*'`)
	doubleQuoteRe   = regexp.MustCompile(`"[^"]*"`)
)

// IsRowReturningStatement determines whether a given SQL statement produces a result set (rows)
// or is an execution command (DDL/DML without a RETURNING clause).
func IsRowReturningStatement(sqlStr string) bool {
	kw := guard.ExtractFirstSQLKeyword(sqlStr)

	// Explicit read-only and metadata inspection statements always return rows
	switch kw {
	case "select", "show", "explain", "describe", "pragma":
		return true
	}

	// For DDL statements, always use ExecContext
	switch kw {
	case "create", "drop", "alter", "truncate":
		return false
	}

	// For INSERT, UPDATE, DELETE, and CTEs (WITH): check if a RETURNING clause exists outside string literals
	cleanSQL := singleQuoteRe.ReplaceAllString(sqlStr, "''")
	cleanSQL = doubleQuoteRe.ReplaceAllString(cleanSQL, `""`)

	if returningWordRe.MatchString(cleanSQL) {
		return true
	}

	if kw == "with" {
		// If CTE doesn't contain INSERT/UPDATE/DELETE, it's typically a SELECT
		isMutating, _ := guard.IsMutating(sqlStr)
		if !isMutating {
			return true
		}
	}

	// Non-returning INSERT, UPDATE, DELETE
	switch kw {
	case "insert", "update", "delete":
		return false
	}

	// Default fallback to true
	return true
}

// DisambiguateColumnNames ensures that duplicate column names in a result set
// (e.g. from JOIN queries: SELECT u.id, o.id) receive unique identifiers (id, id_1, id_2).
func DisambiguateColumnNames(rawCols []*sql.ColumnType) []models.ColumnMeta {
	seen := make(map[string]int)
	result := make([]models.ColumnMeta, len(rawCols))

	for i, col := range rawCols {
		name := col.Name()
		if name == "" {
			name = fmt.Sprintf("column_%d", i+1)
		}

		count := seen[name]
		seen[name] = count + 1

		uniqueName := name
		if count > 0 {
			uniqueName = fmt.Sprintf("%s_%d", name, count)
		}

		typeName := strings.ToLower(col.DatabaseTypeName())
		if typeName == "" {
			typeName = "unknown"
		}

		result[i] = models.ColumnMeta{
			Name: uniqueName,
			Type: typeName,
		}
	}

	return result
}

// NormalizeValue transforms raw driver scan outputs into clean JSON-serializable primitives.
// Specifically prevents Go []byte from being base64 encoded and sanitizes IEEE 754 NaN/Inf floats.
func NormalizeValue(v any) any {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case []byte:
		// Many drivers (MySQL, SQLite) scan TEXT/VARCHAR as []byte.
		// If valid UTF-8, convert to native string.
		if utf8.Valid(val) {
			return string(val)
		}
		// If binary, format as hex representation
		return fmt.Sprintf("\\x%x", val)

	case time.Time:
		return val.Format(time.RFC3339Nano)

	case float64:
		if math.IsNaN(val) {
			return "NaN"
		}
		if math.IsInf(val, 1) {
			return "+Infinity"
		}
		if math.IsInf(val, -1) {
			return "-Infinity"
		}
		return val

	case float32:
		f64 := float64(val)
		if math.IsNaN(f64) {
			return "NaN"
		}
		if math.IsInf(f64, 1) {
			return "+Infinity"
		}
		if math.IsInf(f64, -1) {
			return "-Infinity"
		}
		return f64

	case [16]byte:
		// Standard 16-byte UUID representation
		return fmt.Sprintf("%x-%x-%x-%x-%x", val[0:4], val[4:6], val[6:8], val[8:10], val[10:16])

	default:
		return val
	}
}

// ScanPaginatedRows iterates through an open *sql.Rows cursor, applying offset skipping,
// batch limit buffering, and a 1-row lookahead probe to determine has_more.
func ScanPaginatedRows(rows *sql.Rows, limit, offset int) ([]models.ColumnMeta, []map[string]any, bool, error) {
	defer rows.Close()

	if limit <= 0 {
		limit = 100
	} else if limit > 5000 {
		limit = 5000
	}

	if offset < 0 {
		offset = 0
	}

	rawColTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to retrieve column metadata: %w", err)
	}

	columns := DisambiguateColumnNames(rawColTypes)
	numCols := len(columns)

	// Fast path: query returned 0 columns
	if numCols == 0 {
		return []models.ColumnMeta{}, []map[string]any{}, false, nil
	}

	// 1. Skip offset rows
	if offset > 0 {
		dummyValues := make([]any, numCols)
		dummyArgs := make([]any, numCols)
		for j := range dummyValues {
			dummyArgs[j] = &dummyValues[j]
		}

		for i := 0; i < offset; i++ {
			if !rows.Next() {
				// Offset is beyond total rows
				if err := rows.Err(); err != nil {
					return nil, nil, false, err
				}
				return columns, []map[string]any{}, false, nil
			}
			_ = rows.Scan(dummyArgs...)
		}
	}

	// 2. Scan up to limit rows
	rowsList := make([]map[string]any, 0, min(limit, 100))
	for len(rowsList) < limit && rows.Next() {
		values := make([]any, numCols)
		scanArgs := make([]any, numCols)
		for j := range values {
			scanArgs[j] = &values[j]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, nil, false, fmt.Errorf("failed to scan row: %w", err)
		}

		rowMap := make(map[string]any, numCols)
		for j, col := range columns {
			rowMap[col.Name] = NormalizeValue(values[j])
		}
		rowsList = append(rowsList, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, false, err
	}

	// 3. Lookahead 1 row to evaluate has_more
	hasMore := false
	if len(rowsList) == limit && rows.Next() {
		hasMore = true
	}

	return columns, rowsList, hasMore, nil
}
