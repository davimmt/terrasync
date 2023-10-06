package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"

	"terrasync/src"
)

const (
	tfPlanPath        = "./terrasync.tf.plan"
	tfRootWorkingDir  = "/terraform"
	tfSyncTimeSeconds = "120"
)

func main() {
	tfExecPath, err := exec.LookPath("terraform")
	if err == nil {
		tfExecPath, err = filepath.Abs(tfExecPath)
	}
	if err != nil {
		log.Fatal(err)
	}

	rootWorkingDir := os.Getenv("TERRASYNC_ROOT_WORKING_DIR")
	if rootWorkingDir == "" {
		rootWorkingDir = tfRootWorkingDir
	}

	syncTimeSeconds := os.Getenv("TERRASYNC_SYNC_TIME_SECONDS")
	if syncTimeSeconds == "" {
		syncTimeSeconds = tfSyncTimeSeconds
	}

	syncTimeSecondsInt, err := strconv.Atoi(syncTimeSeconds)
	if err != nil {
		log.Fatalf("Error running Atoi: %s", err)
	}

	// Find all directories with .tf files
	dirs := src.FindTfDirs(rootWorkingDir)

	// Prepare variable to receive terrasyncChannel output
	result := make([]src.TfObject, len(dirs))

	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
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
	})

	// Open HTTP channel using goroutine so terraform plan routine runs forever
	// and updates the server respose
	httpChannel := make(chan bool)
	go http.ListenAndServe(":8080", nil)
	log.Println("Listening on port 8080...")

	terrasyncChannel := make(chan src.TfObject)
	for true {
		// Run terraform plan on all dirs with goroutine
		for _, dir := range dirs {
			go src.TfExec(dir, tfExecPath, tfPlanPath, terrasyncChannel)
		}

		// Get all terraform dirs output status from channel and assign to a variable
		for i := range result {
			result[i] = <-terrasyncChannel
			// fmt.Println(result[i])
		}

		time.Sleep(time.Duration(syncTimeSecondsInt) * time.Second)
	}

	<-httpChannel
}
