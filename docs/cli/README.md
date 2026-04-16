# HAMi CLI

HAMi CLI is a command-line tool for querying and managing GPU resources in a Kubernetes cluster managed by HAMi (Heterogeneous AI Computing Virtualization Middleware).

## Features

- **Node Management**: List and inspect GPU nodes
- **Device Management**: View GPU devices across the cluster
- **Multiple Output Formats**: Support for table, JSON, and YAML output
- **Kubernetes Integration**: Works with kubeconfig for cluster access

## Installation

### Build from Source

```bash
cd /path/to/HAMi
make hami-cli
```

The binary will be available at `bin/hami-cli`.

### Using Pre-built Binary

Download the latest release from the [releases page](https://github.com/Project-HAMi/HAMi/releases).

## Usage

### Basic Commands

```bash
# Show version
hami-cli version

# List all GPU nodes
hami-cli node list

# Show detailed information about a node
hami-cli node detail <node-name>

# List all GPU devices
hami-cli device list

# Show detailed information about a device
hami-cli device detail <uuid>
```

### Global Flags

- `--kubeconfig <path>`: Path to kubeconfig file (default: `$KUBECONFIG` or `~/.kube/config`)
- `-o, --output <format>`: Output format - `table` (default), `json`, or `yaml`

### Examples

#### List nodes in table format (default)

```bash
$ hami-cli node list
NAME           DEVICES  CORE (USED/TOTAL)    MEMORY (USED/TOTAL)      GPU PODS
gpu-node-1     4        0/400 (0%)           0Mi/40960Mi (0%)         0
gpu-node-2     4        200/400 (50%)        20480Mi/40960Mi (50%)    2
```

#### List nodes in JSON format

```bash
$ hami-cli node list -o json
[
  {
    "name": "gpu-node-1",
    "devices": [...],
    "totalCore": 400,
    "usedCore": 0,
    "totalMemory": 40960,
    "usedMemory": 0,
    "gpuPods": []
  }
]
```

#### Show node details

```bash
$ hami-cli node detail gpu-node-1
Node: gpu-node-1

Total Resources:
  Devices: 4
  Cores:   0/400 (0%)
  Memory:  0Mi/40960Mi (0%)

DEVICE        INDEX  TYPE     MODE       CORE     MEMORY              HEALTH  PODS
GPU-12345...  0      NVIDIA   hami-core  0/100    0Mi/10240Mi        Yes     0
GPU-67890...  1      NVIDIA   hami-core  0/100    0Mi/10240Mi        Yes     0
...
```

#### List all devices

```bash
$ hami-cli device list
UUID          INDEX  NODE         VENDOR  TYPE     MODE       CORE         MEMORY              HEALTH  PODS
GPU-12345...  0      gpu-node-1   NVIDIA  A100     hami-core  0/100 (0%)   0Mi/10240Mi (0%)   Yes     0
GPU-67890...  1      gpu-node-1   NVIDIA  A100     hami-core  50/100 (50%) 5120Mi/10240Mi (50%) Yes   1
```

#### Show device details

```bash
$ hami-cli device detail GPU-12345678-1234-1234-1234-123456789012
Device: GPU-12345678-1234-1234-1234-123456789012

Node:    gpu-node-1
Vendor:  NVIDIA
Type:    A100
Mode:    hami-core
Index:   0
NUMA:    0
Health:  Yes

Resources:
  Cores:  50/100 (50%)
  Memory: 5120Mi/10240Mi (50%)

POD                    NAMESPACE  MEMORY     CORES
tensorflow-training    default    5120Mi     50
```

#### Use custom kubeconfig

```bash
$ hami-cli node list --kubeconfig /path/to/kubeconfig
```

## How It Works

HAMi CLI queries the Kubernetes API directly to retrieve GPU information:

1. **Node Information**: Reads device registration annotations from node objects
2. **Device Allocation**: Parses pod annotations to determine GPU usage
3. **Data Aggregation**: Combines node and pod information to calculate utilization

The CLI does not require direct access to the HAMi scheduler API, making it lightweight and easy to use.

## Output Formats

### Table Format (Default)

Human-readable table output suitable for terminal display.

### JSON Format

Structured JSON output for programmatic consumption:

```bash
hami-cli node list -o json | jq '.[] | select(.usedCore > 0)'
```

### YAML Format

YAML output for configuration management:

```bash
hami-cli device list -o yaml > devices.yaml
```

## Architecture

The CLI is organized into the following components:

- **cmd/hami-cli**: Command-line interface entry points
  - `node/`: Node management commands
  - `device/`: Device management commands
- **pkg/cli**: Shared CLI utilities
  - `client/`: Kubernetes client initialization
  - `aggregator/`: Data aggregation logic
  - `formatter/`: Output formatting (table, JSON, YAML)

## Troubleshooting

### No GPU nodes found

This means no nodes in your cluster have HAMi device annotations. Ensure:
- HAMi device plugins are running on GPU nodes
- Nodes have the correct device registration annotations

### Unable to connect to cluster

Check your kubeconfig:
```bash
kubectl cluster-info
hami-cli node list --kubeconfig ~/.kube/config
```

### Permission denied

Ensure your kubeconfig has permissions to:
- List nodes
- List pods
- Read node and pod annotations

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](../../CONTRIBUTING.md) for details.

## License

HAMi CLI is licensed under the Apache License 2.0. See [LICENSE](../../LICENSE) for details.
