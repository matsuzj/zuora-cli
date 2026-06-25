package listcmd_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil/listcmd"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// memoSpec models the creditmemo/debitmemo first-wave shape: a static path,
// conditional string query flags, money columns, and a page-style nextPage.
func memoSpec() listcmd.Spec {
	return listcmd.Spec{
		Use:   "list",
		Short: "List demo memos",
		Flags: []listcmd.Flag{
			{Name: "account-id", Query: "accountId", Usage: "Filter by account ID"},
			{Name: "status", Query: "status", Usage: "Filter by status"},
			{Name: "page", Query: "page", Usage: "Page number"},
			{Name: "page-size", Query: "pageSize", Usage: "Results per page"},
		},
		Path:     func(args []string, flags map[string]string) string { return "/v1/memos" },
		ItemsKey: "memos",
		Columns: []listcmd.ColumnSpec{
			{Header: "ID", Key: "id"},
			{Header: "AMOUNT", Key: "amount", Money: true},
			{Header: "STATUS", Key: "status"},
		},
		NextPage: listcmd.NextPage{Flag: "page", FromURL: "page"},
	}
}

// cursorSpec models the account/list shape: an always-sent int page size with
// a default, a cursor flag, a repeatable filter, and a cursor-style nextPage.
func cursorSpec() listcmd.Spec {
	return listcmd.Spec{
		Use:   "list",
		Short: "List demo accounts",
		Flags: []listcmd.Flag{
			{Name: "page-size", Query: "pageSize", Usage: "Results per page", Int: true, IntDefault: 20},
			{Name: "cursor", Query: "cursor", Usage: "Pagination cursor"},
			{Name: "filter", Query: "filter[]", Usage: "Filter expressions", Repeatable: true},
		},
		Path:     func(args []string, flags map[string]string) string { return "/object-query/demo" },
		ItemsKey: "data",
		Columns: []listcmd.ColumnSpec{
			{Header: "ID", Key: "id"},
		},
		NextPage: listcmd.NextPage{Flag: "cursor"},
	}
}

// keySpec models the order list-by-* / subscription list shape: the path is
// built from a positional arg plus a required path-only flag.
func keySpec() listcmd.Spec {
	return listcmd.Spec{
		Use:  "list <key>",
		Args: cobra.ExactArgs(1),
		Flags: []listcmd.Flag{
			{Name: "account", Usage: "Account key", Required: true},
			{Name: "page", Query: "page", Usage: "Page number"},
		},
		Path: func(args []string, flags map[string]string) string {
			return fmt.Sprintf("/v1/demo/%s/accounts/%s", url.PathEscape(args[0]), url.PathEscape(flags["account"]))
		},
		ItemsKey: "items",
		Columns: []listcmd.ColumnSpec{
			{Header: "ID", Key: "id"},
		},
		NextPage: listcmd.NextPage{Flag: "page", FromURL: "page"},
	}
}

func newCmd(spec listcmd.Spec) func(*factory.Factory) *cobra.Command {
	return func(f *factory.Factory) *cobra.Command {
		return listcmd.New(f, spec)
	}
}

func TestList_TableCellsAndMoneyZeroValue(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/memos", map[string]interface{}{
		"memos": []map[string]interface{}{
			{"id": "m-1", "amount": 100.0, "status": "Posted"},
			{"id": "m-2", "status": nil}, // amount absent, status null
		},
	})

	stdout, _, err := cmdtest.Run(t, "demo", newCmd(memoSpec()), handler, "demo", "list")
	require.NoError(t, err)

	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "AMOUNT")
	assert.Contains(t, stdout, "m-1")
	assert.Contains(t, stdout, "100.00") // float64 renders %.2f
	assert.Contains(t, stdout, "m-2")
	assert.Contains(t, stdout, "0.00") // absent money key renders the typed-struct zero value
	assert.NotContains(t, stdout, "<nil>")
}

