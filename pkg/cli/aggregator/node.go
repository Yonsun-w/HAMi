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
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/Project-HAMi/HAMi/pkg/device"
	"github.com/Project-HAMi/HAMi/pkg/util"
)

// NodeSummary contains aggregated information about a node's GPU resources.
type NodeSummary struct {
	Name        string          `json:"name"`
	Devices     []DeviceSummary `json:"devices"`
	TotalCore   int32           `json:"totalCore"`
	UsedCore    int32           `json:"usedCore"`
	TotalMemory int32           `json:"totalMemory"`
	UsedMemory  int32           `json:"usedMemory"`
	GPUPods     []PodGPUInfo    `json:"gpuPods"`
}

// DeviceSummary contains information about a single GPU device.
type DeviceSummary struct {
	UUID        string       `json:"uuid"`
	Index       uint         `json:"index"`
	NodeName    string       `json:"nodeName"`
	Vendor      string       `json:"vendor"`
	Type        string       `json:"type"`
	Mode        string       `json:"mode"`
	CoreUsed    int32        `json:"coreUsed"`
	CoreTotal   int32        `json:"coreTotal"`
	MemoryUsed  int32        `json:"memoryUsed"`
	MemoryTotal int32        `json:"memoryTotal"`
	Health      bool         `json:"health"`
	Numa        int          `json:"numa"`
	Pods        []PodGPUInfo `json:"pods"`
}

// PodGPUInfo contains information about a pod using GPU resources.
type PodGPUInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	UUID      string `json:"uuid,omitempty"`
	Memory    int32  `json:"memory,omitempty"`
	Cores     int32  `json:"cores,omitempty"`
}

// ListNodes returns a summary of all GPU nodes in the cluster.
func ListNodes(clientset kubernetes.Interface) ([]NodeSummary, error) {
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

	// Build pod assignment map
	podsByNode := make(map[string][]PodGPUInfo)
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
					podsByNode[nodeName] = append(podsByNode[nodeName], info)
					if info.UUID != "" {
						podsByDevice[info.UUID] = append(podsByDevice[info.UUID], info)
					}
				}
			}
		}
	}

	// Process each node
	var result []NodeSummary
	for _, node := range nodes.Items {
		summary, err := getNodeSummary(&node, podsByNode, podsByDevice)
		if err != nil {
			// Skip nodes without GPU devices
			continue
		}
		if summary != nil {
			result = append(result, *summary)
		}
	}

	return result, nil
}

// GetNodeDetail returns detailed information about a specific node.
func GetNodeDetail(clientset kubernetes.Interface, nodeName string) (*NodeSummary, error) {
	// Get the node
	node, err := clientset.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node %s: %w", nodeName, err)
	}

	// Get all pods with GPU assignments
	pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Build pod assignment map
	podsByNode := make(map[string][]PodGPUInfo)
	podsByDevice := make(map[string][]PodGPUInfo)

	for _, pod := range pods.Items {
		assignedNode, ok := pod.Annotations[util.AssignedNodeAnnotations]
		if !ok || assignedNode != nodeName {
			continue
		}

		// Parse device allocations
		for vendor, annoKey := range device.SupportDevices {
			if allocation, ok := pod.Annotations[annoKey]; ok && allocation != "" {
				podInfo := parsePodDeviceAllocation(&pod, allocation, vendor)
				for _, info := range podInfo {
					podsByNode[nodeName] = append(podsByNode[nodeName], info)
					if info.UUID != "" {
						podsByDevice[info.UUID] = append(podsByDevice[info.UUID], info)
					}
				}
			}
		}
	}

	return getNodeSummary(node, podsByNode, podsByDevice)
}

// getNodeSummary builds a NodeSummary from a node and pod assignments.
func getNodeSummary(node *corev1.Node, podsByNode map[string][]PodGPUInfo, podsByDevice map[string][]PodGPUInfo) (*NodeSummary, error) {
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
		return nil, fmt.Errorf("no GPU devices found on node")
	}

	// Decode device information
	devices, err := device.DecodeNodeDevices(deviceAnno)
	if err != nil {
		return nil, fmt.Errorf("failed to decode device info: %w", err)
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("no devices found")
	}

	// Build device summaries
	var deviceSummaries []DeviceSummary
	var totalCore, usedCore, totalMemory, usedMemory int32

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

		deviceSummaries = append(deviceSummaries, deviceSum)
		totalCore += dev.Devcore
		totalMemory += dev.Devmem
		usedCore += deviceSum.CoreUsed
		usedMemory += deviceSum.MemoryUsed
	}

	return &NodeSummary{
		Name:        node.Name,
		Devices:     deviceSummaries,
		TotalCore:   totalCore,
		UsedCore:    usedCore,
		TotalMemory: totalMemory,
		UsedMemory:  usedMemory,
		GPUPods:     podsByNode[node.Name],
	}, nil
}

// parsePodDeviceAllocation parses device allocation annotation.
func parsePodDeviceAllocation(pod *corev1.Pod, allocation string, vendor string) []PodGPUInfo {
	var result []PodGPUInfo

	// Parse allocation format: "uuid,memory,cores:uuid,memory,cores;..."
	// Split by semicolon for multiple containers
	//nolint:modernize // Using Split for consistency with existing codebase
	containers := strings.Split(allocation, device.OnePodMultiContainerSplitSymbol)

	for _, containerAlloc := range containers {
		// Split by colon for multiple devices per container
		//nolint:modernize // Using Split for consistency with existing codebase
		deviceAllocs := strings.Split(containerAlloc, device.OneContainerMultiDeviceSplitSymbol)

		for _, deviceAlloc := range deviceAllocs {
			if deviceAlloc == "" {
				continue
			}

			// Parse device allocation: "uuid,memory,cores"
			parts := strings.Split(deviceAlloc, ",")
			if len(parts) >= 3 {
				var memory, cores int32
				fmt.Sscanf(parts[1], "%d", &memory)
				fmt.Sscanf(parts[2], "%d", &cores)

				result = append(result, PodGPUInfo{
					Name:      pod.Name,
					Namespace: pod.Namespace,
					UUID:      parts[0],
					Memory:    memory,
					Cores:     cores,
				})
			}
		}
	}

	return result
}

// getDeviceAnnotationKeys returns annotation keys for device registration.
func getDeviceAnnotationKeys() map[string]string {
	// Map vendor name to node annotation key
	return map[string]string{
		"NVIDIA":    "hami.io/node-nvidia-register",
		"AMD":       "hami.io/node-amd-register",
		"Cambricon": "hami.io/node-cambricon-register",
		"Hygon":     "hami.io/node-hygon-register",
		"Iluvatar":  "hami.io/node-iluvatar-register",
		"Mthreads":  "hami.io/node-mthreads-register",
	}
}
