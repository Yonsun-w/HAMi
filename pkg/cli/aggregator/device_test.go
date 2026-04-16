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

func TestListDevices(t *testing.T) {
	// Initialize device support map
	device.SupportDevices = map[string]string{
		"NVIDIA": "hami.io/vgpu-devices-allocated",
	}

	// Create fake clientset with test data
	node1 := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node-1",
			Annotations: map[string]string{
				"hami.io/node-nvidia-register": "GPU-1111,1,10240,100,NVIDIA,0,true,0,hami-core:",
			},
		},
	}

	node2 := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node-2",
			Annotations: map[string]string{
				"hami.io/node-nvidia-register": "GPU-2222,1,10240,100,NVIDIA,0,true,0,hami-core:",
			},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Annotations: map[string]string{
				util.AssignedNodeAnnotations:     "test-node-1",
				"hami.io/vgpu-devices-allocated": "GPU-1111,5120,50",
			},
		},
	}

	clientset := fake.NewSimpleClientset(node1, node2, pod)

	// Test ListDevices
	devices, err := ListDevices(clientset)
	if err != nil {
		t.Fatalf("ListDevices() error = %v", err)
	}

	if len(devices) != 2 {
		t.Fatalf("Expected 2 devices, got %d", len(devices))
	}

	// Check first device (should have usage)
	found := false
	for _, dev := range devices {
		if dev.UUID == "GPU-1111" {
			found = true
			if dev.MemoryUsed != 5120 {
				t.Errorf("Expected memory used 5120, got %d", dev.MemoryUsed)
			}
			if dev.CoreUsed != 50 {
				t.Errorf("Expected core used 50, got %d", dev.CoreUsed)
			}
			if len(dev.Pods) != 1 {
				t.Errorf("Expected 1 pod, got %d", len(dev.Pods))
			}
		}
	}

	if !found {
		t.Error("Device GPU-1111 not found in results")
	}
}

func TestListDevicesEmpty(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	devices, err := ListDevices(clientset)
	if err != nil {
		t.Fatalf("ListDevices() error = %v", err)
	}

	if len(devices) != 0 {
		t.Errorf("Expected 0 devices for empty cluster, got %d", len(devices))
	}
}

func TestGetDeviceDetail(t *testing.T) {
	// Initialize device support map
	device.SupportDevices = map[string]string{
		"NVIDIA": "hami.io/vgpu-devices-allocated",
	}

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			Annotations: map[string]string{
				"hami.io/node-nvidia-register": "GPU-test-uuid,1,10240,100,NVIDIA,0,true,0,hami-core:",
			},
		},
	}

	clientset := fake.NewSimpleClientset(node)

	// Test GetDeviceDetail - found
	detail, err := GetDeviceDetail(clientset, "GPU-test-uuid")
	if err != nil {
		t.Fatalf("GetDeviceDetail() error = %v", err)
	}

	if detail.UUID != "GPU-test-uuid" {
		t.Errorf("Expected UUID 'GPU-test-uuid', got '%s'", detail.UUID)
	}

	// Test GetDeviceDetail - not found
	_, err = GetDeviceDetail(clientset, "nonexistent-uuid")
	if err == nil {
		t.Error("Expected error for nonexistent device, got nil")
	}
}
