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

	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new [path-to-create]",
	Short: "create a sample config from given profile",
	Example: `
slogen -n sumo-logs team-rockstar/important-service.yaml
`,
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("new called")
	},
}

const FlagProfileLong = "profile"
const FlagProfileShort = "p"

func init() {

	rootCmd.AddCommand(newCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	newCmd.Flags().StringP(FlagProfileLong, FlagProfileShort, "sumo-logs",
		"template profile for the slo config, allowed values : sumo-logs | sumo-metrics",
	)

}
