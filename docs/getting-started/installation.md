# Installation

## Requirements

- Go 1.22 or higher
- A Globus account (create one at [globus.org](https://www.globus.org))

## Install via go get

Add the SDK to your project:

```bash
go get github.com/scttfrdmn/globus-go-sdk/v3
```

## Version Selection

The SDK uses semantic versioning with v3 as the major version:

```bash
# Get latest v3.x release
go get github.com/scttfrdmn/globus-go-sdk/v3

# Get specific version
go get github.com/scttfrdmn/globus-go-sdk/v3@v3.65.0-1
```

## Import Path

In your Go code:

```go
import (
    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
)
```

## Verify Installation

Create a simple test program:

```go
package main

import (
    "fmt"
    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
)

func main() {
    fmt.Println("Globus Go SDK installed successfully!")
    fmt.Printf("SDK Version: %s\n", core.Version)
}
```

Run it:

```bash
go run main.go
```

## go.mod Configuration

Your `go.mod` should include:

```
module your-project

go 1.22

require (
    github.com/scttfrdmn/globus-go-sdk/v3 v3.65.0-1
)
```

## Update to Latest Version

Update to the latest version:

```bash
go get -u github.com/scttfrdmn/globus-go-sdk/v3
```

## Vendoring

If you use vendoring:

```bash
go mod vendor
```

## Package Documentation

Complete API documentation is available at:

**[pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v3](https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v3)**

## Available Packages

Main packages in the SDK:

### Core Packages

- `pkg/core` - Core SDK functionality and configuration
- `pkg/core/authorizers` - Authentication and authorization
- `pkg/core/errors` - Error types and handling
- `pkg/core/http` - HTTP client with connection pooling

### Service Packages

- `pkg/services/auth` - Authentication service
- `pkg/services/transfer` - Transfer service
- `pkg/services/search` - Search service
- `pkg/services/groups` - Groups service
- `pkg/services/flows` - Flows service
- `pkg/services/timers` - Timers service
- `pkg/services/compute` - Compute service

## Dependencies

The SDK has minimal dependencies:

- Standard library packages
- No heavy third-party dependencies

View the complete dependency tree:

```bash
go mod graph
```

## Next Steps

Now that you have the SDK installed, proceed to the [Quick Start](quickstart.md) guide to write your first code.
