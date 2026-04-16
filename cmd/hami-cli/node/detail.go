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

func newDetailCommand(kubeconfig, output *string) *cobra.Command {
	return &cobra.Command{
		Use:   "detail <node-name>",
		Short: "Show detailed information about a GPU node",
		Long:  "Display detailed information about GPU devices and pods on a specific node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeName := args[0]

			// Create Kubernetes client
			clientset, err := cliclient.NewClient(*kubeconfig)
			if err != nil {
				return fmt.Errorf("failed to create kubernetes client: %w", err)
			}

			// Get node detail
			node, err := aggregator.GetNodeDetail(clientset, nodeName)
			if err != nil {
				return fmt.Errorf("failed to get node detail: %w", err)
			}

			// Format and output
			return formatter.FormatOutput(os.Stdout, *output, node, formatNodeDetailTable)
		},
	}
}

// formatNodeDetailTable formats node detail as a table.
func formatNodeDetailTable(writer io.Writer, data any) error {
	node, ok := data.(*aggregator.NodeSummary)
	if !ok {
		return fmt.Errorf("invalid data type for node detail")
	}

	// Node summary
	fmt.Fprintf(writer, "Node: %s\n\n", node.Name)
	fmt.Fprintf(writer, "Total Resources:\n")
	fmt.Fprintf(writer, "  Devices: %d\n", len(node.Devices))
	fmt.Fprintf(writer, "  Cores:   %d/%d (%s)\n",
		node.UsedCore, node.TotalCore,
		formatter.FormatPercentage(node.UsedCore, node.TotalCore))
	fmt.Fprintf(writer, "  Memory:  %s/%s (%s)\n\n",
		formatter.FormatMemory(node.UsedMemory),
		formatter.FormatMemory(node.TotalMemory),
		formatter.FormatPercentage(node.UsedMemory, node.TotalMemory))

	// Device details
	if len(node.Devices) > 0 {
		w := formatter.NewTableWriter(writer)
		fmt.Fprintln(w, "DEVICE\tINDEX\tTYPE\tMODE\tCORE\tMEMORY\tHEALTH\tPODS")

		for _, dev := range node.Devices {
			coreUsage := fmt.Sprintf("%d/%d", dev.CoreUsed, dev.CoreTotal)
			memUsage := fmt.Sprintf("%s/%s",
				formatter.FormatMemory(dev.MemoryUsed),
				formatter.FormatMemory(dev.MemoryTotal))

			fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\t%s\t%d\n",
				formatter.TruncateString(dev.UUID, 12),
				dev.Index,
				dev.Type,
				dev.Mode,
				coreUsage,
				memUsage,
				formatter.FormatBool(dev.Health),
				len(dev.Pods),
			)
		}
		w.Flush()
		fmt.Fprintln(writer)
	}

	// Pod list
	if len(node.GPUPods) > 0 {
		w := formatter.NewTableWriter(writer)
		fmt.Fprintln(w, "POD\tNAMESPACE\tDEVICE\tMEMORY\tCORES")

		for _, pod := range node.GPUPods {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n",
				pod.Name,
				pod.Namespace,
				formatter.TruncateString(pod.UUID, 12),
				formatter.FormatMemory(pod.Memory),
				pod.Cores,
			)
		}
		w.Flush()
	}

	return nil
}
