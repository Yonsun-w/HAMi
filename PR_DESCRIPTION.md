# Add hami-cli: Command-Line Tool for GPU Resource Management

## Overview

This PR adds `hami-cli`, a new command-line tool for querying and managing GPU resources in HAMi-managed Kubernetes clusters. This addresses issue #1638.

## AI Assistance Disclosure

> [!IMPORTANT]
> **This PR was developed with AI assistance (Claude Code).**
>
> The implementation was created using Claude Code for:
> - Code generation (all Go source files)
> - Test writing (unit tests)
> - Documentation (user and developer guides)
>
> All code has been reviewed, tested, and verified to:
> - Pass `make verify` (all linter checks)
> - Pass all unit tests (11 tests, >50% coverage)
> - Follow HAMi coding standards
> - Build successfully with `make hami-cli`

## Motivation

Currently, HAMi exposes GPU information primarily through:
- WebUI (visual interface)
- Prometheus metrics (monitoring)

Users need a **lightweight, scriptable CLI** for:
- Quick inspection of GPU resources
- CI/CD pipeline integration
- Automation scripts
- AI agent compatibility
- Debugging and troubleshooting

## Features

### Commands

#### Node Management
```bash
# List all GPU nodes
hami-cli node list

# Show detailed node information
hami-cli node detail <node-name>
```

#### Device Management
```bash
# List all GPU devices
hami-cli device list

# Show detailed device information
hami-cli device detail <device-uuid>
```

### Output Formats

- **Table** (default) - Human-readable terminal output
- **JSON** - Programmatic consumption, pipe to `jq`
- **YAML** - Configuration files

```bash
hami-cli node list -o json
hami-cli device list -o yaml
```

### Kubeconfig Support

- `--kubeconfig` flag
- `KUBECONFIG` environment variable
- Default path `~/.kube/config`
- In-cluster configuration

## Implementation Details

### Architecture

```
cmd/hami-cli/          # CLI entry points
  ├── main.go          # Root command
  ├── node/            # Node commands
  └── device/          # Device commands

pkg/cli/               # Shared utilities
  ├── client/          # Kubernetes client
  ├── aggregator/      # Data aggregation
  └── formatter/       # Output formatting
```

### Design Decisions

1. **Direct Kubernetes API Access**
   - Queries K8s API directly (no scheduler API changes)
   - Parses GPU info from node/pod annotations
   - Works with read-only K8s access

2. **Client-Side Aggregation**
   - Aggregates data in CLI
   - Flexible filtering and sorting
   - No scheduler modifications needed

3. **Code Reuse**
   - Reuses `pkg/device/devices.go:DecodeNodeDevices()`
   - Reuses `pkg/util/types.go` annotation constants
   - Reuses `pkg/version/version.go` version command
   - Follows existing patterns from `pkg/util/client`

## Testing

### Unit Tests
- ✅ Formatter tests: 7 tests, 69% coverage
- ✅ Aggregator tests: 4 tests, 53% coverage
- ✅ All tests passing

```bash
go test ./pkg/cli/... -v
```

### Verification
- ✅ `make verify` - All linter checks pass
- ✅ `make hami-cli` - Builds successfully
- ✅ Code follows project conventions

### Manual Testing
```bash
# Build
make hami-cli

# Test commands
./bin/hami-cli version
./bin/hami-cli --help
./bin/hami-cli node --help
./bin/hami-cli device --help
```

## Documentation

- **User Guide**: `docs/cli/README.md`
  - Installation instructions
  - Usage examples
  - Output format details
  - Troubleshooting

- **Implementation Guide**: `docs/cli/IMPLEMENTATION.md`
  - Architecture overview
  - Design decisions
  - Extension points
  - Testing strategy

## Files Changed

### New Files (17 files, ~2,500 lines)
- 7 command files (`cmd/hami-cli/`)
- 4 utility files (`pkg/cli/`)
- 2 test files
- 2 documentation files
- 1 summary document

### Modified Files
- `version.mk` - Added `hami-cli` to build targets

## Breaking Changes

None. This is a new tool with no impact on existing functionality.

## Future Enhancements

Potential improvements for future PRs:
- [ ] Quota management commands
- [ ] Watch mode (`--watch`)
- [ ] Filtering options (`--vendor`, `--health`)
- [ ] Export/import functionality
- [ ] Statistics and trends
- [ ] Shell completion (bash/zsh)

## Checklist

- [x] Code follows project style guidelines
- [x] All tests passing
- [x] `make verify` passes
- [x] Documentation added
- [x] Commit message follows conventions
- [x] AI assistance disclosed
- [ ] Tested on real HAMi cluster (needs cluster access)

## Related Issues

Closes #1638

## Screenshots

Example output:

```
$ hami-cli node list
NAME         DEVICES  CORE (USED/TOTAL)   MEMORY (USED/TOTAL)    GPU PODS
gpu-node-1   4        200/400 (50%)       20480Mi/40960Mi (50%)  2
gpu-node-2   4        0/400 (0%)          0Mi/40960Mi (0%)       0

$ hami-cli device list -o json
[
  {
    "uuid": "GPU-12345678-...",
    "index": 0,
    "nodeName": "gpu-node-1",
    "vendor": "NVIDIA",
    "type": "A100",
    "mode": "hami-core",
    "coreUsed": 50,
    "coreTotal": 100,
    "memoryUsed": 5120,
    "memoryTotal": 10240,
    "health": true,
    "pods": [...]
  }
]
```

## Request for Feedback

This is my first contribution to HAMi. I'd appreciate feedback on:
1. Code structure and organization
2. Test coverage (is 50%+ sufficient for initial PR?)
3. Documentation clarity
4. Any missing edge cases
5. Integration testing approach (I don't have a HAMi cluster for testing)

Thank you for reviewing!