func TestList_MoneyCellNonFloatPassthrough(t *testing.T) {
	// JSON numbers unmarshal to float64, so Money cells normally hit the %.2f
	// path. But a money value delivered as a STRING is passed through verbatim —
	// NOT coerced to two decimals (a known divergence from a strict money
	// contract; see #352/F-29). Pin it so any change to the Money-cell rule is a
	// visible, reviewed edit rather than a silent display change.
	handler := cmdtest.OK(t, "GET", "/v1/memos", map[string]interface{}{
		"memos": []map[string]interface{}{
			{"id": "s-1", "amount": "100.5", "status": "Posted"}, // string money
			{"id": "f-1", "amount": 1000.0, "status": "Posted"},  // float64 money
		},
	})

	stdout, _, err := cmdtest.Run(t, "demo", newCmd(memoSpec()), handler, "demo", "list")
	require.NoError(t, err)

	assert.Contains(t, stdout, "100.5")     // string passes through verbatim...
	assert.NotContains(t, stdout, "100.50") // ...NOT re-formatted to two decimals
	assert.Contains(t, stdout, "1000.00")   // float64 still gets %.2f
}

func TestList_ConditionalQueryAssembly(t *testing.T) {
	var gotQuery url.Values
	handler := func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		assert.Equal(t, "/v1/memos", r.URL.Path)
		fmt.Fprint(w, `{"memos": []}`)
	}

	_, _, err := cmdtest.Run(t, "demo", newCmd(memoSpec()), handler,
		"demo", "list", "--status", "Posted", "--page-size", "5")
	require.NoError(t, err)

	assert.Equal(t, "Posted", gotQuery.Get("status"))
	assert.Equal(t, "5", gotQuery.Get("pageSize"))
	assert.False(t, gotQuery.Has("accountId"), "empty flags must not be sent")
	assert.False(t, gotQuery.Has("page"), "empty flags must not be sent")
}

func TestList_IntDefaultAlwaysSentAndRepeatable(t *testing.T) {
	var gotQuery url.Values
	handler := func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		fmt.Fprint(w, `{"data": []}`)
	}

	_, _, err := cmdtest.Run(t, "demo", newCmd(cursorSpec()), handler,
		"demo", "list", "--filter", "status.EQ:Active", "--filter", "name.EQ:Acme")
	require.NoError(t, err)

	assert.Equal(t, "20", gotQuery.Get("pageSize"), "int flag is always sent with its default")
	assert.Equal(t, []string{"status.EQ:Active", "name.EQ:Acme"}, gotQuery["filter[]"])
	assert.False(t, gotQuery.Has("cursor"))
}

func TestList_PathFromArgsAndPathOnlyFlag(t *testing.T) {
	var gotPath string
	var gotQuery url.Values
	handler := func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath() // keep %-escapes visible for the assertion
		gotQuery = r.URL.Query()
		fmt.Fprint(w, `{"items": []}`)
	}

	_, _, err := cmdtest.Run(t, "demo", newCmd(keySpec()), handler,
		"demo", "list", "K-001", "--account", "A 1")
	require.NoError(t, err)

	assert.Equal(t, "/v1/demo/K-001/accounts/A%201", gotPath)
	assert.False(t, gotQuery.Has("account"), "path-only flags must not be sent as query params")
}

func TestList_RequiredFlagEnforced(t *testing.T) {
	_, _, err := cmdtest.Run(t, "demo", newCmd(keySpec()), nil, "demo", "list", "K-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "account")
}

func TestList_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 58730122, "no permission")

	_, _, err := cmdtest.Run(t, "demo", newCmd(memoSpec()), handler, "demo", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no permission")
}

