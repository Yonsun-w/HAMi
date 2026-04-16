# HAMi CLI Implementation Details

## Overview

This document describes the implementation details of the HAMi CLI tool, which provides command-line access to GPU resource information in HAMi-managed Kubernetes clusters.

## Architecture

### Component Structure

```
cmd/hami-cli/                      # CLI entry points
├── main.go                        # Root command
├── node/                          # Node commands
│   ├── node.go                   # Node command group
│   ├── list.go                   # List nodes
│   └── detail.go                 # Node detail
└── device/                        # Device commands
    ├── device.go                 # Device command group
    ├── list.go                   # List devices
    └── detail.go                 # Device detail

pkg/cli/                           # Shared utilities
├── client/                        # Kubernetes client
│   └── client.go                 # Client initialization
├── aggregator/                    # Data aggregation
│   ├── node.go                   # Node data aggregation
│   └── device.go                 # Device data aggregation
└── formatter/                     # Output formatting
    └── formatter.go              # Table/JSON/YAML formatters
```

### Data Flow

1. **User Command** → CLI parses flags and arguments
2. **Client Initialization** → Creates Kubernetes clientset
3. **Data Aggregation** → Queries K8s API for nodes and pods
4. **Data Processing** → Parses annotations and calculates usage
5. **Output Formatting** → Formats data as table/JSON/YAML
6. **Display** → Outputs to stdout

## Key Design Decisions

### 1. Direct Kubernetes API Access

**Decision**: Query Kubernetes API directly instead of HAMi scheduler API.

**Rationale**:
- HAMi scheduler doesn't expose a query API
- All GPU data is stored in K8s annotations
- Makes CLI lightweight and independent
- Works with any kubeconfig

**Implementation**:
```go
// Read device info from node annotations
deviceAnno := node.Annotations["hami.io/node-nvidia-register"]
devices, _ := device.DecodeNodeDevices(deviceAnno)

// Read pod assignments from pod annotations
nodeName := pod.Annotations["hami.io/vgpu-node"]
allocation := pod.Annotations["hami.io/vgpu-devices-allocated"]
```

### 2. Annotation Format

HAMi stores GPU information in Kubernetes annotations:

#### Node Device Registration
```
hami.io/node-nvidia-register: "UUID,count,memory,cores,type,numa,health,index,mode:..."
```

Example:
```
GPU-12345678,1,10240,100,NVIDIA,0,true,0,hami-core:GPU-87654321,1,10240,100,NVIDIA,0,true,1,hami-core:
```

#### Pod Device Allocation
```
hami.io/vgpu-devices-allocated: "UUID,memory,cores:UUID,memory,cores;..."
```

Example:
```
GPU-12345678,5120,50:GPU-87654321,2560,25
```

- `:` separates devices in same container
- `;` separates containers in same pod

### 3. Client-Side Aggregation

**Decision**: Aggregate data in CLI rather than server-side.

**Benefits**:
- No scheduler modifications needed
- Flexible filtering and sorting
- Easy to add new features
- Works with read-only K8s access

**Trade-offs**:
- More network requests
- Higher latency for large clusters
- Client does the computation

### 4. Output Format Design

**Table Format** (default):
- Human-readable
- Terminal-friendly
- Uses tabwriter for alignment

**JSON Format**:
- Programmatic access
- Pipe to jq for filtering
- Full structured data

**YAML Format**:
- Configuration files
- Easier to read than JSON
- K8s native format

## Code Reuse

The CLI reuses existing HAMi code:

### Device Parsing
```go
import "github.com/Project-HAMi/HAMi/pkg/device"

devices, err := device.DecodeNodeDevices(annotation)
```

Reuses: `pkg/device/devices.go:DecodeNodeDevices()`

### Client Initialization
```go
import "github.com/Project-HAMi/HAMi/pkg/util/client"

clientset, err := client.NewClient()
```

Pattern based on: `pkg/util/client/client.go`

### Annotation Keys
```go
import "github.com/Project-HAMi/HAMi/pkg/util"

nodeName := pod.Annotations[util.AssignedNodeAnnotations]
```

Reuses: `pkg/util/types.go:AssignedNodeAnnotations`

## Testing Strategy

### Unit Tests

**Formatter Tests** (`pkg/cli/formatter/formatter_test.go`):
- Test utility functions
- Verify output formats
- Check edge cases

**Aggregator Tests** (`pkg/cli/aggregator/node_test.go`):
- Use fake Kubernetes clientset
- Test data aggregation logic
- Verify parsing functions

