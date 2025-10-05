# Go Command Structure Design

## Project Overview

This document defines the command structure for the HashPost Go application using the cobra CLI framework. The goal is to create a clean, extensible command pattern with a `run` command for the server.

## Design Goals

### Primary Objectives
- **Clean CLI Interface**: Intuitive command structure
- **Extensibility**: Easy to add new commands
- **Server Management**: Dedicated `run` command for server operations
- **Configuration**: Command-line configuration options
- **Help System**: Comprehensive help and documentation

### Technical Goals
- Use cobra for CLI framework
- Implement command pattern for server operations
- Support configuration via flags and environment variables
- Provide clear error messages and help text

## Architecture Decisions

### Command Structure (Separate Binaries)
- **PDS Binary**: `hashpost-pds` - Personal Data Server for atproto protocol compliance
- **AppView Binary**: `hashpost-appview` - Stateless aggregator for data presentation
- **Shared Commands**: `hashpost migrate`, `hashpost generate` - Shared utilities
- **PDS Focus**: PDS handles all data and business logic
- **AppView Focus**: AppView is stateless aggregator
- **Independent Deployment**: Each component can be deployed separately

### Framework Choice
- **Cobra**: Industry standard for Go CLI applications
- **Viper**: Configuration management integration
- **Standard Library**: Use net/http for server implementation

## Command Implementation

### Shared Utilities Root Command
```go
// cmd/shared/root.go
var rootCmd = &cobra.Command{
    Use:   "hashpost",
    Short: "HashPost shared utilities",
    Long:  `HashPost shared utilities for database management and code generation.`,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### PDS Binary Implementation
```go
// cmd/pds/main.go
var rootCmd = &cobra.Command{
    Use:   "hashpost-pds",
    Short: "HashPost Personal Data Server",
    Long:  `HashPost PDS handles atproto protocol compliance and business logic.`,
}

var runCmd = &cobra.Command{
    Use:   "run",
    Short: "Start the PDS server",
    Long:  `Start the HashPost PDS server with the specified configuration.`,
    RunE:  runPDSServer,
}

func init() {
    rootCmd.AddCommand(runCmd)
    
    // PDS configuration flags
    runCmd.Flags().String("host", "localhost", "PDS host")
    runCmd.Flags().Int("port", 8080, "PDS port")
    runCmd.Flags().String("config", "", "Configuration file path")
    runCmd.Flags().Bool("dev", false, "Enable development mode")
}

func runPDSServer(cmd *cobra.Command, args []string) error {
    // Start PDS server with database access
    return nil
}
```

### AppView Binary Implementation
```go
// cmd/appview/main.go
var rootCmd = &cobra.Command{
    Use:   "hashpost-appview",
    Short: "HashPost AppView",
    Long:  `HashPost AppView aggregates data from PDS for presentation.`,
}

var runCmd = &cobra.Command{
    Use:   "run",
    Short: "Start the AppView server",
    Long:  `Start the HashPost AppView server with the specified configuration.`,
    RunE:  runAppViewServer,
}

func init() {
    rootCmd.AddCommand(runCmd)
    
    // AppView configuration flags
    runCmd.Flags().String("host", "localhost", "AppView host")
    runCmd.Flags().Int("port", 8081, "AppView port")
    runCmd.Flags().String("pds-url", "http://localhost:8080", "PDS URL")
    runCmd.Flags().String("config", "", "Configuration file path")
    runCmd.Flags().Bool("dev", false, "Enable development mode")
}

func runAppViewServer(cmd *cobra.Command, args []string) error {
    // Start AppView server that aggregates from PDS
    return nil
}
```

### Configuration Management
```go
// internal/config/pds_config.go
type PDSConfig struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Atproto  AtprotoConfig  `mapstructure:"atproto"`
}

type ServerConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
    Dev  bool   `mapstructure:"dev"`
}

type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Database string `mapstructure:"database"`
    Username string `mapstructure:"username"`
    Password string `mapstructure:"password"`
}

type AtprotoConfig struct {
    DIDResolver string `mapstructure:"did_resolver"`
    HandleBase  string `mapstructure:"handle_base"`
}

// internal/config/appview_config.go
type AppViewConfig struct {
    Server ServerConfig `mapstructure:"server"`
    PDS    PDSConfig    `mapstructure:"pds"`
}

type ServerConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
    Dev  bool   `mapstructure:"dev"`
}

type PDSConfig struct {
    URL string `mapstructure:"url"`
}
```

## Shared Utilities

### Shared Commands Binary
```go
// cmd/shared/main.go
var rootCmd = &cobra.Command{
    Use:   "hashpost",
    Short: "HashPost shared utilities",
    Long:  `HashPost shared utilities for database management and code generation.`,
}