func TestList_ParseError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{invalid`)
	}

	_, _, err := cmdtest.Run(t, "demo", newCmd(memoSpec()), handler, "demo", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing response")
}

func TestList_JSONPassthroughNoHint(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/memos", map[string]interface{}{
		"memos":    []map[string]interface{}{{"id": "m-1"}},
		"nextPage": "https://api.example.com/v1/memos?page=2",
	})

	stdout, stderr, err := cmdtest.Run(t, "demo", newCmd(memoSpec()), handler,
		"demo", "list", "--json")
	require.NoError(t, err)

	assert.Contains(t, stdout, `"nextPage"`)
	assert.NotContains(t, stderr, "More results available", "--json output must not carry the hint")
}

func TestList_HintPageStyleReconstructsCommand(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/memos", map[string]interface{}{
		"memos":    []map[string]interface{}{{"id": "m-1"}},
		"nextPage": "https://api.example.com/v1/memos?page=3&pageSize=5",
	})

	_, stderr, err := cmdtest.Run(t, "demo", newCmd(memoSpec()), handler,
		"demo", "list", "--status", "Posted", "--page-size", "5")
	require.NoError(t, err)

	assert.Contains(t, stderr, "More results available")
	assert.Contains(t, stderr, "zr demo list --status Posted --page-size 5 --page 3")
}

func TestList_HintCursorStyleQuotesAndDefaults(t *testing.T) {
	spec := cursorSpec()
	handler := cmdtest.OK(t, "GET", "/object-query/demo", map[string]interface{}{
		"data":     []map[string]interface{}{{"id": "a-1"}},
		"nextPage": "cursor with spaces",
	})

	_, stderr, err := cmdtest.Run(t, "demo", newCmd(spec), handler, "demo", "list")
	require.NoError(t, err)

	assert.Contains(t, stderr, "More results available")
	assert.Contains(t, stderr, `zr demo list --cursor 'cursor with spaces'`)
	assert.NotContains(t, stderr, "--page-size", "default int values are not re-emitted")
}

func TestList_HintShellSafeQuoting(t *testing.T) {
	spec := cursorSpec()
	handler := cmdtest.OK(t, "GET", "/object-query/demo", map[string]interface{}{
		"data":     []map[string]interface{}{{"id": "a-1"}},
		"nextPage": "pre$TOKEN`cmd`'q'",
	})

	_, stderr, err := cmdtest.Run(t, "demo", newCmd(spec), handler, "demo", "list")
	require.NoError(t, err)

	// Single-quote escaping: $ and backticks must paste verbatim, embedded
	// single quotes via the standard '\'' sequence.
	assert.Contains(t, stderr, `--cursor 'pre$TOKEN`+"`cmd`"+`'\''q'\'''`)
}

func TestList_HintPositionalArgsAndNonDefaultInt(t *testing.T) {
	spec := cursorSpec()
	handler := cmdtest.OK(t, "GET", "/object-query/demo", map[string]interface{}{
		"data":     []map[string]interface{}{{"id": "a-1"}},
		"nextPage": "tok-1",
	})

	_, stderr, err := cmdtest.Run(t, "demo", newCmd(spec), handler,
		"demo", "list", "--page-size", "5", "--filter", "status.EQ:Active")
	require.NoError(t, err)

	assert.Contains(t, stderr, "zr demo list --page-size 5 --filter status.EQ:Active --cursor tok-1")
}

func TestList_HintFallbackWhenPageMissing(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/memos", map[string]interface{}{
		"memos":    []map[string]interface{}{{"id": "m-1"}},
		"nextPage": "https://api.example.com/v1/memos", // no page param to extract
	})

	_, stderr, err := cmdtest.Run(t, "demo", newCmd(memoSpec()), handler, "demo", "list")
	require.NoError(t, err)

	assert.Contains(t, stderr, "More results available. Use --json to see nextPage URL.")
	assert.NotContains(t, stderr, "Next page:")
}

func TestList_NoHintWithoutNextPage(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/memos", map[string]interface{}{
		"memos": []map[string]interface{}{{"id": "m-1"}},
	})

	_, stderr, err := cmdtest.Run(t, "demo", newCmd(memoSpec()), handler, "demo", "list")
	require.NoError(t, err)

	assert.NotContains(t, stderr, "More results available")
}

