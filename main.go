package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"terrasync/src"
)

const (
	tfPlanPath        = "./terrasync.tf.plan"
	tfRootWorkingDir  = "/terraform"
	tfSyncTimeSeconds = "120"
	httpServerPort    = "8080"
)

func main() {
	// Find terraform binary
	tfExecPath, err := exec.LookPath("terraform")
	if err == nil {
		tfExecPath, err = filepath.Abs(tfExecPath)
	}
	if err != nil {
		log.Fatal(err)
	}

	// Configure variables
	serverPort := os.Getenv("TERRASYNC_HTTP_SERVER_PORT")
	if serverPort == "" {
		serverPort = httpServerPort
	}

	rootWorkingDir := os.Getenv("TERRASYNC_ROOT_WORKING_DIR")
	if rootWorkingDir == "" {
		rootWorkingDir = tfRootWorkingDir
	}

	syncTimeSeconds := os.Getenv("TERRASYNC_SYNC_TIME_SECONDS")
	if syncTimeSeconds == "" {
		syncTimeSeconds = tfSyncTimeSeconds
	}

	// Find all directories with .tf files
	dirs := src.FindTfDirs(rootWorkingDir)

	// Prepare variable to receive terrasyncChannel output
	result := make([]src.TfObject, len(dirs))

	// This is the function that outputs the results as HTTP respose
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		src.HttpServerHandler(w, result)
	})

	// Open HTTP server using goroutine so terraform plan routine runs forever
	// and updates the server respose
	go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", httpServerPort), nil)
	log.Printf("Listening on port %s...\n", httpServerPort)
	log.Printf("Directories found: %q", dirs)

	terrasyncChannel := make(chan src.TfObject)
	for true {
		// Run terraform plan on all dirs with goroutine
		for _, dir := range dirs {
			go src.TfExec(dir, tfExecPath, tfPlanPath, terrasyncChannel)
		}

		// Get all terraform dirs output status from channel and assign to a variable
		for i := range result {
			result[i] = <-terrasyncChannel
		}

		// Time to wait before next iteration
		s, err := strconv.Atoi(syncTimeSeconds)
		if err != nil {
			log.Fatalf("Error running Atoi: %s", err)
		}
		time.Sleep(time.Duration(s) * time.Second)
	}
}
