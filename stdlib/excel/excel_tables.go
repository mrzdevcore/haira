package excel

import (
	"fmt"
	"strconv"
	"strings"

	haira "haira-go-runtime/haira"
	"haira-go-runtime/postgres"

	"github.com/xuri/excelize/v2"
)

// ExcelSheet holds one sheet's data: headers, rows, and inferred column types.
type ExcelSheet struct {
	Name      string
	Headers   []string
	Rows      []map[string]any
	Types     map[string]string
	TableType string // "configuration", "run", or "" (unclassified)
}

// ExcelTables holds all sheets read from an Excel file.
type ExcelTables struct {
	sheets map[string]*ExcelSheet
	order  []string
	// tableTypes maps table name → type ("configuration" or "run").
	// Set by ExcelReadConfig when mappings are provided.
	tableTypes map[string]string
}

// ExcelReadSheets opens an Excel file and reads all sheets into an ExcelTables object.
func ExcelReadSheets(filePath string) (*ExcelTables, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("excel open: %w", err)
	}
	defer f.Close()

	tables := &ExcelTables{
		sheets:     make(map[string]*ExcelSheet),
		tableTypes: make(map[string]string),
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

// ExcelReadConfig reads an Excel file with a sheet-to-table mapping.
// mappings is a map of Excel sheet name → map with keys:
//
//	"table" (string):      the database table name to use as the key
//	"table_type" (string): "configuration" or "run" (for seeds/oneshots split)
//
// Sheets not in the mapping are skipped.
// The returned ExcelTables uses table names as keys (not sheet names).
func ExcelReadConfig(filePath string, mappings map[string]any) (*ExcelTables, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("excel open: %w", err)
	}
	defer f.Close()

	// Build reverse mapping: sheet name → (table name, table type)
	type sheetInfo struct {
		tableName string
		tableType string
	}
	sheetMap := make(map[string]sheetInfo)
	for sheetName, v := range mappings {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		si := sheetInfo{tableName: haira.Str(m["table"]), tableType: haira.Str(m["table_type"])}
		if si.tableName == "" {
			si.tableName = sheetName
		}
		sheetMap[sheetName] = si
	}

	tables := &ExcelTables{
		sheets:     make(map[string]*ExcelSheet),
		tableTypes: make(map[string]string),
	}

	for _, name := range f.GetSheetList() {
		info, ok := sheetMap[name]
		if !ok {
			continue // Skip sheets not in mapping
		}

		rows, err := f.GetRows(name)
		if err != nil {
			return nil, fmt.Errorf("excel read sheet %q: %w", name, err)
		}
		if len(rows) < 1 {
			continue
		}

		headers := rows[0]
		sheet := &ExcelSheet{
			Name:      info.tableName,
			Headers:   headers,
			Types:     make(map[string]string),
			TableType: info.tableType,
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

		for _, h := range headers {
			if h != "" {
				sheet.Types[h] = "string"
			}
		}

		tables.sheets[info.tableName] = sheet
		tables.order = append(tables.order, info.tableName)
		tables.tableTypes[info.tableName] = info.tableType
	}

	return tables, nil
}

// Names returns the table names in order.
func (t *ExcelTables) Names() []any {
	result := make([]any, len(t.order))
	for i, name := range t.order {
		result[i] = name
	}
	return result
}

// Sheet returns the rows for a given table name.
func (t *ExcelTables) Sheet(name any) []map[string]any {
	if s, ok := t.sheets[haira.Str(name)]; ok {
		return s.Rows
	}
	return nil
}

// SheetHeaders returns the column headers for a given table name.
func (t *ExcelTables) SheetHeaders(name any) []any {
	if s, ok := t.sheets[haira.Str(name)]; ok {
		result := make([]any, len(s.Headers))
		for i, h := range s.Headers {
			result[i] = h
		}
		return result
	}
	return nil
}

// Len returns the number of sheets/tables.
func (t *ExcelTables) Len() int {
	return len(t.order)
}

// TableType returns the type of a table ("configuration", "run", or "").
func (t *ExcelTables) TableType(name any) string {
	return t.tableTypes[haira.Str(name)]
}

// Summary returns a human-readable summary of all tables with row/column counts.
func (t *ExcelTables) Summary() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Excel data: %d tables\n\n", len(t.order)))
	for _, name := range t.order {
		s := t.sheets[name]
		tt := t.tableTypes[name]
		if tt == "" {
			tt = "unclassified"
		}
		b.WriteString(fmt.Sprintf("- %s (%s): %d rows, %d columns\n",
			name, tt, len(s.Rows), len(s.Headers)))
	}
	return b.String()
}

