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

package aggregator

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/Project-HAMi/HAMi/pkg/device"
	"github.com/Project-HAMi/HAMi/pkg/util"
)

func TestListNodes(t *testing.T) {
	// Initialize device support map
	device.SupportDevices = map[string]string{
		"NVIDIA": "hami.io/vgpu-devices-allocated",
	}

	// Create fake clientset with test data
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			Annotations: map[string]string{
				"hami.io/node-nvidia-register": "GPU-1234,1,10240,100,NVIDIA,0,true,0,hami-core:GPU-5678,1,10240,100,NVIDIA,0,true,1,hami-core:",
			},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Annotations: map[string]string{
				util.AssignedNodeAnnotations:     "test-node",
				"hami.io/vgpu-devices-allocated": "GPU-1234,5120,50",
			},
		},
	}

	clientset := fake.NewSimpleClientset(node, pod)

	// Test ListNodes
	nodes, err := ListNodes(clientset)
	if err != nil {
		t.Fatalf("ListNodes() error = %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(nodes))
	}

	nodeSum := nodes[0]
	if nodeSum.Name != "test-node" {
		t.Errorf("Expected node name 'test-node', got '%s'", nodeSum.Name)
	}

	if len(nodeSum.Devices) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(nodeSum.Devices))
	}

	if nodeSum.TotalCore != 200 {
		t.Errorf("Expected total core 200, got %d", nodeSum.TotalCore)
	}

	if nodeSum.TotalMemory != 20480 {
		t.Errorf("Expected total memory 20480, got %d", nodeSum.TotalMemory)
	}

	// Check device usage
	device0 := nodeSum.Devices[0]
	if device0.UUID != "GPU-1234" {
		t.Errorf("Expected device UUID 'GPU-1234', got '%s'", device0.UUID)
	}

	if device0.MemoryUsed != 5120 {
		t.Errorf("Expected device memory used 5120, got %d", device0.MemoryUsed)
	}

	if device0.CoreUsed != 50 {
		t.Errorf("Expected device core used 50, got %d", device0.CoreUsed)
	}

	if len(device0.Pods) != 1 {
		t.Errorf("Expected 1 pod on device, got %d", len(device0.Pods))
	}
}

func TestGetNodeDetail(t *testing.T) {
	// Initialize device support map
	device.SupportDevices = map[string]string{
		"NVIDIA": "hami.io/vgpu-devices-allocated",
	}

	// Create fake clientset with test data
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			Annotations: map[string]string{
				"hami.io/node-nvidia-register": "GPU-1234,1,10240,100,NVIDIA,0,true,0,hami-core:",
			},
		},
	}

	clientset := fake.NewSimpleClientset(node)

	// Test GetNodeDetail
	nodeDetail, err := GetNodeDetail(clientset, "test-node")
	if err != nil {
		t.Fatalf("GetNodeDetail() error = %v", err)
	}

	if nodeDetail.Name != "test-node" {
		t.Errorf("Expected node name 'test-node', got '%s'", nodeDetail.Name)
	}

	if len(nodeDetail.Devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len(nodeDetail.Devices))
	}
}

func TestGetNodeDetailNotFound(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	_, err := GetNodeDetail(clientset, "nonexistent-node")
	if err == nil {
		t.Error("Expected error for nonexistent node, got nil")
	}
}

func TestParsePodDeviceAllocation(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}

	tests := []struct {
		name       string
		allocation string
		vendor     string
		wantLen    int
		wantUUID   string
		wantMemory int32
		wantCores  int32
	}{
		{
			name:       "single device",
			allocation: "GPU-1234,5120,50",
			vendor:     "NVIDIA",
			wantLen:    1,
			wantUUID:   "GPU-1234",
			wantMemory: 5120,
			wantCores:  50,
		},
		{
			name:       "multiple devices per container",
			allocation: "GPU-1234,5120,50:GPU-5678,2560,25",
			vendor:     "NVIDIA",
			wantLen:    2,
			wantUUID:   "GPU-1234",
			wantMemory: 5120,
			wantCores:  50,
		},
		{
			name:       "empty allocation",
			allocation: "",
			vendor:     "NVIDIA",
			wantLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePodDeviceAllocation(pod, tt.allocation, tt.vendor)

			if len(result) != tt.wantLen {
				t.Errorf("Expected %d devices, got %d", tt.wantLen, len(result))
				return
			}

			if tt.wantLen > 0 {
				if result[0].UUID != tt.wantUUID {
					t.Errorf("Expected UUID '%s', got '%s'", tt.wantUUID, result[0].UUID)
				}
				if result[0].Memory != tt.wantMemory {
					t.Errorf("Expected memory %d, got %d", tt.wantMemory, result[0].Memory)
				}
				if result[0].Cores != tt.wantCores {
					t.Errorf("Expected cores %d, got %d", tt.wantCores, result[0].Cores)
				}
			}
		})
	}
}
