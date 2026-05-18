package output

import (
	"bytes"
	"fmt"
)

func ExampleWriteJSON() {
	var buf bytes.Buffer
	_ = WriteJSON(&buf, map[string]string{"id": "abc123", "title": "Example"})
	fmt.Print(buf.String())
	// Output:
	// {"id":"abc123","title":"Example"}
}

func ExampleWriteTable() {
	var buf bytes.Buffer
	_ = WriteTable(&buf, []string{"ID", "TITLE"}, [][]string{{"abc123", "Example"}})
	fmt.Print(buf.String())
	// Output:
	// ID      TITLE
	// abc123  Example
}
