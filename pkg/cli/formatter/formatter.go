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

package formatter

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v2"
)

// OutputFormat specifies the output format for CLI commands.
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
	FormatYAML  OutputFormat = "yaml"
)

// FormatOutput formats the data according to the specified format.
func FormatOutput(writer io.Writer, format string, data any, tableFunc func(io.Writer, any) error) error {
	switch OutputFormat(format) {
	case FormatJSON:
		return formatJSON(writer, data)
	case FormatYAML:
		return formatYAML(writer, data)
	case FormatTable:
		if tableFunc != nil {
			return tableFunc(writer, data)
		}
		return fmt.Errorf("table formatter not provided")
	default:
		return fmt.Errorf("unsupported output format: %s (supported: table, json, yaml)", format)
	}
}

// formatJSON formats data as JSON.
func formatJSON(writer io.Writer, data any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// formatYAML formats data as YAML.
func formatYAML(writer io.Writer, data any) error {
	encoder := yaml.NewEncoder(writer)
	defer encoder.Close()
	return encoder.Encode(data)
}

// NewTableWriter creates a new tabwriter for table output.
func NewTableWriter(writer io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
}

// FormatPercentage formats a percentage value.
func FormatPercentage(used, total int32) string {
	if total == 0 {
		return "0%"
	}
	percentage := float64(used) / float64(total) * 100
	return fmt.Sprintf("%.0f%%", percentage)
}

// FormatMemory formats memory in GiB.
func FormatMemory(memMB int32) string {
	return fmt.Sprintf("%dMi", memMB)
}

// TruncateString truncates a string to the specified length.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// FormatBool formats a boolean as Yes/No.
func FormatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// FormatList formats a slice as comma-separated string.
func FormatList(items []string) string {
	return strings.Join(items, ", ")
}
