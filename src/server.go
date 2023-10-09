package src

import (
	"net/http"
	"regexp"

	"github.com/jedib0t/go-pretty/v6/table"
)

func HttpServerHandler(w http.ResponseWriter, result []TfObject) {
	w.WriteHeader(http.StatusOK)
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.SetAllowedRowLength(210)
	t.AppendHeader(table.Row{"PATH", "SYNCED", "OUTPUT"})
	removeFromPath := regexp.MustCompile(`(/?[^/]+){1,5}$`)

	for i := range result {
		msg := result[i].Msg
		outOfSync := result[i].OutOfSync
		var synced string
		path := result[i].Path

		if result[i].Error {
			synced = "error"
		} else if !outOfSync {
			synced = "true"
			msg = "synced"
		} else if outOfSync {
			synced = "false"
		}

		if path == "" {
			path = "In progress..."
			msg = "---"
			synced = "unknown"
		} else {
			path = removeFromPath.FindString(path)
		}

		t.AppendRow([]interface{}{path, synced, msg})
		t.AppendSeparator()
	}

	t.Render()
}
