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
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		//libs.BadInfo("Resource that will be destroyed\n\n")
		outDir, err := cmd.Flags().GetString(FlagOutDirLong)
		if err != nil {
			libs.BadResult("Unable to run destroy \nerror : %s\n", err.Error())
			return
		}

		err = libs.TFExec(outDir, libs.TFPlanDestroy)

		if err != nil {
			libs.BadResult("\nunable to run destroy : %s\n", err.Error())
			return
		}

		fmt.Printf("\n\n")

		if destroyPrompt(outDir) {
			libs.TFExec(outDir, libs.TFDestroy)
		}

	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().StringP(FlagOutDirLong, FlagOutDirShort, "tf",
		"output directory where to apply step was executed",
	)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// destroyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// destroyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

const destroyYesOption = "Yes, run destroy."
const destroyNoOption = "No, cancel it."

func destroyPrompt(outDir string) bool {
	prompt := promptui.Select{
		Label: "Are you sure to delete the dashboards, views and monitor generated at : " + outDir,
		Items: []string{destroyNoOption, destroyYesOption},
	}

	_, result, err := prompt.Run()

	if err != nil {
		libs.BadInfo("Prompt failed %v\n", err)
		return false
	}

	if result == destroyYesOption {
		return true
	}

	return false
}
