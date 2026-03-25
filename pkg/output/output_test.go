package output

import (
	"bytes"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintTable(t *testing.T) {
	var buf bytes.Buffer
	cols := []Column{
		{Header: "ID", Field: "id"},
		{Header: "NAME", Field: "name"},
	}
	rows := [][]string{
		{"1", "Alice"},
		{"2", "Bob"},
	}
	err := PrintTable(&buf, rows, cols)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "Alice")
	assert.Contains(t, output, "Bob")
}

func TestPrintDetail(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	fields := []DetailField{
		{Key: "Name", Value: "Test Account"},
		{Key: "Status", Value: "Active"},
	}
	err := PrintDetail(ios, fields)
	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "Name:")
	assert.Contains(t, output, "Test Account")
	assert.Contains(t, output, "Status:")
	assert.Contains(t, output, "Active")
}

func TestPrintJSON_PrettyPrint(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	data := []byte(`{"name":"test","id":1}`)
	err := PrintJSON(ios, data, "")
	require.NoError(t, err)
	assert.Contains(t, out.String(), "\"name\": \"test\"")
}

func TestPrintJSON_WithJQ(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	data := []byte(`{"items":[{"name":"a"},{"name":"b"}]}`)
	err := PrintJSON(ios, data, ".items[].name")
	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "\"a\"")
	assert.Contains(t, output, "\"b\"")
}

func TestPrintTemplate(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	data := []byte(`{"name":"test","status":"Active"}`)
	err := PrintTemplate(ios, data, "{{.name}} is {{.status}}")
	require.NoError(t, err)
	assert.Contains(t, out.String(), "test is Active")
}

func TestPrintCSV(t *testing.T) {
	var buf bytes.Buffer
	cols := []Column{
		{Header: "ID", Field: "id"},
		{Header: "NAME", Field: "name"},
	}
	rows := [][]string{
		{"1", "Alice"},
		{"2", "Bob"},
	}
	err := PrintCSV(&buf, rows, cols)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "ID,NAME")
	assert.Contains(t, output, "1,Alice")
	assert.Contains(t, output, "2,Bob")
}

func TestRender_JSON(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	data := []byte(`{"ok":true}`)
	opts := FormatOptions{JSON: true}
	err := Render(ios, data, opts, nil, nil)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "\"ok\": true")
}

func TestRender_Table(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	data := []byte(`{}`)
	opts := FormatOptions{}
	cols := []Column{{Header: "ID"}, {Header: "NAME"}}
	rows := [][]string{{"1", "Test"}}
	err := Render(ios, data, opts, rows, cols)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Test")
}

func TestRenderDetail_JSON(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	data := []byte(`{"name":"test"}`)
	opts := FormatOptions{JSON: true}
	err := RenderDetail(ios, data, opts, nil)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "\"name\": \"test\"")
}

func TestRenderDetail_Detail(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	data := []byte(`{}`)
	opts := FormatOptions{}
	fields := []DetailField{{Key: "Name", Value: "Test"}}
	err := RenderDetail(ios, data, opts, fields)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Name:")
	assert.Contains(t, out.String(), "Test")
}
