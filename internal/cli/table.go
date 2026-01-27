package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/hchargois/flexwriter"
)

type ColumnType int

const (
	Rigid ColumnType = iota
	Shrinkable
)

type ColumnConfig struct {
	Type ColumnType
	Name string
	Min  int
	Max  int
}

type Format string

const (
	FormatTable  Format = "table"
	FormatJSON   Format = "json"
	FormatStream Format = "stream"
	FormatRaw    Format = "raw"
)

type Output struct {
	Columns []ColumnConfig
	Rows    [][]string
}

type TableRenderer interface {
	Render(columns []ColumnConfig, rows [][]string, out io.Writer) error
}

type OutputRenderer interface {
	Render(output Output, format Format, out io.Writer) error
}

type FlexTableRenderer struct{}

func (r *FlexTableRenderer) Render(columns []ColumnConfig, rows [][]string, out io.Writer) error {
	writer := flexwriter.New()
	writer.SetOutput(out)
	writer.SetDecorator(flexwriter.BoxDrawingTableDecorator())

	flexCols := make([]flexwriter.Column, len(columns))
	for i, col := range columns {
		switch col.Type {
		case Rigid:
			flexCols[i] = flexwriter.Rigid{Min: col.Min, Max: col.Max, Align: flexwriter.Left}
		case Shrinkable:
			flexCols[i] = flexwriter.Shrinkable{Min: col.Min, Max: col.Max, Align: flexwriter.Left}
		}
	}
	writer.SetColumns(flexCols...)

	headers := make([]any, len(columns))
	for i, col := range columns {
		headers[i] = col.Name
	}
	writer.WriteRow(headers...)

	for _, row := range rows {
		rowAny := make([]any, len(row))
		for i, v := range row {
			rowAny[i] = v
		}
		writer.WriteRow(rowAny...)
	}

	return writer.Flush()
}

type JSONRenderer struct{}

func (r *JSONRenderer) Render(output Output, format Format, out io.Writer) error {
	if format != FormatJSON {
		return nil
	}

	headers := make([]string, len(output.Columns))
	for i, col := range output.Columns {
		headers[i] = col.Name
	}

	result := make([]map[string]string, 0, len(output.Rows))
	for _, row := range output.Rows {
		if len(row) != len(headers) {
			continue
		}
		m := make(map[string]string)
		for i, v := range row {
			m[headers[i]] = v
		}
		result = append(result, m)
	}

	enc := json.NewEncoder(out)
	return enc.Encode(result)
}

type MockTableRenderer struct {
	Calls []struct {
		Columns []ColumnConfig
		Rows    [][]string
		Out     io.Writer
	}
	RenderError error
}

func (m *MockTableRenderer) Render(columns []ColumnConfig, rows [][]string, out io.Writer) error {
	m.Calls = append(m.Calls, struct {
		Columns []ColumnConfig
		Rows    [][]string
		Out     io.Writer
	}{Columns: columns, Rows: rows, Out: out})

	writer := flexwriter.New()
	writer.SetOutput(out)
	writer.SetDecorator(flexwriter.BoxDrawingTableDecorator())

	flexCols := make([]flexwriter.Column, len(columns))
	for i, col := range columns {
		flexCols[i] = flexwriter.Rigid{Min: col.Min, Max: col.Max, Align: flexwriter.Left}
	}
	writer.SetColumns(flexCols...)

	headers := make([]any, len(columns))
	for i, col := range columns {
		headers[i] = col.Name
	}
	writer.WriteRow(headers...)

	for _, row := range rows {
		rowAny := make([]any, len(row))
		for i, v := range row {
			rowAny[i] = v
		}
		writer.WriteRow(rowAny...)
	}

	return writer.Flush()
}

func (m *MockTableRenderer) LastCall() (columns []ColumnConfig, rows [][]string) {
	if len(m.Calls) == 0 {
		return nil, nil
	}
	last := m.Calls[len(m.Calls)-1]
	return last.Columns, last.Rows
}

func (m *MockTableRenderer) Reset() {
	m.Calls = nil
}

type MockOutputRenderer struct {
	Calls []struct {
		Output Output
		Format Format
		Out    io.Writer
	}
	RenderError error
}

func (m *MockOutputRenderer) Render(output Output, format Format, out io.Writer) error {
	m.Calls = append(m.Calls, struct {
		Output Output
		Format Format
		Out    io.Writer
	}{Output: output, Format: format, Out: out})

	if format == FormatTable {
		renderer := &FlexTableRenderer{}
		return renderer.Render(output.Columns, output.Rows, out)
	}

	if format == FormatJSON {
		renderer := &JSONRenderer{}
		return renderer.Render(output, format, out)
	}

	return nil
}

func (m *MockOutputRenderer) LastCall() (output Output, format Format) {
	if len(m.Calls) == 0 {
		return Output{}, ""
	}
	last := m.Calls[len(m.Calls)-1]
	return last.Output, last.Format
}

func (m *MockOutputRenderer) Reset() {
	m.Calls = nil
}

func RenderListTable(rows [][]string, stdout io.Writer) error {
	renderer := &FlexTableRenderer{}
	columns := []ColumnConfig{
		{Type: Rigid, Name: "HANDLE", Min: 15, Max: 20},
		{Type: Shrinkable, Name: "PURPOSE", Min: 15, Max: 0},
		{Type: Rigid, Name: "REPO", Min: 8, Max: 15},
		{Type: Rigid, Name: "CREATED", Min: 16, Max: 16},
	}
	return renderer.Render(columns, rows, stdout)
}

func RenderCapturesTable(rows [][]string, stdout io.Writer) error {
	renderer := &FlexTableRenderer{}
	columns := []ColumnConfig{
		{Type: Rigid, Name: "ID", Min: 26, Max: 26},
		{Type: Shrinkable, Name: "NAME", Min: 15, Max: 0},
		{Type: Rigid, Name: "KIND", Min: 8, Max: 15},
		{Type: Rigid, Name: "REPOS", Min: 6, Max: 8},
		{Type: Rigid, Name: "CREATED", Min: 16, Max: 16},
	}
	return renderer.Render(columns, rows, stdout)
}

func DetectFormatFromFilePath(path string) Format {
	if strings.HasSuffix(strings.ToLower(path), ".json") {
		return FormatJSON
	}
	return FormatTable
}

func ValidFormatsForCommand(cmd string) []string {
	switch cmd {
	case "exec":
		return []string{"stream", "json"}
	case "path":
		return []string{"raw", "table", "json"}
	case "list", "inspect", "captures", "repos", "create":
		return []string{"table", "json", "raw"}
	default:
		return []string{"table", "json"}
	}
}

func ValidateFormat(format Format, cmd string) error {
	valid := ValidFormatsForCommand(cmd)
	for _, v := range valid {
		if format == Format(v) {
			return nil
		}
	}
	return fmt.Errorf("unsupported format %q. Valid: %s", format, strings.Join(valid, "|"))
}
