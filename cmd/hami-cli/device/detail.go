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

func newDetailCommand(kubeconfig, output *string) *cobra.Command {
	return &cobra.Command{
		Use:   "detail <uuid>",
		Short: "Show detailed information about a GPU device",
		Long:  "Display detailed information about a specific GPU device and its usage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uuid := args[0]

			// Create Kubernetes client
			clientset, err := cliclient.NewClient(*kubeconfig)
			if err != nil {
				return fmt.Errorf("failed to create kubernetes client: %w", err)
			}

			// Get device detail
			device, err := aggregator.GetDeviceDetail(clientset, uuid)
			if err != nil {
				return fmt.Errorf("failed to get device detail: %w", err)
			}

			// Format and output
			return formatter.FormatOutput(os.Stdout, *output, device, formatDeviceDetailTable)
		},
	}
}

// formatDeviceDetailTable formats device detail as a table.
func formatDeviceDetailTable(writer io.Writer, data any) error {
	device, ok := data.(*aggregator.DeviceSummary)
	if !ok {
		return fmt.Errorf("invalid data type for device detail")
	}

	// Device summary
	fmt.Fprintf(writer, "Device: %s\n\n", device.UUID)
	fmt.Fprintf(writer, "Node:    %s\n", device.NodeName)
	fmt.Fprintf(writer, "Vendor:  %s\n", device.Vendor)
	fmt.Fprintf(writer, "Type:    %s\n", device.Type)
	fmt.Fprintf(writer, "Mode:    %s\n", device.Mode)
	fmt.Fprintf(writer, "Index:   %d\n", device.Index)
	fmt.Fprintf(writer, "NUMA:    %d\n", device.Numa)
	fmt.Fprintf(writer, "Health:  %s\n\n", formatter.FormatBool(device.Health))

	fmt.Fprintf(writer, "Resources:\n")
	fmt.Fprintf(writer, "  Cores:  %d/%d (%s)\n",
		device.CoreUsed, device.CoreTotal,
		formatter.FormatPercentage(device.CoreUsed, device.CoreTotal))
	fmt.Fprintf(writer, "  Memory: %s/%s (%s)\n\n",
		formatter.FormatMemory(device.MemoryUsed),
		formatter.FormatMemory(device.MemoryTotal),
		formatter.FormatPercentage(device.MemoryUsed, device.MemoryTotal))

	// Pod list
	if len(device.Pods) > 0 {
		w := formatter.NewTableWriter(writer)
		fmt.Fprintln(w, "POD\tNAMESPACE\tMEMORY\tCORES")

		for _, pod := range device.Pods {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\n",
				pod.Name,
				pod.Namespace,
				formatter.FormatMemory(pod.Memory),
				pod.Cores,
			)
		}
		w.Flush()
	} else {
		fmt.Fprintln(writer, "No pods using this device")
	}

	return nil
}
