package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"time"

	"github.com/hashicorp/terraform-exec/tfexec"
)

const (
	tfPlanPath       = "./terrasync.tf.plan"
	tfRootWorkingDir = "/terraform"
	syncTimeSeconds  = 120
)

type TfObject struct {
	Path   string
	Synced string
	Msg    string
}

// var tfPlanOutClean = [2]string{
// 	"You can apply this plan to save these new output values to the Terraform",
// 	"state, without changing any real infrastructure.",
// }

func main() {
	// Get Terraform executable path from $PATH
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

	// Recursivly find all directories with .tf extension files
	dirs := []string{}
	filepath.WalkDir(rootWorkingDir, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}

		dir := filepath.Dir(s)
		if filepath.Ext(d.Name()) == ".tf" && !slices.Contains(dirs, dir) {
			dirs = append(dirs, dir)
		}

		return nil
	})

	// Execute Terraform commands on all directories
	for true {
		for _, workingDir := range dirs {
			obj := TfObject{Path: workingDir}
			tf, err := tfexec.NewTerraform(workingDir, tfExecPath)
			if err != nil {
				log.Fatalf("Error running NewTerraform: %s", err)
			}

			planOpts := []tfexec.PlanOption{
				tfexec.Lock(false),
				tfexec.Out(tfPlanPath),
			}

			plan, err := tf.Plan(context.Background(), planOpts...)
			if err != nil {
				obj.Synced = fmt.Sprintf("%q", plan)
				obj.Msg = fmt.Sprintf("%q", err) // if obj.Msg != "": error
				fmt.Printf("%+v\n", obj)
				continue
			}

			obj.Synced = fmt.Sprintf("%q", plan)
			fmt.Printf("%+v\n", obj)

			// if plan {
			// 	planOut, err := tf.ShowPlanFileRaw(context.Background(), tfPlanPath)
			// 	if err != nil {
			// 		log.Fatalf("error running ShowPlanFileRaw: %s", err)
			// 	}
			//
			// 	for _, s := range tfPlanOutClean {
			// 		planOut = strings.TrimSuffix(strings.ReplaceAll(planOut, s, ""), "\n")
			// 	}
			//
			// 	fmt.Printf("%q %s---\n", workingDir, planOut)
			// }
		}

		time.Sleep(syncTimeSeconds * time.Second)
	}
}
