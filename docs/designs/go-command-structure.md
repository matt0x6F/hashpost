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
- **Root Command**: `hashpost` - Main application entry point
- **Run Command**: `hashpost run` - Start the server
- **Future Commands**: `migrate`, `generate`, etc. as needed

### Framework Choice
- **Cobra**: Industry standard for Go CLI applications
- **Viper**: Configuration management integration
- **Standard Library**: Use net/http for server implementation

## Command Implementation

### Root Command Structure
```go
// cmd/root.go
var rootCmd = &cobra.Command{
    Use:   "hashpost",
    Short: "HashPost - A modern forum platform",
    Long:  `HashPost is a modern, scalable forum platform built with Go.`,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Run Command Implementation
```go
// cmd/run.go
var runCmd = &cobra.Command{
    Use:   "run",
    Short: "Start the HashPost server",
    Long:  `Start the HashPost server with the specified configuration.`,
    RunE:  runServer,
}

func init() {
    rootCmd.AddCommand(runCmd)
    
    // Server configuration flags
    runCmd.Flags().String("host", "localhost", "Server host")
    runCmd.Flags().Int("port", 8080, "Server port")
    runCmd.Flags().String("config", "", "Configuration file path")
    runCmd.Flags().Bool("dev", false, "Enable development mode")
}

func runServer(cmd *cobra.Command, args []string) error {
    // Implementation here
    return nil
}
```

### Configuration Management
```go
// internal/config/config.go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    API      APIConfig      `mapstructure:"api"`
}

type ServerConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
    Dev  bool   `mapstructure:"dev"`
}

func LoadConfig() (*Config, error) {
    // Load from file, environment, and flags
}
```

## Command Patterns

### Server Command Pattern
```go
// cmd/server.go
type ServerCommand struct {
    config *config.Config
    server *http.Server
}

func NewServerCommand() *ServerCommand {
    return &ServerCommand{}
}

func (s *ServerCommand) Execute(cmd *cobra.Command, args []string) error {
    // Initialize server
    // Start server
    // Handle shutdown
    return nil
}
```

### Configuration Pattern
```go
// cmd/config.go
func bindFlags(cmd *cobra.Command, cfg *config.Config) {
    cmd.Flags().StringVar(&cfg.Server.Host, "host", "localhost", "Server host")
    cmd.Flags().IntVar(&cfg.Server.Port, "port", 8080, "Server port")
    cmd.Flags().BoolVar(&cfg.Server.Dev, "dev", false, "Development mode")
}
```

## Implementation Plan

### Phase 1: Basic Structure
- [ ] Set up cobra CLI framework
- [ ] Create root command
- [ ] Implement `run` command
- [ ] Add basic configuration

### Phase 2: Server Integration
- [ ] Integrate with net/http server
- [ ] Add configuration management
- [ ] Implement graceful shutdown
- [ ] Add logging

### Phase 3: Advanced Features
- [ ] Add more commands as needed
- [ ] Implement configuration validation
- [ ] Add help and documentation
- [ ] Add testing

## Key Components

### Commands
- **Root Command**: Application entry point
- **Run Command**: Server startup and management
- **Future Commands**: Migrate, generate, etc.

### Configuration
- **File-based**: YAML/JSON configuration files
- **Environment**: Environment variable support
- **Flags**: Command-line flag support
- **Validation**: Configuration validation

### Server Management
- **Startup**: Server initialization
- **Shutdown**: Graceful shutdown handling
- **Health Checks**: Server health monitoring
- **Logging**: Structured logging

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

1. **Set up cobra framework** - Initialize CLI structure
2. **Implement run command** - Basic server command
3. **Add configuration** - File and environment support
4. **Integrate server** - Connect with net/http server
5. **Add testing** - Test command functionality

## Notes

- Use cobra for consistent CLI experience
- Follow Go CLI best practices
- Plan for future command additions
- Keep commands focused and single-purpose

---

*Last Updated: [Current Date]*
*Status: Design Phase*
