/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

const (
	calicoVersion = "v3.26.1"
	antreaVersion = "v1.14.0"
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
IF YOU WANT TO MIGRATE YOUR SERVICE, PLEASE CHECK THE REQUIREMENTS TABLE.`)
		printRequirementsTable()
	},
}

func printRequirementsTable() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.Style().Options.DrawBorder = true
	t.Style().Options.SeparateRows = true
	t.Style().Format = table.FormatOptions{
		Footer: text.FormatLower,
		Header: text.FormatLower,
		Row:    text.FormatLower,
	}
	// set max width of the table

	t.SetTitle("Requirements Calico Configuration")
	t.AppendHeader(table.Row{"Calico Version", calicoVersion})
	// t.AppendHeader(table.Row{"Cluster IPAM Plugin", "Calico-ipam"})
	t.AppendRow(table.Row{"Cluster IPAM Plugin", "Calico-ipam"})
	t.AppendRow(table.Row{"ipip mode", "always"})
	t.AppendRow(table.Row{"natOutgoing", "true"})
	t.AppendRow(table.Row{"alloweduses", "workload, tunnel"})
	t.AppendRow(table.Row{"ipamconfig.autoAllocateBlocks", "true"})
	t.Render()
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