func TestList_JQAndTemplateSuppressHint(t *testing.T) {
	body := map[string]interface{}{
		"memos":    []map[string]interface{}{{"id": "m-1"}},
		"nextPage": "https://api.example.com/v1/memos?page=2",
	}

	for _, tc := range []struct {
		name string
		args []string
	}{
		{"jq", []string{"demo", "list", "--jq", ".memos"}},
		{"template", []string{"demo", "list", "--template", "{{.nextPage}}"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler := cmdtest.OK(t, "GET", "/v1/memos", body)
			_, stderr, err := cmdtest.Run(t, "demo", newCmd(memoSpec()), handler, tc.args...)
			require.NoError(t, err)
			assert.NotContains(t, stderr, "More results available")
		})
	}
}

func TestList_ItemsKeyNotArrayErrors(t *testing.T) {
	// The hand-written typed structs error on a non-array items key; a silent
	// empty table would hide a response-shape change.
	handler := cmdtest.OK(t, "GET", "/v1/memos", map[string]interface{}{
		"memos": map[string]interface{}{"unexpected": "object"},
	})

	_, _, err := cmdtest.Run(t, "demo", newCmd(memoSpec()), handler, "demo", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing response")
	assert.Contains(t, err.Error(), "memos")
}

func TestList_AbsentItemsKeyRendersZeroRows(t *testing.T) {
	// An absent items key matches the typed-struct zero slice: empty table.
	handler := cmdtest.OK(t, "GET", "/v1/memos", map[string]interface{}{})

	stdout, stderr, err := cmdtest.Run(t, "demo", newCmd(memoSpec()), handler, "demo", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ID")
	assert.NotContains(t, stderr, "More results available")
}

func TestList_IntFlagValueAvailableToPath(t *testing.T) {
	var gotPath string
	handler := func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		fmt.Fprint(w, `{"items": []}`)
	}

	spec := listcmd.Spec{
		Use: "list",
		Flags: []listcmd.Flag{
			{Name: "version", Usage: "Version segment", Int: true, IntDefault: 2},
		},
		Path: func(args []string, flags map[string]string) string {
			return "/v" + flags["version"] + "/items"
		},
		ItemsKey: "items",
		Columns:  []listcmd.ColumnSpec{{Header: "ID", Key: "id"}},
	}

	_, _, err := cmdtest.Run(t, "demo", newCmd(spec), handler, "demo", "list", "--version", "3")
	require.NoError(t, err)
	assert.Equal(t, "/v3/items", gotPath)
}

func TestList_HintStripsControlCharacters(t *testing.T) {
	// nextPage values must not smuggle terminal escape sequences into the
	// reconstructed command (same CWE-150 posture as the table renderer).
	spec := cursorSpec()
	handler := cmdtest.OK(t, "GET", "/object-query/demo", map[string]interface{}{
		"data":     []map[string]interface{}{{"id": "a-1"}},
		"nextPage": "tok\x1b[31mred\x1b[0m\nnext",
	})

	_, stderr, err := cmdtest.Run(t, "demo", newCmd(spec), handler, "demo", "list")
	require.NoError(t, err)

	assert.NotContains(t, stderr, "\x1b[31m")
	assert.Contains(t, stderr, "--cursor 'tok [31mred [0m next'")
}

// renamedSpec models a P5-1 flag rename: the canonical name is new, the old
// spelling is a deprecated alias, and the flag is required.
func renamedSpec() listcmd.Spec {
	return listcmd.Spec{
		Use: "list",
		Flags: []listcmd.Flag{
			{Name: "account-key", Usage: "Account key", Required: true, DeprecatedName: "account"},
		},
		Path: func(args []string, flags map[string]string) string {
			return "/v1/demo/accounts/" + url.PathEscape(flags["account-key"])
		},
		ItemsKey: "items",
		Columns: []listcmd.ColumnSpec{
			{Header: "ID", Key: "id"},
		},
	}
}

func TestList_DeprecatedAliasFeedsCanonicalFlag(t *testing.T) {
	var gotPath string
	handler := func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		fmt.Fprint(w, `{"items": []}`)
	}

	// The OLD spelling must keep working through the deprecation window and
	// must satisfy the requirement (cobra's own required check would not see
	// the alias — the runner enforces the shared value instead).
	_, _, err := cmdtest.Run(t, "demo", newCmd(renamedSpec()), handler,
		"demo", "list", "--account", "A-9")
	require.NoError(t, err)
	assert.Equal(t, "/v1/demo/accounts/A-9", gotPath)
}

