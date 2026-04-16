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

package device

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
		Short: "List all GPU devices",
		Long:  "Display a list of all GPU devices across the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create Kubernetes client
			clientset, err := cliclient.NewClient(*kubeconfig)
			if err != nil {
				return fmt.Errorf("failed to create kubernetes client: %w", err)
			}

			// Get device list
			devices, err := aggregator.ListDevices(clientset)
			if err != nil {
				return fmt.Errorf("failed to list devices: %w", err)
			}

			if len(devices) == 0 {
				fmt.Fprintln(os.Stderr, "No GPU devices found")
				return nil
			}

			// Format and output
			return formatter.FormatOutput(os.Stdout, *output, devices, formatDeviceListTable)
		},
	}
}

// formatDeviceListTable formats device list as a table.
func formatDeviceListTable(writer io.Writer, data any) error {
	devices, ok := data.([]aggregator.DeviceSummary)
	if !ok {
		return fmt.Errorf("invalid data type for device list")
	}

	w := formatter.NewTableWriter(writer)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "UUID\tINDEX\tNODE\tVENDOR\tTYPE\tMODE\tCORE\tMEMORY\tHEALTH\tPODS")

	// Rows
	for _, dev := range devices {
		coreUsage := fmt.Sprintf("%d/%d (%s)",
			dev.CoreUsed, dev.CoreTotal,
			formatter.FormatPercentage(dev.CoreUsed, dev.CoreTotal))

		memUsage := fmt.Sprintf("%s/%s (%s)",
			formatter.FormatMemory(dev.MemoryUsed),
			formatter.FormatMemory(dev.MemoryTotal),
			formatter.FormatPercentage(dev.MemoryUsed, dev.MemoryTotal))

		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\n",
			formatter.TruncateString(dev.UUID, 12),
			dev.Index,
			dev.NodeName,
			dev.Vendor,
			dev.Type,
			dev.Mode,
			coreUsage,
			memUsage,
			formatter.FormatBool(dev.Health),
			len(dev.Pods),
		)
	}

	return nil
}
