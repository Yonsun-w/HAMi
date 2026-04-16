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
	"bytes"
	"strings"
	"testing"
)

func TestFormatPercentage(t *testing.T) {
	tests := []struct {
		name     string
		used     int32
		total    int32
		expected string
	}{
		{"zero total", 0, 0, "0%"},
		{"zero used", 0, 100, "0%"},
		{"half used", 50, 100, "50%"},
		{"full used", 100, 100, "100%"},
		{"over 100", 150, 100, "150%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPercentage(tt.used, tt.total)
			if result != tt.expected {
				t.Errorf("FormatPercentage(%d, %d) = %s, want %s", tt.used, tt.total, result, tt.expected)
			}
		})
	}
}

func TestFormatMemory(t *testing.T) {
	tests := []struct {
		name     string
		memMB    int32
		expected string
	}{
		{"zero", 0, "0Mi"},
		{"small", 512, "512Mi"},
		{"large", 10240, "10240Mi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatMemory(tt.memMB)
			if result != tt.expected {
				t.Errorf("FormatMemory(%d) = %s, want %s", tt.memMB, result, tt.expected)
			}
		})
	}
}

func TestFormatBool(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		expected string
	}{
		{"true", true, "Yes"},
		{"false", false, "No"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBool(tt.value)
			if result != tt.expected {
				t.Errorf("FormatBool(%v) = %s, want %s", tt.value, result, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"very short max", "hello", 4, "h..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateString(%s, %d) = %s, want %s", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	data := map[string]any{
		"name":  "test",
		"count": 42,
	}

	var buf bytes.Buffer
	err := formatJSON(&buf, data)
	if err != nil {
		t.Fatalf("formatJSON() error = %v", err)
	}

	result := buf.String()
	if !strings.Contains(result, `"name": "test"`) {
		t.Errorf("formatJSON() output doesn't contain expected JSON")
	}
	if !strings.Contains(result, `"count": 42`) {
		t.Errorf("formatJSON() output doesn't contain expected JSON")
	}
}

func TestFormatYAML(t *testing.T) {
	data := map[string]any{
		"name":  "test",
		"count": 42,
	}

	var buf bytes.Buffer
	err := formatYAML(&buf, data)
	if err != nil {
		t.Fatalf("formatYAML() error = %v", err)
	}

	result := buf.String()
	if !strings.Contains(result, "name: test") {
		t.Errorf("formatYAML() output doesn't contain expected YAML")
	}
	if !strings.Contains(result, "count: 42") {
		t.Errorf("formatYAML() output doesn't contain expected YAML")
	}
}

func TestFormatList(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected string
	}{
		{"empty", []string{}, ""},
		{"single", []string{"one"}, "one"},
		{"multiple", []string{"one", "two", "three"}, "one, two, three"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatList(tt.items)
			if result != tt.expected {
				t.Errorf("FormatList(%v) = %s, want %s", tt.items, result, tt.expected)
			}
		})
	}
}
