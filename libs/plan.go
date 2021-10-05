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
	TFPlan  TFAction = "PLAN"
	TFApply TFAction = "APPLY"
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

	log.Infow("env used", "env", env)
	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)
	if err != nil {
		log.Errorf("error running NewTerraform: %s", err)
		return err
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Errorf("error running Init: %s", err)
		return err
	}

	_, err = tf.Show(context.Background())
	if err != nil {
		log.Errorf("error running Show: %s", err)
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

	//output, err := tf.Output(context.Background())
	//
	//if err != nil {
	//	log.Error(err)
	//	return err
	//}

	//log.Infof("tf action '%s' done", action)
	//log.Infow("output given", LookupTableIdKey, output[LookupTableIdKey].Value)

	return nil
}

const LookupTableIdKey = "lookupTableId"