// ToUiTable converts all sheets into a UiTable with tabs for direct UI rendering.
// Each tab shows up to maxRows rows (default 20). The data goes straight to the
// frontend via tool_render SSE — the LLM only receives a compact summary.
func (t *ExcelTables) ToUiTable(maxRows ...int) haira.UiTable {
	limit := 20
	if len(maxRows) > 0 && maxRows[0] > 0 {
		limit = maxRows[0]
	}

	var tabs []any
	for _, name := range t.order {
		sheet := t.sheets[name]
		hdrs := make([]any, len(sheet.Headers))
		for i, h := range sheet.Headers {
			hdrs[i] = h
		}

		rowCount := len(sheet.Rows)
		if rowCount > limit {
			rowCount = limit
		}
		rows := make([]any, rowCount)
		for i := 0; i < rowCount; i++ {
			cells := make([]any, len(sheet.Headers))
			for j, h := range sheet.Headers {
				cells[j] = haira.Str(sheet.Rows[i][h])
			}
			rows[i] = cells
		}

		tabName := fmt.Sprintf("%s (%d)", name, len(sheet.Rows))
		tabs = append(tabs, haira.UiTab{Name: tabName, Headers: hdrs, Rows: rows})
	}

	return haira.UiTable{
		Title: fmt.Sprintf("Configuration Data — %d tables", len(t.order)),
		Tabs:  tabs,
	}
}

