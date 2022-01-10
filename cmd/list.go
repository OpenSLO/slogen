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
	"github.com/SumoLogic-Labs/slogen/libs"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "utility command to get additional info about your sumo resources e.g. ",
	Long: `
The OpenSLO config might need some additional info about your sumo resources that can be difficult to collect from the UI.
This command helps you list them down, for e.g. list of monitors with their ID to specify in the alert notifications.
`,
	Run: func(cmd *cobra.Command, args []string) {
		conns, err := libs.GiveConnectionIDS("")

		if err != nil {
			libs.BadUResult("\n%s\n", err.Error())
			return
		}

		for _, conn := range conns {
			fmt.Printf("\n ID : ")
			libs.GoodInfo("%s\t", conn.ID)
			fmt.Printf("\tType : ")
			libs.OkInfo("%16s", conn.Type)
			fmt.Printf("\t Name : ")
			libs.OkInfo("%s\n", conn.Name)
		}

	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	listCmd.Flags().BoolP("connections", "c", true, "list all monitor connections")
}
