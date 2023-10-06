package src

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
)

type TfObject struct {
	Path      string `json:"path"`
	OutOfSync bool   `json:"outOfSync"`
	Error     bool   `json:"error"`
	Msg       string `json:"msg"`
}

var tfPlanOutClean = [2]string{
	"You can apply this plan to save these new output values to the Terraform",
	"state, without changing any real infrastructure.",
}

func TfExec(workingDir string, tfExecPath string, tfPlanPath string, c chan TfObject) {
	obj := TfObject{Path: workingDir}
	tf, err := tfexec.NewTerraform(workingDir, tfExecPath)
	if err != nil {
		log.Fatalf("Error running NewTerraform: %s", err)
	}

	obj.Error = false

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	planOpts := []tfexec.PlanOption{
		tfexec.Lock(false),
		tfexec.Out(tfPlanPath),
	}
	plan, err := tf.Plan(context.Background(), planOpts...)
	if err != nil {
		obj.Error = true
		obj.Msg = fmt.Sprintf("%s", err)
	}

	obj.OutOfSync = plan

	if plan {
		planOut, err := tf.ShowPlanFileRaw(context.Background(), tfPlanPath)
		if err != nil {
			log.Fatalf("error running ShowPlanFileRaw: %s", err)
		}

		for _, s := range tfPlanOutClean {
			planOut = strings.TrimSuffix(strings.ReplaceAll(planOut, s, ""), "\n")
		}

		obj.Msg = fmt.Sprintf("%s", planOut)
	}

	c <- obj
}
