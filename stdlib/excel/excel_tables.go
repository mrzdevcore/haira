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

	// Build reverse mapping: sheet name → (table name, table type, column aliases)
	type sheetInfo struct {
		tableName     string
		tableType     string
		columnAliases map[string]string // Excel column name → DB column name
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
		// Parse optional column_aliases: { "excel_col": "db_col", ... }
		if aliases, ok := m["column_aliases"]; ok {
			if am, ok := aliases.(map[string]any); ok {
				si.columnAliases = make(map[string]string, len(am))
				for k, av := range am {
					si.columnAliases[k] = haira.Str(av)
				}
			}
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

		rawHeaders := rows[0]
		// Apply column aliases: rename Excel column names to DB column names
		headers := make([]string, len(rawHeaders))
		for i, h := range rawHeaders {
			if alias, ok := info.columnAliases[h]; ok {
				headers[i] = alias
			} else {
				headers[i] = h
			}
		}
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
// Each tab shows up to maxRows rows (0 = all rows). The data goes straight to the
// frontend via tool_render SSE — the LLM only receives a compact summary.
func (t *ExcelTables) ToUiTable(maxRows ...int) haira.UiTable {
	limit := 0
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
		if limit > 0 && rowCount > limit {
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
	} else if vr.WarningCount > 0 {
		status = "warning"
		title = "Validation Passed with Warnings"
		message = fmt.Sprintf("%d tables, %d rows — %d columns not found in DB schema (dropped from SQL)", vr.TablesCount, vr.TotalRows, vr.WarningCount)
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

	// Warning table (if any) — always show so dropped columns are visible
	if vr.WarningCount > 0 {
		limit := vr.WarningCount
		if limit > 50 {
			limit = 50
		}
		hdrs := []any{"#", "Warning"}
		rows := make([]any, limit)
		for i := 0; i < limit; i++ {
			rows[i] = []any{fmt.Sprintf("%d", i+1), haira.Str(vr.Warnings[i])}
		}
		wtitle := fmt.Sprintf("Warnings (%d)", vr.WarningCount)
		if vr.WarningCount > limit {
			wtitle = fmt.Sprintf("Warnings (showing %d of %d)", limit, vr.WarningCount)
		}
		children = append(children, haira.UiTable{Title: wtitle, Headers: hdrs, Rows: rows})
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

// FKFix describes a foreign key violation with the missing values
// and enough context to generate a fix suggestion.
type FKFix struct {
	ParentTable   string
	ParentColumn  string
	ChildTable    string
	ChildColumn   string
	MissingValues []string
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
	FKFixes      []FKFix
}

// validateCore performs structural validation of Excel data against a PgSchema:
// unknown tables, missing required columns, unknown columns, null required fields,
// type mismatches (number/boolean), and duplicate primary keys.
// It does NOT check cross-table FK references (use ValidateWithSchema for that).
func (t *ExcelTables) validateCore(schema *postgres.PgSchema) (errors, warnings []string, totalRows int) {
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
			if !info.Nullable && !info.HasDefault && !info.AutoGenerated && !info.PrimaryKey && !headerSet[colName] {
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
	return
}

// buildValidateResult constructs a ValidateResult from errors, warnings, fkFixes, and metadata.
func buildValidateResult(errors, warnings []string, totalRows, tablesCount int, fkFixes []FKFix) *ValidateResult {
	valid := len(errors) == 0
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Validation: %d errors, %d warnings\n", len(errors), len(warnings)))
	b.WriteString(fmt.Sprintf("Tables: %d | Total rows: %d\n", tablesCount, totalRows))

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
		TablesCount:  tablesCount,
		TotalRows:    totalRows,
		Report:       b.String(),
		FKFixes:      fkFixes,
	}
}

// ValidateStructure validates column/type/null/PK structure only — no FK cross-checks.
// Use this when the workflow will rely on SQL dry-run to catch FK violations against the
// live database (zero false positives). This avoids the false positives that occur when
// parent FK values exist in the database but not in the current Excel file.
func (t *ExcelTables) ValidateStructure(schema *postgres.PgSchema) *ValidateResult {
	errors, warnings, totalRows := t.validateCore(schema)
	return buildValidateResult(errors, warnings, totalRows, len(t.order), nil)
}

// ValidateWithSchema performs rich validation of Excel data against a PgSchema.
// Checks: unknown tables, missing required columns, unknown columns, null required fields,
// type mismatches (number/boolean), duplicate primary keys, AND cross-table FK references.
// Note: FK cross-checks only look at values within the Excel file. For values that may
// already exist in the database, use ValidateWithDb or prefer ValidateStructure + dry-run.
func (t *ExcelTables) ValidateWithSchema(schema *postgres.PgSchema) *ValidateResult {
	errors, warnings, totalRows := t.validateCore(schema)

	// Cross-table FK value validation: for each FK where both child and parent
	// tables are in the Excel, verify that every FK column value in the child
	// exists in the parent table's referenced column. Missing values indicate
	// rows that need to be added to the parent table.
	var fkFixes []FKFix
	for _, fk := range schema.FKColumns {
		childSheet, childOk := t.sheets[fk.ChildTable]
		parentSheet, parentOk := t.sheets[fk.ParentTable]
		if !childOk || !parentOk {
			continue // both tables must be in the Excel for local cross-checking
		}
		// Check the child column exists in Excel headers
		hasChildCol := false
		for _, h := range childSheet.Headers {
			if h == fk.ChildColumn {
				hasChildCol = true
				break
			}
		}
		if !hasChildCol {
			continue
		}

		// Collect all values present in the parent table's referenced column
		parentValues := make(map[string]bool, len(parentSheet.Rows))
		for _, row := range parentSheet.Rows {
			val := haira.Str(row[fk.ParentColumn])
			if val != "" {
				parentValues[val] = true
			}
		}

		// Find child FK values missing from the parent
		var missing []string
		missingSet := make(map[string]bool)
		for _, row := range childSheet.Rows {
			val := haira.Str(row[fk.ChildColumn])
			if val != "" && !parentValues[val] && !missingSet[val] {
				missing = append(missing, val)
				missingSet[val] = true
			}
		}

		if len(missing) > 0 {
			for _, v := range missing {
				errors = append(errors, fmt.Sprintf(
					"[%s] FK violation: column %q value %q not found in %s.%s — add it to the %q sheet",
					fk.ChildTable, fk.ChildColumn, v, fk.ParentTable, fk.ParentColumn, fk.ParentTable))
			}
			fkFixes = append(fkFixes, FKFix{
				ParentTable:   fk.ParentTable,
				ParentColumn:  fk.ParentColumn,
				ChildTable:    fk.ChildTable,
				ChildColumn:   fk.ChildColumn,
				MissingValues: missing,
			})
		}
	}

	return buildValidateResult(errors, warnings, totalRows, len(t.order), fkFixes)
}

// ValidateWithDB performs the same validation as ValidateWithSchema, but additionally
// checks FK parent values against the live database to eliminate false positives.
// Values that already exist in the DB (from prior migrations) are not reported as missing.
func (t *ExcelTables) ValidateWithDb(schema *postgres.PgSchema, db *postgres.DB) *ValidateResult {
	result := t.ValidateWithSchema(schema)
	if db == nil || len(result.FKFixes) == 0 {
		return result
	}

	// For each FKFix, query the DB to check which "missing" values actually exist
	var realFixes []FKFix
	var realErrors []string

	for _, fix := range result.FKFixes {
		// Build parameterized IN clause: WHERE parent_col IN ($1, $2, ...)
		placeholders := make([]string, len(fix.MissingValues))
		args := make([]any, len(fix.MissingValues))
		for i, v := range fix.MissingValues {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = v
		}
		query := fmt.Sprintf("SELECT DISTINCT %s FROM %s WHERE %s IN (%s)",
			postgres.QuoteIdentifier(fix.ParentColumn),
			postgres.QuoteIdentifier(fix.ParentTable),
			postgres.QuoteIdentifier(fix.ParentColumn),
			strings.Join(placeholders, ", "))
		rows, err := db.Query(query, args...)

		existsInDB := make(map[string]bool)
		if err == nil {
			for _, row := range rows {
				val := haira.Str(row[fix.ParentColumn])
				if val != "" {
					existsInDB[val] = true
				}
			}
		}

		// Filter to only truly missing values
		var stillMissing []string
		for _, v := range fix.MissingValues {
			if !existsInDB[v] {
				stillMissing = append(stillMissing, v)
			}
		}

		if len(stillMissing) > 0 {
			realFixes = append(realFixes, FKFix{
				ParentTable:   fix.ParentTable,
				ParentColumn:  fix.ParentColumn,
				ChildTable:    fix.ChildTable,
				ChildColumn:   fix.ChildColumn,
				MissingValues: stillMissing,
			})
			for _, v := range stillMissing {
				realErrors = append(realErrors, fmt.Sprintf(
					"[%s] FK violation: column %q value %q not found in %s.%s (not in Excel nor database)",
					fix.ChildTable, fix.ChildColumn, v, fix.ParentTable, fix.ParentColumn))
			}
		}
	}

	// Rebuild errors: keep non-FK errors, replace with only real FK errors
	fkPrefix := "] FK violation: column "
	var filteredErrors []string
	for _, e := range result.Errors {
		s := haira.Str(e)
		if !strings.Contains(s, fkPrefix) {
			filteredErrors = append(filteredErrors, s)
		}
	}
	filteredErrors = append(filteredErrors, realErrors...)

	// Convert to []any
	errAny := make([]any, len(filteredErrors))
	for i, e := range filteredErrors {
		errAny[i] = e
	}

	// Rebuild report
	valid := len(filteredErrors) == 0
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Validation: %d errors, %d warnings\n", len(filteredErrors), result.WarningCount))
	b.WriteString(fmt.Sprintf("Tables: %d | Total rows: %d\n", result.TablesCount, result.TotalRows))
	if len(filteredErrors) > 0 {
		b.WriteString("\nErrors:\n")
		limit := len(filteredErrors)
		if limit > 25 {
			limit = 25
		}
		for _, e := range filteredErrors[:limit] {
			b.WriteString("  " + e + "\n")
		}
		if len(filteredErrors) > 25 {
			b.WriteString(fmt.Sprintf("  ... and %d more errors\n", len(filteredErrors)-25))
		}
	}
	if result.WarningCount > 0 {
		b.WriteString("\nWarnings:\n")
		warnLimit := result.WarningCount
		if warnLimit > 10 {
			warnLimit = 10
		}
		for i := 0; i < warnLimit && i < len(result.Warnings); i++ {
			b.WriteString("  " + haira.Str(result.Warnings[i]) + "\n")
		}
		if result.WarningCount > 10 {
			b.WriteString(fmt.Sprintf("  ... and %d more warnings\n", result.WarningCount-10))
		}
	}
	if valid {
		b.WriteString("\nResult: PASSED — data is ready for SQL generation.")
	} else {
		b.WriteString("\nResult: FAILED — review errors above.")
	}

	return &ValidateResult{
		Valid:        valid,
		Errors:       errAny,
		Warnings:     result.Warnings,
		ErrorCount:   len(filteredErrors),
		WarningCount: result.WarningCount,
		TablesCount:  result.TablesCount,
		TotalRows:    result.TotalRows,
		Report:       b.String(),
		FKFixes:      realFixes,
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

	// Sort tables by FK dependency order so referenced (parent) tables are inserted first.
	// This prevents "violates foreign key constraint" errors when both parent and child rows
	// are in the same batch.
	orderedNames := postgres.TopoSort(tables.order, schema.FKDeps)

	for _, name := range orderedNames {
		sheet := tables.sheets[name]
		ts, ok := schema.Tables[name]
		if !ok || len(sheet.Rows) == 0 {
			continue
		}

		// Determine conflict columns: custom override or PK
		var conflictCols []string
		usingCustomConflict := false
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
				if len(conflictCols) > 0 {
					usingCustomConflict = true
				}
			}
		}
		if len(conflictCols) == 0 {
			// Use PK columns as conflict target.
			// If ALL PK columns are auto-generated AND none of them appear in the
			// Excel headers (i.e. the user has no values for them), fall back to
			// the first unique index so ON CONFLICT targets a business-key constraint.
			// When the user explicitly provides PK values in the Excel (even for serial
			// columns), use the PK — they are doing an explicit upsert by ID.
			var pkCols []string
			allPKAutoGeneratedAndAbsent := true
			headerSet := make(map[string]bool, len(sheet.Headers))
			for _, h := range sheet.Headers {
				headerSet[h] = true
			}
			for _, col := range ts.Order {
				if info, ok := ts.Columns[col]; ok && info.PrimaryKey {
					pkCols = append(pkCols, col)
					if !info.AutoGenerated || headerSet[col] {
						allPKAutoGeneratedAndAbsent = false
					}
				}
			}
			if len(pkCols) > 0 && !allPKAutoGeneratedAndAbsent {
				conflictCols = pkCols
			} else if len(ts.UniqueIndexes) > 0 {
				// All PKs are serial and not in Excel — use the first unique index instead
				conflictCols = ts.UniqueIndexes[0]
			} else {
				conflictCols = pkCols // last resort
			}
		}

		// Find insertable columns: include any column present in the Excel headers
		// that exists in the DB schema. All columns including auto-generated PKs are
		// included — the Excel's UUID values are preserved in the INSERT.
		var insertCols []string
		for _, header := range sheet.Headers {
			if header == "" {
				continue
			}
			if _, ok := ts.Columns[header]; ok {
				insertCols = append(insertCols, header)
			}
		}

		if len(insertCols) == 0 {
			continue
		}

		// When using custom conflict columns (business key), collect auto-generated
		// PK columns to exclude from the DO UPDATE SET. On business key match, the
		// existing row keeps its DB-assigned PK; only non-key columns are updated.
		var excludeFromUpdate []string
		if usingCustomConflict {
			for _, col := range ts.Order {
				if info, ok := ts.Columns[col]; ok && info.AutoGenerated && info.PrimaryKey {
					excludeFromUpdate = append(excludeFromUpdate, col)
				}
			}
		}

		sql := generateTableUpsert(name, insertCols, conflictCols, excludeFromUpdate, sheet.Rows)

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

// deduplicateRows removes duplicate rows that share the same conflict-column key,
// keeping only the first occurrence. This prevents the PostgreSQL error
// "ON CONFLICT DO UPDATE command cannot affect row a second time".
func deduplicateRows(rows []map[string]any, conflictCols []string) []map[string]any {
	if len(conflictCols) == 0 {
		return rows
	}
	seen := make(map[string]bool, len(rows))
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		parts := make([]string, len(conflictCols))
		for i, c := range conflictCols {
			parts[i] = haira.Str(row[c])
		}
		key := strings.Join(parts, "\x00")
		if !seen[key] {
			seen[key] = true
			out = append(out, row)
		}
	}
	return out
}

func generateTableUpsert(tableName string, insertCols, conflictCols, excludeFromUpdate []string, rows []map[string]any) string {
	// Deduplicate rows by conflict columns to avoid the PostgreSQL error:
	// "ON CONFLICT DO UPDATE command cannot affect row a second time"
	rows = deduplicateRows(rows, conflictCols)

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
			val := haira.Str(row[col])
			if val == "" {
				values[i] = "NULL"
			} else {
				values[i] = postgres.PostgresEscape(val)
			}
		}
		valueTuples = append(valueTuples, fmt.Sprintf("  (%s)", strings.Join(values, ", ")))
	}

	stmt := fmt.Sprintf("INSERT INTO %s (%s)\nVALUES\n%s",
		quotedTable,
		strings.Join(quotedInsertCols, ", "),
		strings.Join(valueTuples, ",\n"))

	if len(conflictCols) > 0 {
		// Build exclusion set: conflict columns + any extra columns (e.g. auto-generated PKs)
		// that should not appear in the DO UPDATE SET clause.
		excludeSet := make(map[string]bool, len(conflictCols)+len(excludeFromUpdate))
		for _, c := range conflictCols {
			excludeSet[c] = true
		}
		for _, c := range excludeFromUpdate {
			excludeSet[c] = true
		}
		quotedConflictCols := make([]string, len(conflictCols))
		for i, c := range conflictCols {
			quotedConflictCols[i] = postgres.QuoteIdentifier(c)
		}
		var updateCols []string
		for _, col := range insertCols {
			if !excludeSet[col] {
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
