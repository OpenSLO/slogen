/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/SumoLogic-Incubator/slogen/libs"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"strings"

	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "slogen [paths to yaml config]...",
	Example: `slogen service/search.yaml 
slogen ~/team-a/slo/ ~/team-b/slo ~/core/slo/login.yaml
slogen ~/team-a/slo/ -o team-a/tf
`,
	Short: "generates terraform files from openslo compatible yaml configs",
	Long: `
Generates terraform files from openslo compatible yaml configs. 
Generated terraform files can be used to configure SLO monitors, scheduled views & dashboards in sumo.
One or more config or directory containing configs can be given as arg. Doesn't supports regex/wildcards as input.
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		color.Blue("gen called on %s", strings.Join(args, ", "))
		if len(args) < 1 {
			libs.BadInfo("\nno arguments provided\n")
			cmd.Help()
			return
		}

		c, err := GetGenConf(cmd)
		if err != nil {
			libs.BadInfo("\nerror parsing out flag : %s\n", err.Error())
			return
		}

		if c.IgnoreError {
			color.HiCyan("\nignoring errors\n")
		}

		var slos map[string]*libs.SLO
		for _, path := range args {
			slos, _ = libs.ParseDir(path, c.IgnoreError)
		}

		path, err := libs.GenTerraform(slos, *c)

		if err != nil {
			libs.BadResult("\nerror generating terraform for : %s\n", path)
			libs.BadInfo("%s\n\n", err)
			return
		}

		if c.DoPlan {
			err = libs.TFExec(c.OutDir, libs.TFPlan)
			if err != nil {
				libs.BadResult("\nerror planning terraform")
				libs.BadInfo("%s\n\n", err)
				return
			}
		}

		if c.DoApply {
			libs.TFExec(c.OutDir, libs.TFApply)
			if err != nil {
				libs.BadResult("\nerror applying terraform")
				libs.BadInfo("%s\n\n", err)
				return
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

const FlagOutDirLong = "out"
const FlagOutDirShort = "o"
const FlagPlanLong = "plan"
const FlagPlanShort = "p"
const FlagApplyLong = "apply"
const FlagApplyShort = "a"
const FlagCleanLong = "clean"
const FlagCleanShort = "c"
const FlagDashboardFolderLong = "dashboardFolder"
const FlagDashboardFolderShort = "d"
const FlagMonitorFolderLong = "monitorFolder"
const FlagMonitorFolderShort = "m"
const FlagViewPrefixLong = "viewPrefix"
const FlagViewPrefixShort = "v"
const FlagViewDestroy = "viewDestroy"
const FlagDestroy = "destroy"

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.slogen.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP(FlagOutDirLong, FlagOutDirShort, "tf",
		"output directory where to create the terraform files",
	)

	rootCmd.Flags().StringP(FlagDashboardFolderLong, FlagDashboardFolderShort, "slogen-tf-dashboards",
		"output directory where to create the terraform files",
	)

	rootCmd.Flags().StringP(FlagMonitorFolderLong, FlagMonitorFolderShort, "slogen-tf-monitors",
		"output directory where to create the terraform files",
	)

	//rootCmd.Flags().StringP(FlagViewPrefixLong, FlagViewPrefixShort, libs.ViewPrefix,
	//	"output directory where to create the terraform files",
	//)

	rootCmd.Flags().BoolP(FlagIgnoreErrorLong, FlagIgnoreErrorShort, false,
		"whether to continue validation even after encountering errors",
	)
	rootCmd.Flags().BoolP(FlagPlanLong, FlagPlanShort, false,
		"show plan output after generating the terraform config",
	)
	rootCmd.Flags().BoolP(FlagApplyLong, FlagApplyShort, false,
		"apply the generated terraform config as well",
	)
	//rootCmd.Flags().Bool(FlagViewDestroy, false,
	//	"whether to destroy old view on change of attributes like query, start_time & parsing mode",
	//)
	rootCmd.Flags().BoolP(FlagCleanLong, FlagCleanShort, false,
		"clean the old tf files for which openslo config were not found in the path args",
	)
	rootCmd.Flags().SortFlags = false
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".slogen" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".slogen")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func GetGenConf(cmd *cobra.Command) (*libs.GenConf, error) {
	ie, err := cmd.Flags().GetBool(FlagIgnoreErrorLong)
	if err != nil {
		return nil, err
	}

	outDir, err := cmd.Flags().GetString(FlagOutDirLong)
	if err != nil {
		return nil, err
	}

	clean, err := cmd.Flags().GetBool(FlagCleanLong)
	if err != nil {
		return nil, err
	}

	dashDir, err := cmd.Flags().GetString(FlagDashboardFolderLong)
	if err != nil {
		return nil, err
	}

	monDir, err := cmd.Flags().GetString(FlagMonitorFolderLong)
	if err != nil {
		return nil, err
	}

	doPlan, err := cmd.Flags().GetBool(FlagPlanLong)
	if err != nil {
		return nil, err
	}

	doApply, err := cmd.Flags().GetBool(FlagApplyLong)
	if err != nil {
		return nil, err
	}

	conf := &libs.GenConf{
		IgnoreError:   ie,
		OutDir:        outDir,
		Clean:         clean,
		DashFolder:    dashDir,
		MonitorFolder: monDir,
		DoPlan:        doPlan,
		DoApply:       doApply,
	}

	return conf, nil
}