var migrateCmd = &cobra.Command{
    Use:   "migrate",
    Short: "Database migration commands",
    Long:  `Manage database migrations for HashPost.`,
}

var generateCmd = &cobra.Command{
    Use:   "generate",
    Short: "Code generation commands",
    Long:  `Generate code from OpenAPI specs and database schemas.`,
}

func init() {
    rootCmd.AddCommand(migrateCmd)
    rootCmd.AddCommand(generateCmd)
}
```

### Deployment Patterns
```go
// cmd/deploy/main.go
var rootCmd = &cobra.Command{
    Use:   "hashpost-deploy",
    Short: "HashPost deployment utilities",
    Long:  `Deploy HashPost PDS and AppView components.`,
}

var deployPDS = &cobra.Command{
    Use:   "pds",
    Short: "Deploy PDS component",
    Long:  `Deploy HashPost PDS to target environment.`,
}

var deployAppView = &cobra.Command{
    Use:   "appview",
    Short: "Deploy AppView component",
    Long:  `Deploy HashPost AppView to target environment.`,
}
```

### Configuration Pattern
```go
// cmd/pds/config.go
func bindPDSFlags(cmd *cobra.Command, cfg *config.PDSConfig) {
    cmd.Flags().StringVar(&cfg.Server.Host, "host", "localhost", "PDS host")
    cmd.Flags().IntVar(&cfg.Server.Port, "port", 8080, "PDS port")
    cmd.Flags().BoolVar(&cfg.Server.Dev, "dev", false, "Development mode")
}

// cmd/appview/config.go
func bindAppViewFlags(cmd *cobra.Command, cfg *config.AppViewConfig) {
    cmd.Flags().StringVar(&cfg.Server.Host, "host", "localhost", "AppView host")
    cmd.Flags().IntVar(&cfg.Server.Port, "port", 8081, "AppView port")
    cmd.Flags().StringVar(&cfg.PDS.URL, "pds-url", "http://localhost:8080", "PDS URL")
}
```

## Implementation Plan

### Phase 1: Separate Binaries
- [ ] Set up cobra CLI framework for each binary
- [ ] Create PDS binary (`hashpost-pds`)
- [ ] Create AppView binary (`hashpost-appview`)
- [ ] Create shared utilities binary (`hashpost`)
- [ ] Add basic configuration for each component

### Phase 2: Server Integration
- [ ] Integrate PDS with net/http server and database
- [ ] Integrate AppView with net/http server and PDS APIs
- [ ] Add configuration management for each component
- [ ] Implement graceful shutdown for each component
- [ ] Add logging for each component

### Phase 3: Advanced Features
- [ ] Add deployment commands
- [ ] Implement configuration validation
- [ ] Add help and documentation for each binary
- [ ] Add testing for each component
- [ ] Add shared utilities (migrate, generate)

## Key Components

### Binaries
- **hashpost-pds**: Personal Data Server with database access
- **hashpost-appview**: Stateless aggregator for data presentation
- **hashpost**: Shared utilities (migrate, generate, etc.)
- **hashpost-deploy**: Deployment utilities

### Configuration
- **PDS Config**: Database, atproto, and server configuration
- **AppView Config**: Server and PDS connection configuration
- **File-based**: YAML/JSON configuration files
- **Environment**: Environment variable support
- **Flags**: Command-line flag support
- **Validation**: Configuration validation

### Server Management
- **Independent Deployment**: Each component can be deployed separately
- **Startup**: Server initialization for each component
- **Shutdown**: Graceful shutdown handling for each component
- **Health Checks**: Server health monitoring for each component
- **Logging**: Structured logging for each component

## Success Criteria

### Usability
- Clear command structure
- Helpful error messages
- Comprehensive help system
- Easy configuration

### Maintainability
- Clean command organization
- Easy to add new commands
- Good separation of concerns
- Comprehensive testing

### Performance
- Fast startup time
- Efficient configuration loading
- Minimal memory overhead
- Quick command execution

## Next Steps

1. **Set up separate binaries** - Initialize CLI structure for each component
2. **Implement PDS binary** - Basic PDS server command with database access
3. **Implement AppView binary** - Basic AppView server command with PDS integration
4. **Add shared utilities** - Migration and generation commands
5. **Add testing** - Test each component independently

## Notes

- Use cobra for consistent CLI experience
- Follow Go CLI best practices
- Plan for future command additions
- Keep commands focused and single-purpose

---

*Last Updated: [Current Date]*
*Status: Design Phase*
