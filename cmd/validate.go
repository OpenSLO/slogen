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
	"github.com/OpenSLO/slogen/libs"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"strings"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [list of paths to openslo configs]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Example: "slogen validate service/search.yaml \nslogen validate ~/team-a/slo/ ~/team-b/slo ~/core/slo/login.yaml",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		color.Blue("validate called on %s", strings.Join(args, ", "))
		if len(args) < 1 {
			color.Red("no arguments provided")
			cmd.Help()
			return
		}

		ie, err := cmd.Flags().GetBool(FlagIgnoreErrorLong)

		if err != nil {
			color.Red("error parsing ignore flag", err)
		}

		if ie {
			color.HiCyan("ignoring errors")
		}

		sloMap := make(map[string]*libs.SLOMultiVerse)
		for _, path := range args {
			libs.ParseDir(path, ie, sloMap)
		}
	},
}

const FlagIgnoreErrorLong = "ignoreErrors"
const FlagIgnoreErrorShort = "i"

func init() {
	rootCmd.AddCommand(validateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// validateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	validateCmd.Flags().BoolP(FlagIgnoreErrorLong, FlagIgnoreErrorShort, false,
		"whether to continue validation even after encountering errors",
	)
}
