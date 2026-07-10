# Installation

## Requirements

- Go 1.18+ (for generics support)
- Git

## Install

```bash
go get github.com/chuccp/go-web-frame
```

## Verify Installation

Create a simple test file:

```go
package main

import (
    wf "github.com/chuccp/go-web-frame"
)

func main() {
    _ = wf.NewBuilder(nil)
}
```

Run:

```bash
go run main.go
```

## Dependencies

Go Web Frame automatically installs the following dependencies:

- Gin (HTTP framework)
- GORM (ORM)
- Viper (configuration)
- Zap (logging)
- Sonic (JSON)

## Next Steps

- [Quick Start](quick-start.md) - Create your first application