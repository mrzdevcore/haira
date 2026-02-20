package haira

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// Workbook wraps an Excel file for reading.
type Workbook struct {
	file *excelize.File
}

// ExcelOpen opens an Excel (.xlsx) file for reading.
func ExcelOpen(filePath string) (*Workbook, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("excel open: %w", err)
	}
	return &Workbook{file: f}, nil
}

// SheetNames returns the list of sheet names in the workbook.
func (wb *Workbook) SheetNames() []any {
	names := wb.file.GetSheetList()
	result := make([]any, len(names))
	for i, n := range names {
		result[i] = n
	}
	return result
}

// ReadSheet reads a sheet by name and returns rows as a slice of maps.
// The first row is treated as header (column names).
// Returns each subsequent row as a map of column_name → value.
func (wb *Workbook) ReadSheet(name any) ([]map[string]any, error) {
	rows, err := wb.file.GetRows(Str(name))
	if err != nil {
		return nil, fmt.Errorf("excel read sheet %q: %w", name, err)
	}

	if len(rows) < 2 {
		return []map[string]any{}, nil
	}

	headers := rows[0]
	var results []map[string]any

	for _, row := range rows[1:] {
		record := make(map[string]any, len(headers))
		for i, header := range headers {
			if header == "" {
				continue
			}
			if i < len(row) {
				record[header] = row[i]
			} else {
				record[header] = ""
			}
		}
		results = append(results, record)
	}

	return results, nil
}

// Close closes the workbook and releases resources.
func (wb *Workbook) Close() error {
	return wb.file.Close()
}
