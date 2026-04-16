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
	"github.com/spf13/cobra"
)

// NewNodeCommand creates the node command.
func NewNodeCommand(kubeconfig, output *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Manage GPU nodes",
		Long:  "Query and display information about GPU nodes in the cluster",
	}

	cmd.AddCommand(newListCommand(kubeconfig, output))
	cmd.AddCommand(newDetailCommand(kubeconfig, output))

	return cmd
}
