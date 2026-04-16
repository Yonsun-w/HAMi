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
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/Project-HAMi/HAMi/pkg/device"
	"github.com/Project-HAMi/HAMi/pkg/util"
)

// ListDevices returns a list of all GPU devices across all nodes.
func ListDevices(clientset kubernetes.Interface) ([]DeviceSummary, error) {
	// List all nodes
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Get all pods with GPU assignments
	pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Build pod-to-device mapping
	podsByDevice := make(map[string][]PodGPUInfo)
	for _, pod := range pods.Items {
		nodeName, ok := pod.Annotations[util.AssignedNodeAnnotations]
		if !ok || nodeName == "" {
			continue
		}

		// Parse device allocations
		for vendor, annoKey := range device.SupportDevices {
			if allocation, ok := pod.Annotations[annoKey]; ok && allocation != "" {
				podInfo := parsePodDeviceAllocation(&pod, allocation, vendor)
				for _, info := range podInfo {
					if info.UUID != "" {
						podsByDevice[info.UUID] = append(podsByDevice[info.UUID], info)
					}
				}
			}
		}
	}

	// Collect all devices from all nodes
	var result []DeviceSummary
	for _, node := range nodes.Items {
		// Check for device annotations
		deviceAnno := ""
		vendor := ""
		for v, annoKey := range getDeviceAnnotationKeys() {
			if anno, ok := node.Annotations[annoKey]; ok && anno != "" {
				deviceAnno = anno
				vendor = v
				break
			}
		}

		if deviceAnno == "" {
			continue
		}

		// Decode device information
		devices, err := device.DecodeNodeDevices(deviceAnno)
		if err != nil {
			continue
		}

		// Add each device to result
		for _, dev := range devices {
			deviceSum := DeviceSummary{
				UUID:        dev.ID,
				Index:       dev.Index,
				NodeName:    node.Name,
				Vendor:      vendor,
				Type:        dev.Type,
				Mode:        dev.Mode,
				CoreTotal:   dev.Devcore,
				MemoryTotal: dev.Devmem,
				Health:      dev.Health,
				Numa:        dev.Numa,
				Pods:        podsByDevice[dev.ID],
			}

			// Calculate usage from pods
			for _, pod := range deviceSum.Pods {
				deviceSum.CoreUsed += pod.Cores
				deviceSum.MemoryUsed += pod.Memory
			}

			result = append(result, deviceSum)
		}
	}

	return result, nil
}

// GetDeviceDetail returns detailed information about a specific device.
func GetDeviceDetail(clientset kubernetes.Interface, uuid string) (*DeviceSummary, error) {
	devices, err := ListDevices(clientset)
	if err != nil {
		return nil, err
	}

	for _, dev := range devices {
		if dev.UUID == uuid {
			return &dev, nil
		}
	}

	return nil, fmt.Errorf("device not found: %s", uuid)
}
