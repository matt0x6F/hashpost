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

### Command Structure
- **PDS Binary**: `hashpost-pds` - Personal Data Server for atproto protocol compliance
- **AppView Binary**: `hashpost-appview` - Stateful web application for forum functionality
- **Taskfile Targets**: Shared utilities via Taskfile (migrate, generate, test, dev)
- **PDS Focus**: PDS handles atproto protocol compliance and canonical data storage
- **AppView Focus**: AppView is stateful web application with business logic
- **Independent Deployment**: Each component can be deployed separately
- **Development Workflow**: Taskfile manages common development tasks

### Framework Choice
- **Cobra**: Industry standard for Go CLI applications
- **Viper**: Configuration management integration
- **Standard Library**: Use net/http for server implementation

## Command Implementation

### Taskfile Targets
The Taskfile provides common development tasks including:
- `task dev` - Start development environment
- `task test` - Run tests with Docker Compose
- `task migrate:up/down` - Database migrations
- `task generate:sqlc` - Generate database code
- `task generate:openapi` - Generate OpenAPI client
- `task build` - Build all binaries
- `task clean` - Clean all artifacts

See `Taskfile.yml` for complete implementation details.

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
// internal/config/shared.go
type ServerConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
    Dev  bool   `mapstructure:"dev"`
}

// internal/config/pds_config.go
type PDSConfig struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Atproto  AtprotoConfig  `mapstructure:"atproto"`
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
    PDS    PDSConnection `mapstructure:"pds"`
}

type PDSConnection struct {
    URL string `mapstructure:"url"`
}
```

## Development Workflow

### Taskfile Management
- **Common Tasks**: All shared utilities managed via Taskfile targets
- **Development**: `task dev` starts Docker Compose environment
- **Testing**: `task test` runs full test suite with Docker Compose
- **Database**: `task migrate:up` and `task migrate:down` for migrations
- **Code Generation**: `task generate:sqlc` and `task generate:openapi`
- **Building**: `task build` compiles all binaries

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

### Phase 1: Separate Binaries + Taskfile
- [x] Set up cobra CLI framework for PDS and AppView binaries
- [x] Create PDS binary (`hashpost-pds`)
- [x] Create AppView binary (`hashpost-appview`)
- [x] Create Taskfile with common development tasks
- [x] Add basic configuration for each component

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

## Implementation Complete

### Phase 1: Foundation ✅
- **✅ Cobra Binaries**: Both PDS and AppView binaries created with cobra framework
- **✅ Taskfile**: Comprehensive development workflow with targets for all common tasks
- **✅ Configuration**: Development and test configuration files created
- **✅ Docker Integration**: Taskfile targets integrate with Docker Compose for development and testing
- **✅ Project Structure**: Clean separation between PDS, AppView, and shared utilities

### Phase 2: Core Implementation ✅
- **✅ Database Integration**: SQLC queries and migrations working
- **✅ Event Processing**: NATS integration for real-time updates
- **✅ RBAC System**: Role-based access control with SQLC
- **✅ Authentication**: DID-based auth with OAuth 2.0 support

### Phase 3: Advanced Features ✅
- **✅ RBAC Refactoring**: Converted all raw SQL to SQLC queries
- **✅ DID Resolution**: Identity service with caching
- **✅ Handler Integration**: All handlers use generated queries
- **✅ Test Coverage**: All tests passing with SQLC integration

### Available Commands
- `task dev` - Start development environment
- `task test:unit` - Run unit tests
- `task generate:sqlc` - Generate database code
- `task generate:openapi` - Generate API client
- `task build` - Build all binaries
- `task clean` - Clean all artifacts

### Current Status
- ✅ Complete command structure for both PDS and AppView
- ✅ All database operations use SQLC generated code
- ✅ Zero raw SQL queries in codebase
- ✅ Type-safe database operations throughout
- ✅ All tests passing with comprehensive functionality

## Notes

- Use cobra for consistent CLI experience
- Follow Go CLI best practices
- Plan for future command additions
- Keep commands focused and single-purpose
- Taskfile provides better developer experience than separate shared binary

---

*Last Updated: [Current Date]*
*Status: Phase 1 Complete - Ready for Phase 2*
