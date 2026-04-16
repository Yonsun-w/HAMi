/*
Copyright 2024 The HAMi Authors.

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

package node

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/Project-HAMi/HAMi/pkg/cli/aggregator"
	cliclient "github.com/Project-HAMi/HAMi/pkg/cli/client"
	"github.com/Project-HAMi/HAMi/pkg/cli/formatter"
)

func newListCommand(kubeconfig, output *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all GPU nodes",
		Long:  "Display a list of all nodes with GPU resources managed by HAMi",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create Kubernetes client
			clientset, err := cliclient.NewClient(*kubeconfig)
			if err != nil {
				return fmt.Errorf("failed to create kubernetes client: %w", err)
			}

			// Get node list
			nodes, err := aggregator.ListNodes(clientset)
			if err != nil {
				return fmt.Errorf("failed to list nodes: %w", err)
			}

			if len(nodes) == 0 {
				fmt.Fprintln(os.Stderr, "No GPU nodes found")
				return nil
			}

			// Format and output
			return formatter.FormatOutput(os.Stdout, *output, nodes, formatNodeListTable)
		},
	}
}

// formatNodeListTable formats node list as a table.
func formatNodeListTable(writer io.Writer, data any) error {
	nodes, ok := data.([]aggregator.NodeSummary)
	if !ok {
		return fmt.Errorf("invalid data type for node list")
	}

	w := formatter.NewTableWriter(writer)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "NAME\tDEVICES\tCORE (USED/TOTAL)\tMEMORY (USED/TOTAL)\tGPU PODS")

	// Rows
	for _, node := range nodes {
		coreUsage := fmt.Sprintf("%d/%d (%s)",
			node.UsedCore, node.TotalCore,
			formatter.FormatPercentage(node.UsedCore, node.TotalCore))

		memUsage := fmt.Sprintf("%s/%s (%s)",
			formatter.FormatMemory(node.UsedMemory),
			formatter.FormatMemory(node.TotalMemory),
			formatter.FormatPercentage(node.UsedMemory, node.TotalMemory))

		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%d\n",
			node.Name,
			len(node.Devices),
			coreUsage,
			memUsage,
			len(node.GPUPods),
		)
	}

	return nil
}