// ToUiValidation converts a ValidateResult into a UiGroup with a status card
// and an error/warning table for direct UI rendering.
func (vr *ValidateResult) ToUiValidation() haira.UiGroup {
	var children []any

	// Status card
	status := "success"
	title := "Validation Passed"
	message := fmt.Sprintf("%d tables, %d rows — ready for SQL generation", vr.TablesCount, vr.TotalRows)
	if !vr.Valid {
		status = "error"
		title = "Validation Failed"
		message = fmt.Sprintf("%d errors, %d warnings across %d tables", vr.ErrorCount, vr.WarningCount, vr.TablesCount)
	}
	children = append(children, haira.UiStatusCard{Status: status, Title: title, Message: message})

	// Error table (if any)
	if vr.ErrorCount > 0 {
		limit := vr.ErrorCount
		if limit > 50 {
			limit = 50
		}
		hdrs := []any{"#", "Error"}
		rows := make([]any, limit)
		for i := 0; i < limit; i++ {
			rows[i] = []any{fmt.Sprintf("%d", i+1), haira.Str(vr.Errors[i])}
		}
		title := fmt.Sprintf("Validation Errors (%d)", vr.ErrorCount)
		if vr.ErrorCount > limit {
			title = fmt.Sprintf("Validation Errors (showing %d of %d)", limit, vr.ErrorCount)
		}
		children = append(children, haira.UiTable{Title: title, Headers: hdrs, Rows: rows})
	}

	return haira.UiGroup{Children: children}
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

// ValidateResult holds detailed validation results.
type ValidateResult struct {
	Valid        bool
	Errors       []any
	Warnings     []any
	ErrorCount   int
	WarningCount int
	TablesCount  int
	TotalRows    int
	Report       string
}

// ValidateWithSchema performs rich validation of Excel data against a PgSchema.
// Checks: unknown tables, missing required columns, unknown columns, null required fields,
// type mismatches (number/boolean), and duplicate primary keys.
func (t *ExcelTables) ValidateWithSchema(schema *postgres.PgSchema) *ValidateResult {
	var errors, warnings []string
	totalRows := 0

	for _, name := range t.order {
		sheet := t.sheets[name]
		ts, ok := schema.Tables[name]
		if !ok {
			warnings = append(warnings, fmt.Sprintf("[%s] table not found in database schema", name))
			continue
		}

		if len(sheet.Rows) == 0 {
			warnings = append(warnings, fmt.Sprintf("[%s] empty (no data rows)", name))
			continue
		}
		totalRows += len(sheet.Rows)

		// Check for missing required columns
		headerSet := make(map[string]bool, len(sheet.Headers))
		for _, h := range sheet.Headers {
			headerSet[h] = true
		}
		for colName, info := range ts.Columns {
			if !info.Nullable && !info.AutoGenerated && !info.PrimaryKey && !headerSet[colName] {
				errors = append(errors, fmt.Sprintf("[%s] missing required column %q", name, colName))
			}
		}

		// Check for unknown columns
		for _, h := range sheet.Headers {
			if h == "" {
				continue
			}
			if _, ok := ts.Columns[h]; !ok {
				warnings = append(warnings, fmt.Sprintf("[%s] unknown column %q (not in DB schema)", name, h))
			}
		}

		// Per-row validation
		pkSeen := make(map[string]bool)
		for rowIdx, row := range sheet.Rows {
			rowNum := rowIdx + 2 // Excel row number (1-indexed + header)

			for colName, val := range row {
				info, ok := ts.Columns[colName]
				if !ok {
					continue
				}
				strVal := haira.Str(val)

				// Check null required fields
				if strVal == "" && !info.Nullable && !info.AutoGenerated {
					errors = append(errors, fmt.Sprintf("[%s] row %d: required field %q is empty", name, rowNum, colName))
					continue
				}

				if strVal == "" {
					continue
				}

				// Type validation
				pgType := strings.ToLower(info.Type)
				switch {
				case isNumericPgType(pgType):
					if _, err := strconv.ParseFloat(strVal, 64); err != nil {
						errors = append(errors, fmt.Sprintf("[%s] row %d: column %q value %q is not a valid number", name, rowNum, colName, strVal))
					}
				case pgType == "boolean" || pgType == "bool":
					lower := strings.ToLower(strVal)
					if lower != "true" && lower != "false" && lower != "1" && lower != "0" && lower != "yes" && lower != "no" {
						errors = append(errors, fmt.Sprintf("[%s] row %d: column %q value %q is not a valid boolean", name, rowNum, colName, strVal))
					}
				}
			}

			// Check duplicate primary keys
			var pkParts []string
			for _, col := range ts.Order {
				if ci, ok := ts.Columns[col]; ok && ci.PrimaryKey {
					pkParts = append(pkParts, haira.Str(row[col]))
				}
			}
			if len(pkParts) > 0 {
				pkKey := strings.Join(pkParts, "|")
				if pkSeen[pkKey] {
					errors = append(errors, fmt.Sprintf("[%s] row %d: duplicate primary key %q", name, rowNum, pkKey))
				}
				pkSeen[pkKey] = true
			}
		}
	}

	// Build report
	valid := len(errors) == 0
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Validation: %d errors, %d warnings\n", len(errors), len(warnings)))
	b.WriteString(fmt.Sprintf("Tables: %d | Total rows: %d\n", len(t.order), totalRows))

	if len(errors) > 0 {
		b.WriteString("\nErrors:\n")
		limit := len(errors)
		if limit > 25 {
			limit = 25
		}
		for _, e := range errors[:limit] {
			b.WriteString("  " + e + "\n")
		}
		if len(errors) > 25 {
			b.WriteString(fmt.Sprintf("  ... and %d more errors\n", len(errors)-25))
		}
	}

	if len(warnings) > 0 {
		b.WriteString("\nWarnings:\n")
		limit := len(warnings)
		if limit > 10 {
			limit = 10
		}
		for _, w := range warnings[:limit] {
			b.WriteString("  " + w + "\n")
		}
		if len(warnings) > 10 {
			b.WriteString(fmt.Sprintf("  ... and %d more warnings\n", len(warnings)-10))
		}
	}

	if valid {
		b.WriteString("\nResult: PASSED — data is ready for SQL generation.")
	} else {
		b.WriteString("\nResult: FAILED — review errors above.")
	}

	// Convert to []any for Haira compatibility
	errAny := make([]any, len(errors))
	for i, e := range errors {
		errAny[i] = e
	}
	warnAny := make([]any, len(warnings))
	for i, w := range warnings {
		warnAny[i] = w
	}

	return &ValidateResult{
		Valid:        valid,
		Errors:       errAny,
		Warnings:     warnAny,
		ErrorCount:   len(errors),
		WarningCount: len(warnings),
		TablesCount:  len(t.order),
		TotalRows:    totalRows,
		Report:       b.String(),
	}
}

// SkipFirstRow removes the first data row from every sheet.
// Useful when Excel files have a type-hint or description row after the headers.
// If the row contains type keywords, they are used to populate sheet.Types.
func (t *ExcelTables) SkipFirstRow() {
	typeKeywords := map[string]bool{
		"string": true, "number": true, "boolean": true, "bool": true,
		"datetime": true, "date": true, "integer": true, "float": true,
		"text": true, "uuid": true, "int": true, "decimal": true,
		"timestamp": true, "bigint": true, "smallint": true, "numeric": true,
		"double": true, "real": true, "varchar": true, "char": true,
		"json": true, "jsonb": true, "array": true,
	}
	for _, name := range t.order {
		sheet := t.sheets[name]
		if len(sheet.Rows) == 0 {
			continue
		}
		// Check if row 0 is a type-hint row and use it for types
		row := sheet.Rows[0]
		isTypeRow := true
		for _, header := range sheet.Headers {
			if header == "" {
				continue
			}
			val := strings.TrimSpace(strings.ToLower(haira.Str(row[header])))
			if val != "" && !typeKeywords[val] {
				isTypeRow = false
				break
			}
		}
		if isTypeRow {
			for _, header := range sheet.Headers {
				if header == "" {
					continue
				}
				val := strings.TrimSpace(strings.ToLower(haira.Str(row[header])))
				if val != "" {
					sheet.Types[header] = val
				}
			}
		}
		sheet.Rows = sheet.Rows[1:]
	}
}

func isNumericPgType(t string) bool {
	switch t {
	case "integer", "int4", "int8", "bigint", "smallint",
		"numeric", "decimal", "real", "double precision", "float":
		return true
	}
	return false
}

// PostgresGenerateUpsert generates INSERT ... ON CONFLICT UPDATE SQL from ExcelTables and PgSchema.
// Uses PK columns for conflict detection. Tables are split into seeds/oneshots based on
// their TableType field set by ExcelReadConfig ("configuration" -> seeds, "run" -> oneshots).
func PostgresGenerateUpsert(tables *ExcelTables, schema *postgres.PgSchema) *postgres.SqlResult {
	return PostgresGenerateUpsertWithConflicts(tables, schema, nil)
}

// PostgresGenerateUpsertWithConflicts generates upsert SQL with optional per-table conflict columns.
// conflicts is a map of table_name -> []any (column names). If nil or missing for a table,
// the primary key columns from the schema are used as the conflict target.
func PostgresGenerateUpsertWithConflicts(tables *ExcelTables, schema *postgres.PgSchema, conflicts map[string]any) *postgres.SqlResult {
	var seeds, oneshots []string

	for _, name := range tables.order {
		sheet := tables.sheets[name]
		ts, ok := schema.Tables[name]
		if !ok || len(sheet.Rows) == 0 {
			continue
		}

		// Determine conflict columns: custom override or PK
		var conflictCols []string
		if conflicts != nil {
			if custom, ok := conflicts[name]; ok {
				switch v := custom.(type) {
				case []any:
					for _, c := range v {
						conflictCols = append(conflictCols, haira.Str(c))
					}
				case []string:
					conflictCols = v
				}
			}
		}
		if len(conflictCols) == 0 {
			for _, col := range ts.Order {
				if info, ok := ts.Columns[col]; ok && info.PrimaryKey {
					conflictCols = append(conflictCols, col)
				}
			}
		}

		// Find insertable columns (skip auto-generated)
		var insertCols []string
		for _, header := range sheet.Headers {
			if header == "" {
				continue
			}
			if info, ok := ts.Columns[header]; ok && !info.AutoGenerated {
				insertCols = append(insertCols, header)
			}
		}

		if len(insertCols) == 0 {
			continue
		}

		sql := generateTableUpsert(name, insertCols, conflictCols, sheet.Rows)

		// Split into seeds/oneshots based on table type
		tableType := tables.tableTypes[name]
		if tableType == "run" {
			oneshots = append(oneshots, sql)
		} else {
			seeds = append(seeds, sql)
		}
	}

	seedSQL := strings.Join(seeds, "\n\n")
	oneshotSQL := strings.Join(oneshots, "\n\n")

	return &postgres.SqlResult{
		Seeds:    seedSQL,
		Oneshots: oneshotSQL,
		All:      strings.TrimSpace(seedSQL + "\n\n" + oneshotSQL),
	}
}

func generateTableUpsert(tableName string, insertCols, conflictCols []string, rows []map[string]any) string {
	quotedTable := postgres.QuoteIdentifier(tableName)
	quotedInsertCols := make([]string, len(insertCols))
	for i, col := range insertCols {
		quotedInsertCols[i] = postgres.QuoteIdentifier(col)
	}

	// Build all value tuples
	var valueTuples []string
	for _, row := range rows {
		values := make([]string, len(insertCols))
		for i, col := range insertCols {
			values[i] = postgres.PostgresEscape(haira.Str(row[col]))
		}
		valueTuples = append(valueTuples, fmt.Sprintf("  (%s)", strings.Join(values, ", ")))
	}

	stmt := fmt.Sprintf("INSERT INTO %s (%s)\nVALUES\n%s",
		quotedTable,
		strings.Join(quotedInsertCols, ", "),
		strings.Join(valueTuples, ",\n"))

	if len(conflictCols) > 0 {
		conflictSet := make(map[string]bool, len(conflictCols))
		for _, c := range conflictCols {
			conflictSet[c] = true
		}
		quotedConflictCols := make([]string, len(conflictCols))
		for i, c := range conflictCols {
			quotedConflictCols[i] = postgres.QuoteIdentifier(c)
		}
		var updateCols []string
		for _, col := range insertCols {
			if !conflictSet[col] {
				qc := postgres.QuoteIdentifier(col)
				updateCols = append(updateCols, fmt.Sprintf("%s = EXCLUDED.%s", qc, qc))
			}
		}
		if len(updateCols) > 0 {
			stmt += fmt.Sprintf("\nON CONFLICT (%s) DO UPDATE SET\n  %s",
				strings.Join(quotedConflictCols, ", "),
				strings.Join(updateCols, ",\n  "))
		}
	}

	return fmt.Sprintf("-- %s\n%s;", tableName, stmt)
}
