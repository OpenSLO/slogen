package libs

import (
	"context"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/hashicorp/terraform-exec/tfinstall"
	"io/ioutil"
	"os"
)

const (
	EnvKeySumoAccessID    = "SUMOLOGIC_ACCESSID"
	EnvKeySumoAccessKey   = "SUMOLOGIC_ACCESSKEY"
	EnvKeySumoEnvironment = "SUMOLOGIC_ENVIRONMENT"
)

type TFAction string

const (
	TFPlan        TFAction = "PLAN"
	TFApply       TFAction = "APPLY"
	TFPlanDestroy TFAction = "PlanDestroy"
	TFDestroy     TFAction = "DESTROY"
)

func TFExec(wdPath string, action TFAction) error {
	tmpDir, err := ioutil.TempDir("", "tfinstall")
	if err != nil {
		log.Errorf("error creating temp dir: %s", err)
		return err
	}
	defer os.RemoveAll(tmpDir)

	execPath, err := tfinstall.Find(context.Background(), tfinstall.LatestVersion(tmpDir, false))
	if err != nil {
		log.Errorf("error locating Terraform binary: %s", err)
		return err
	}

	tf, err := tfexec.NewTerraform(wdPath, execPath)
	env := map[string]string{
		EnvKeySumoAccessID:    os.Getenv(EnvKeySumoAccessID),
		EnvKeySumoAccessKey:   os.Getenv(EnvKeySumoAccessKey),
		EnvKeySumoEnvironment: os.Getenv(EnvKeySumoEnvironment),
	}

	err = tf.SetEnv(env)

	if err != nil {
		log.Error(err)
		return err
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Errorf("error running Init: %s", err)
		return err
	}

	//_, err = tf.Show(context.Background())
	//if err != nil {
	//	log.Errorf("error running Show: %s", err)
	//	return err
	//}
	tf.SetStdout(os.Stdout)
	tf.SetLogger(InfoLogger{})
	tf.SetStderr(os.Stderr)
	if err != nil {
		log.Errorf("error running NewTerraform: %s", err)
		return err
	}

	var ok bool
	if action == TFPlan {
		ok, err = tf.Plan(context.Background())

		if !ok && err == nil {
			log.Info("no changes found in tf plan")
		}
	}

	if action == TFApply {
		err = tf.Apply(context.Background())
		if err != nil {
			log.Error(err)
			return err
		}
	}

	if action == TFPlanDestroy {
		WarnUInfo("\nresource that will be destroyed\n\n")
		ok, err = tf.Plan(context.Background(), tfexec.Destroy(true))
		if err != nil {
			log.Error(err)
			return err
		}
	}

	if action == TFDestroy {
		err = tf.Destroy(context.Background())
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

const LookupTableIdKey = "lookupTableId"
