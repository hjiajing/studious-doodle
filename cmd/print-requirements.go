/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)


// printRequirementsCmd represents the printRequirements command
var printRequirementsCmd = &cobra.Command{
	Use:   "print-requirements",
	Short: "Prints the requirements of Calico for migration",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`========== WARNING ==========
THIS IS AN EXPERIMENTAL FEATURE.
YOUR SERVICE MAY NOT WORK AS EXPECTED DURING THE MIGRATION.
IF YOU WANT TO MIGRATE YOUR SERVICE, PLEASE MAKE SURE THE CALICO APISERVER IS INSTALLED.`)
	},
}

func init() {
	rootCmd.AddCommand(printRequirementsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// printRequirementsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// printRequirementsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
