package entity

// DataSet is a generic tabular result: column names plus one cell per row.
// Every product/system has a different table with a different set of
// columns, so the repository (which knows the real schema) flattens its
// query result into this shape. Each cell is a string for plain scalar
// columns, or json.RawMessage for columns whose value is itself JSON — so it
// round-trips as real, unescaped JSON in an API response, not a JSON string
// containing JSON text.
type DataSet struct {
	Columns []string
	Rows    [][]any
}