func TestList_CanonicalNameWorksAlongsideAlias(t *testing.T) {
	var gotPath string
	handler := func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		fmt.Fprint(w, `{"items": []}`)
	}

	_, _, err := cmdtest.Run(t, "demo", newCmd(renamedSpec()), handler,
		"demo", "list", "--account-key", "A-10")
	require.NoError(t, err)
	assert.Equal(t, "/v1/demo/accounts/A-10", gotPath)
}

func TestList_RequiredWithAliasUsesCobraWording(t *testing.T) {
	_, _, err := cmdtest.Run(t, "demo", newCmd(renamedSpec()), nil, "demo", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "account-key" not set`)
}

func TestList_DeprecatedAliasHiddenAndMarked(t *testing.T) {
	// Inspect the flag set directly: cobra prints help to the process stdout
	// (root deliberately leaves cmd.SetOut unset), so help text is not
	// capturable through the cmdtest streams.
	cmd := listcmd.New(&factory.Factory{}, renamedSpec())

	alias := cmd.Flags().Lookup("account")
	require.NotNil(t, alias)
	assert.True(t, alias.Hidden, "deprecated alias must be hidden from help")
	assert.Equal(t, "use --account-key instead", alias.Deprecated)

	canonical := cmd.Flags().Lookup("account-key")
	require.NotNil(t, canonical)
	assert.False(t, canonical.Hidden)
	assert.Empty(t, canonical.Deprecated)
}

// TestList_OmitZeroIntFlagSkippedWhenUnset pins the P5-3c page-size
// normalization: an Int+OmitZero flag is sent only when non-zero.
func TestList_OmitZeroIntFlagSkippedWhenUnset(t *testing.T) {
	spec := listcmd.Spec{
		Use: "list",
		Flags: []listcmd.Flag{
			{Name: "page-size", Query: "pageSize", Usage: "n", Int: true, OmitZero: true},
		},
		Path:     func([]string, map[string]string) string { return "/v1/demo" },
		ItemsKey: "items",
		Columns:  []listcmd.ColumnSpec{{Header: "ID", Key: "id"}},
	}
	var got url.Values
	handler := func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query()
		fmt.Fprint(w, `{"items": []}`)
	}

	_, _, err := cmdtest.Run(t, "demo", newCmd(spec), handler, "demo", "list")
	require.NoError(t, err)
	assert.False(t, got.Has("pageSize"), "zero value must not be sent")

	_, _, err = cmdtest.Run(t, "demo", newCmd(spec), handler, "demo", "list", "--page-size", "5")
	require.NoError(t, err)
	assert.Equal(t, "5", got.Get("pageSize"))
}

func TestList_NonStringNextPageEmitsNoHintAndNoPanic(t *testing.T) {
	// The pagination hint reads nextPage via a `.(string)` assertion. A nextPage
	// delivered as a non-string (number / object / bool) must fail that assertion
	// to "" → no hint is emitted, and the command must not panic on the shape.
	for _, np := range []interface{}{12345, map[string]interface{}{"url": "x"}, true} {
		handler := cmdtest.OK(t, "GET", "/v1/memos", map[string]interface{}{
			"memos":    []map[string]interface{}{{"id": "m-1"}},
			"nextPage": np,
		})

		stdout, stderr, err := cmdtest.Run(t, "demo", newCmd(memoSpec()), handler, "demo", "list")
		require.NoError(t, err, "non-string nextPage (%T) must not error", np)
		assert.Contains(t, stdout, "m-1")
		assert.NotContains(t, stderr, "More results available", "non-string nextPage (%T) must not emit a hint", np)
	}
}
