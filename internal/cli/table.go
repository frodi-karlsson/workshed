package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/frodi/workshed/internal/logger"
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

func Render(output Output, format string, w io.Writer) error {
	switch format {
	case "json":
		return renderJSONToWriter(output, w)
	case "raw":
		return renderRawToWriter(output, w)
	case "table":
		renderer := &FlexTableRenderer{}
		return renderer.Render(output.Columns, output.Rows, w)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func renderJSONToWriter(output Output, w io.Writer) error {
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

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func renderRawToWriter(output Output, w io.Writer) error {
	for _, row := range output.Rows {
		for i, cell := range row {
			if i > 0 {
				_, _ = fmt.Fprint(w, "\t")
			}
			_, _ = fmt.Fprint(w, cell)
		}
		_, _ = fmt.Fprintln(w)
	}
	return nil
}

var KeyValueColumns = []ColumnConfig{
	{Type: Rigid, Name: "KEY", Min: 10, Max: 20},
	{Type: Rigid, Name: "VALUE", Min: 20, Max: 0},
}

var ListColumns = []ColumnConfig{
	{Type: Rigid, Name: "HANDLE", Min: 15, Max: 20},
	{Type: Shrinkable, Name: "PURPOSE", Min: 15, Max: 0},
	{Type: Rigid, Name: "REPO", Min: 8, Max: 15},
	{Type: Rigid, Name: "CREATED", Min: 16, Max: 16},
}

var CapturesColumns = []ColumnConfig{
	{Type: Rigid, Name: "ID", Min: 26, Max: 26},
	{Type: Shrinkable, Name: "NAME", Min: 15, Max: 0},
	{Type: Rigid, Name: "KIND", Min: 8, Max: 15},
	{Type: Rigid, Name: "REPOS", Min: 6, Max: 8},
	{Type: Rigid, Name: "CREATED", Min: 16, Max: 16},
}

func RenderKeyValue(data map[string]string, format string, w io.Writer) error {
	var rows [][]string
	for k, v := range data {
		rows = append(rows, []string{k, v})
	}

	switch format {
	case "raw":
		for k, v := range data {
			_, _ = fmt.Fprintf(w, "%s=%s\n", k, v)
		}
		return nil
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	default:
		renderer := &FlexTableRenderer{}
		return renderer.Render(KeyValueColumns, rows, w)
	}
}

func RenderEmptyList(format string, message string, w io.Writer, l *logger.Logger) error {
	if format == "json" {
		_, _ = fmt.Fprintln(w, "[]")
		return nil
	}
	l.Info(message)
	return nil
}