### Integration Tests

Manual testing with real cluster:
```bash
# Setup test cluster with HAMi
minikube start --driver=docker
kubectl apply -f examples/nvidia-vgpu-scheduler.yaml

# Test CLI commands
./bin/hami-cli node list
./bin/hami-cli device list
./bin/hami-cli node detail <node-name>
```

### Test Coverage

Current coverage:
- Formatter: ~90% (all utility functions)
- Aggregator: ~75% (core logic)
- Commands: Manual testing

To run tests:
```bash
go test ./pkg/cli/... -v
go test ./pkg/cli/... -cover
```

## Extension Points

### Adding New Commands

1. Create command directory:
   ```bash
   mkdir cmd/hami-cli/quota
   ```

2. Implement command:
   ```go
   // quota.go
   func NewQuotaCommand(kubeconfig, output *string) *cobra.Command {
       cmd := &cobra.Command{
           Use:   "quota",
           Short: "Manage GPU quotas",
       }
       cmd.AddCommand(newListCommand(kubeconfig, output))
       return cmd
   }
   ```

3. Add to root command:
   ```go
   // main.go
   rootCmd.AddCommand(quota.NewQuotaCommand(&kubeconfig, &output))
   ```

### Adding New Output Formats

1. Implement formatter:
   ```go
   // pkg/cli/formatter/csv.go
   func formatCSV(writer io.Writer, data interface{}) error {
       // CSV formatting logic
   }
   ```

2. Add to format switch:
   ```go
   case "csv":
       return formatCSV(writer, data)
   ```

### Supporting New Vendors

Add vendor annotation mapping:
```go
// pkg/cli/aggregator/node.go
func getDeviceAnnotationKeys() map[string]string {
    return map[string]string{
        "NVIDIA":    "hami.io/node-nvidia-register",
        "AMD":       "hami.io/node-amd-register",
        "NewVendor": "hami.io/node-newvendor-register",
    }
}
```

## Performance Considerations

### Scalability

For large clusters:
- **Nodes**: Linear with node count (1 API call)
- **Pods**: Linear with pod count (1 API call)
- **Processing**: O(n*m) where n=nodes, m=pods

### Optimization Strategies

1. **Field Selectors**: Filter pods by node
   ```go
   selector := fmt.Sprintf("spec.nodeName=%s", node)
   pods, _ := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
       FieldSelector: selector,
   })
   ```

2. **Label Selectors**: Filter GPU nodes
   ```go
   selector := "hami.io/gpu-node=true"
   nodes, _ := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{
       LabelSelector: selector,
   })
   ```

3. **Caching**: Cache results for repeated queries
   ```go
   // Future: Add --cache flag
   ```

## Security Considerations

### RBAC Requirements

Minimum permissions needed:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: hami-cli-reader
rules:
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
```

### Kubeconfig Handling

- Supports `--kubeconfig` flag
- Falls back to `KUBECONFIG` env
- Uses `~/.kube/config` by default
- Supports in-cluster config

### Data Privacy

- CLI only reads (no writes)
- No sensitive data logged
- Respects namespace boundaries
- Works with namespace-scoped tokens

## Future Enhancements

### Phase 2 Features

1. **Quota Management**
   - List ResourceQuotas
   - Show quota usage
   - Compare limits vs usage

2. **Filtering**
   ```bash
   hami-cli device list --vendor nvidia --health true
   hami-cli node list --min-available 50%
   ```

3. **Watch Mode**
   ```bash
   hami-cli node list --watch
   ```

4. **Export**
   ```bash
   hami-cli node list --export > inventory.yaml
   ```

### Phase 3 Features

1. **Statistics**
   - Cluster-wide utilization
   - Vendor breakdown
   - Historical trends

2. **Interactive TUI**
   - Keyboard navigation
   - Real-time updates
   - Drill-down views

3. **Shell Completion**
   - Bash/Zsh completions
   - Node/device name completion

## Contributing

To contribute to HAMi CLI:

1. Follow existing code patterns
2. Add unit tests for new features
3. Update documentation
4. Ensure backward compatibility
5. Run tests: `go test ./pkg/cli/... -v`

## References

- [HAMi Architecture](../architecture.md)
- [Kubernetes API](https://kubernetes.io/docs/reference/kubernetes-api/)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [HAMi Issue #1638](https://github.com/Project-HAMi/HAMi/issues/1638)
