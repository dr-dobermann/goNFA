# Package registry

The `registry` package provides a thread-safe mapping from string names to Guard and Action objects. This enables decoupling of declarative definitions (from YAML files) and actual implementation code.

## Overview

The Registry allows:
- Registration of Guards and Actions by string name
- Thread-safe concurrent access
- YAML file loading support
- Runtime object resolution

## Usage

### Basic Usage

```go
registry := registry.New()

// Register guards and actions
err := registry.RegisterGuard("isManager", &ManagerGuard{})
err = registry.RegisterAction("sendEmail", &EmailAction{})

// Use with YAML loading
definition, err := definition.LoadDefinition(yamlReader, registry)
```

### Complete Example

```go
registry := registry.New()

// Register all required objects
registry.RegisterGuard("isManager", &ManagerGuard{})
registry.RegisterGuard("hasPermission", &PermissionGuard{})
registry.RegisterAction("sendNotification", &NotificationAction{})
registry.RegisterAction("updateDatabase", &DatabaseAction{})
registry.RegisterAction("logEvent", &LogAction{})

// Load definition from YAML
file, err := os.Open("workflow.yaml")
if err != nil {
    return err
}
defer file.Close()

definition, err := definition.LoadDefinition(file, registry)
```

## API Reference

See [GoDoc](https://pkg.go.dev/github.com/dr-dobermann/gonfa/pkg/registry) for complete API documentation.
