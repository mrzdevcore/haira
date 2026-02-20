package haira

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ExcelSheet holds one sheet's data: headers, rows, and inferred column types.
type ExcelSheet struct {
	Name    string
	Headers []string
	Rows    []map[string]any
	Types   map[string]string
}

// ExcelTables holds all sheets read from an Excel file.
type ExcelTables struct {
	sheets map[string]*ExcelSheet
	order  []string
}

// ExcelReadSheets opens an Excel file and reads all sheets into an ExcelTables object.
func ExcelReadSheets(filePath string) (*ExcelTables, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("excel open: %w", err)
	}
	defer f.Close()

	tables := &ExcelTables{
		sheets: make(map[string]*ExcelSheet),
	}

	for _, name := range f.GetSheetList() {
		rows, err := f.GetRows(name)
		if err != nil {
			return nil, fmt.Errorf("excel read sheet %q: %w", name, err)
		}
		if len(rows) < 1 {
			continue
		}

		headers := rows[0]
		sheet := &ExcelSheet{
			Name:    name,
			Headers: headers,
			Types:   make(map[string]string),
		}

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
			sheet.Rows = append(sheet.Rows, record)
		}

		// Infer types from column headers (simple: everything is "string" from Excel)
		for _, h := range headers {
			if h != "" {
				sheet.Types[h] = "string"
			}
		}

		tables.sheets[name] = sheet
		tables.order = append(tables.order, name)
	}

	return tables, nil
}

// Names returns the sheet names in order.
func (t *ExcelTables) Names() []any {
	result := make([]any, len(t.order))
	for i, name := range t.order {
		result[i] = name
	}
	return result
}

// Sheet returns the rows for a given sheet name.
func (t *ExcelTables) Sheet(name any) []map[string]any {
	if s, ok := t.sheets[Str(name)]; ok {
		return s.Rows
	}
	return nil
}

// Headers returns the column headers for a given sheet name.
func (t *ExcelTables) SheetHeaders(name any) []any {
	if s, ok := t.sheets[Str(name)]; ok {
		result := make([]any, len(s.Headers))
		for i, h := range s.Headers {
			result[i] = h
		}
		return result
	}
	return nil
}

// Len returns the number of sheets.
func (t *ExcelTables) Len() int {
	return len(t.order)
}

// ValidateAgainst checks that all sheets and columns exist in the given schema.
// Schema format: map of table_name → map of column_name → column_type.
// Returns an error listing all mismatches, or nil if valid.
func (t *ExcelTables) ValidateAgainst(schema map[string]any) error {
	var errs []string

	for _, name := range t.order {
		sheet := t.sheets[name]
		tableSchema, ok := schema[name]
		if !ok {
			errs = append(errs, fmt.Sprintf("table %q not found in schema", name))
			continue
		}

		colMap, ok := tableSchema.(map[string]any)
		if !ok {
			errs = append(errs, fmt.Sprintf("schema for %q is not a map", name))
			continue
		}

		for _, header := range sheet.Headers {
			if header == "" {
				continue
			}
			if _, ok := colMap[header]; !ok {
				errs = append(errs, fmt.Sprintf("column %q in sheet %q not found in schema", header, name))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation failed:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}
