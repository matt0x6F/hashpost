# HashPost PDS Design

## Project Overview

This document tracks the implementation of the **HashPost PDS** (Personal Data Server) for atproto protocol compliance. The HashPost PDS handles atproto protocol requirements, identity management, and data storage, while the HashPost AppView provides custom application logic and business rules.

## Design Goals

### Primary Objectives
- **Atproto Compatibility**: Full compliance with atproto protocol specifications
- **Decentralized Identity**: Leverage atproto's decentralized identity system
- **Custom Types**: Use custom atproto types for HashPost-specific features
- **Protocol Compliance**: Follow atproto authentication, data, and API patterns
- **Dual Architecture**: Build HashPost AppView + PDS for complete atproto compliance

### Technical Goals
- Implement HashPost PDS using Bluesky Indigo Go libraries
- Implement HashPost AppView for custom application logic
- Support atproto authentication and identity using Indigo packages
- Define custom atproto types for HashPost features
- Integrate AppView with PDS for complete functionality

## Architecture Decisions

### HashPost AppView
- **Purpose**: Stateless aggregator for data presentation
- **API**: OpenAPI spec server for unified data presentation
- **Authentication**: Integrates with PDS for atproto identity
- **Data**: Aggregates data from PDS via APIs

### HashPost PDS
- **Purpose**: Atproto protocol compliance, data storage, and business logic
- **Implementation**: Built using Bluesky Indigo Go libraries
- **API**: Atproto endpoints (`/xrpc/com.atproto.*`) + custom HashPost endpoints
- **Authentication**: DID-based authentication system using Indigo packages
- **Data Structures**: Custom atproto types for HashPost features
- **Business Logic**: RBAC, moderation, voting, ownership models

### Integration
- **AppView → PDS**: AppView aggregates data from PDS via APIs
- **Data Flow**: All business logic in PDS, AppView is stateless aggregator
- **Authentication**: PDS handles all identity and business logic

### AppView Integration Points
- **Data Aggregation**: PDS provides APIs for AppView to aggregate data
- **Atproto Operations**: PDS handles all atproto protocol operations
- **Business Logic**: PDS handles all RBAC, moderation, and business logic
- **User Context**: PDS provides complete user context and permissions to AppView
- **Error Propagation**: PDS returns structured errors for AppView handling
- **AppView Role**: Stateless aggregator that presents unified view to clients

## Bluesky Indigo Libraries

### Key Go Packages
- **api/atproto**: Generated types for `com.atproto.*` Lexicons
- **api/bsky**: Generated types for `app.bsky.*` Lexicons  
- **atproto/crypto**: Cryptographic signing and key serialization
- **atproto/identity**: DID and handle resolution
- **atproto/syntax**: String types and parsers for identifiers
- **atproto/lexicon**: Schema validation of data
- **mst**: Merkle Search Tree implementation
- **repo**: Account data storage
- **xrpc**: HTTP API client

### Implementation Strategy
- **PDS Core**: Use Indigo's `pds` package as foundation
- **Authentication**: Leverage `atproto/identity` and `atproto/crypto`
- **Data Storage**: Use `repo` package for account data
- **API Endpoints**: Implement using Indigo's generated types
- **Custom Types**: Define HashPost-specific atproto lexicons

## Atproto Components

### Core Data Structures
- **Profiles**: Custom HashPost profile types
- **Posts**: Custom HashPost post types
- **Subforum Subscriptions**: Custom HashPost subforum subscription types
- **Votes**: Custom HashPost voting/engagement types
- **Comments**: Custom HashPost comment/reply types
- **Moderation**: Custom HashPost moderation and reporting types

### Authentication System
- **DID Resolution**: Handle decentralized identifiers
- **Handle Resolution**: Username to DID mapping
- **Session Management**: Atproto session tokens
- **Identity Verification**: Cryptographic identity verification

### API Endpoints
- **Authentication**: Login, session management
- **Profile Management**: Create, update, delete profiles
- **Content Management**: Create, update, delete posts
- **Subforum Features**: Subscribe, unsubscribe to subforums
- **Engagement**: Vote on posts, comment, report content
- **Timeline**: Feed generation based on subforum subscriptions
- **Search**: Content and user search within subforums

### PDS Database Access
- **Database Layer**: Only PDS component accesses PostgreSQL database
- **Transaction Boundaries**: PDS handles all transaction boundaries for atproto data and business logic
- **Query Patterns**: PDS uses sqlc for all queries - atproto protocol, business logic, and RBAC
- **Data Consistency**: PDS ensures both atproto data consistency and business rule consistency

## Implementation Plan

### Phase 1: Indigo Research
- [x] Study Bluesky Indigo Go libraries
- [x] Understand Indigo package structure
- [x] Research Indigo PDS implementation patterns
- [x] Analyze Indigo authentication and data handling

### Phase 2: Core Implementation
- [x] Set up Indigo dependencies
- [x] Create Docker Compose development and test environments
- [x] Set up testing infrastructure and test database
- [x] Create comprehensive database schema for HashPost features
- [ ] Implement PDS using Indigo packages
- [ ] Set up DID and handle resolution using `atproto/identity`
- [ ] Implement authentication using `atproto/crypto`

### Phase 3: Feature Implementation
- [ ] Implement profile management using Indigo types
- [ ] Implement post creation and management
- [ ] Implement subforum subscription features
- [ ] Implement voting and comment systems
- [ ] Implement timeline generation based on subforum subscriptions
- [ ] Create test vectors for atproto operations

### Phase 4: Testing & Validation
- [ ] Implement unit tests for PDS components
- [ ] Implement integration tests for atproto endpoints
- [ ] Test with atproto clients (Bluesky web, mobile apps)
- [ ] Validate protocol compliance
- [ ] Performance testing and optimization

### Phase 5: Client Integration
- [ ] Test with HashPost-specific atproto clients
- [ ] Ensure custom type compatibility
- [ ] Implement advanced features
- [ ] End-to-end testing and validation

## Key Components

### Atproto Server
- **PDS Implementation**: Built using Bluesky Indigo libraries
- **Authentication**: DID-based authentication system using Indigo packages
- **Data Management**: Atproto data structure handling with Indigo types
- **API Endpoints**: Atproto-compatible API using Indigo generated types

### Data Structures
- **Profiles**: User profile data
- **Posts**: Content posts
- **Subforum Subscriptions**: User subscriptions to subforums
- **Engagement**: Votes, comments, reports

### Client Support
- **HashPost Clients**: Support for HashPost-specific atproto client applications
- **Custom Types**: Support for HashPost custom atproto types
- **API Compliance**: Full atproto API compliance

## Success Criteria

### Protocol Compliance
- Full atproto protocol compliance
- Successful authentication with atproto clients
- Proper data structure implementation
- API endpoint compatibility

### Custom Type Support
- Support for HashPost custom atproto types
- Proper data exchange with custom types
- Identity resolution
- HashPost-specific client applications

### Performance
- Fast authentication and data access
- Efficient timeline generation
- Scalable social graph operations
- Responsive API endpoints

## Testing Strategy

### Early Verification Approach
Since there are no dedicated atproto PDS validation tools, we'll implement a comprehensive validation strategy using existing atproto implementations as reference, custom validation tools, and real-world testing scenarios.

### Validation Approaches

#### 1. Reference Implementation Validation
- **Millipds Comparison**: Compare our PDS behavior against the Python Millipds implementation
- **Picopds Benchmarking**: Use the minimal Picopds as a compliance baseline
- **Hexpds Reference**: Cross-reference with the Elixir PDS implementation
- **@waverlyai/atproto-pds**: Use TypeScript reference implementation for API validation

#### 2. Atproto Protocol Validation
- **Lexicon Schema Validation**: Implement validators for atproto string formats:
  - `at-identifier` validation (handles, DIDs)
  - `at-uri` validation (record URIs)
  - `cid` validation (content identifiers)
  - `datetime` validation (ISO 8601 timestamps)
- **DID Document Parsing**: Validate DID resolution and document parsing:
  - `alsoKnownAs` array for current handles
  - `verificationMethod` array for public signing keys
  - `service` array for PDS service endpoints
- **MST Integrity Validation**: Verify Merkle Search Tree structure and hashes

#### 3. XRPC Endpoint Validation
- **Core Endpoints**: Test all required `com.atproto.*` endpoints:
  - `com.atproto.server.createSession` - Authentication
  - `com.atproto.server.createAccount` - Account creation
  - `com.atproto.identity.resolveHandle` - Handle resolution
  - `com.atproto.repo.createRecord` - Record creation
  - `com.atproto.repo.getRecord` - Record retrieval
  - `com.atproto.repo.listRecords` - Record listing
- **Response Validation**: Ensure responses match atproto lexicon schemas
- **Error Handling**: Validate error responses follow atproto error format

#### 4. Repository Operations Validation
- **CAR File Import/Export**: Test repository data integrity using CAR files
- **Record CRUD Operations**: Validate all record operations work correctly
- **Commit Validation**: Verify commit signatures and ordering
- **MST Operations**: Test Merkle Search Tree operations and integrity

### Concrete Validation Approaches

#### 1. Atproto String Format Validation
```go
// Custom validators for atproto string formats
func ValidateAtIdentifier(identifier string) error {
    // Validate handles (e.g., "alice.bsky.social")
    // Validate DIDs (e.g., "did:plc:abc123")
}

func ValidateAtURI(uri string) error {
    // Validate record URIs (e.g., "at://alice.bsky.social/app.bsky.feed.post/123")
}

func ValidateCID(cid string) error {
    // Validate content identifiers
}
```

#### 2. DID Document Validation
```go
// Validate DID document parsing
func ValidateDIDDocument(doc *DIDDocument) error {
    // Check alsoKnownAs array for current handles
    // Validate verificationMethod array for public keys
    // Verify service array for PDS endpoints
}
```

#### 3. XRPC Endpoint Testing
```go
// Test core atproto endpoints
func TestCreateSession(t *testing.T) {
    // Test com.atproto.server.createSession
    // Validate request/response format
    // Test error handling
}

func TestCreateAccount(t *testing.T) {
    // Test com.atproto.server.createAccount
    // Validate account creation flow
}
```

#### 4. Reference Implementation Comparison
- **Millipds Integration**: Run our PDS alongside Millipds and compare responses
- **API Compatibility**: Ensure our endpoints return identical responses to reference implementations
- **Data Integrity**: Compare repository operations between implementations

### Test Environment Setup

#### Docker Compose Development Environment
- **PDS Service**: HashPost PDS running in Docker container
- **AppView Service**: HashPost AppView running in Docker container
- **PostgreSQL Database**: Test database with migrations applied
- **Redis Cache**: For session management and caching
- **Test Data**: Pre-populated test users, posts, and subforum subscriptions
- **Mock Services**: Mock external services (DID resolution, other PDS instances)
- **Configuration**: Environment-specific configuration via Docker Compose

#### Docker Compose Services
```yaml
# docker-compose.test.yml
services:
  hashpost-pds:
    build: .
    command: hashpost-pds run --config /app/config/test.yaml
    environment:
      - DATABASE_URL=postgres://hashpost:password@postgres:5432/hashpost_test
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis

  hashpost-appview:
    build: .
    command: hashpost-appview run --config /app/config/test.yaml
    environment:
      - PDS_URL=http://hashpost-pds:8080
    depends_on:
      - hashpost-pds

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=hashpost_test
      - POSTGRES_USER=hashpost
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_test_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_test_data:/data
```

#### Test Data Management
- **Test Vectors**: Standardized test data for atproto operations
- **Fixtures**: Reusable test data for common scenarios
- **Database Seeding**: Automated test data setup and teardown via Docker Compose
- **Isolation**: Each test run starts with fresh Docker Compose environment
- **Volume Management**: Persistent test data volumes for consistent testing

### Key Test Scenarios

#### Authentication & Identity
- **DID Resolution**: Test DID to handle resolution
- **Session Creation**: Test `com.atproto.server.createSession`
- **Session Validation**: Test session token validation
- **Account Creation**: Test `com.atproto.server.createAccount`
- **Handle Resolution**: Test `com.atproto.identity.resolveHandle`

#### Repository Operations
- **Record Creation**: Test creating posts, profiles, follows
- **Record Updates**: Test updating existing records
- **Record Deletion**: Test soft and hard deletion
- **MST Operations**: Test Merkle Search Tree operations
- **Blob Management**: Test media file upload and storage

#### HashPost Features
- **Subforum Subscriptions**: Test subscribe/unsubscribe to subforums
- **Timeline Generation**: Test feed generation based on subforum subscriptions
- **Voting System**: Test voting on posts and comments
- **Comment System**: Test commenting and reply functionality
- **Search**: Test content and user search within subforums

#### Error Handling
- **Invalid Requests**: Test malformed requests
- **Authentication Failures**: Test invalid credentials
- **Rate Limiting**: Test anti-abuse measures
- **Network Failures**: Test resilience to external service failures

### Validation Tools & Infrastructure

#### Custom Validation Tools
- **Atproto String Validators**: Custom validators for atproto string formats
  - `at-identifier` (handles, DIDs)
  - `at-uri` (record URIs) 
  - `cid` (content identifiers)
  - `datetime` (ISO 8601 timestamps)
- **DID Document Parser**: Custom DID document validation and parsing
- **MST Validator**: Merkle Search Tree integrity validation
- **Lexicon Schema Validator**: Custom tool to validate atproto lexicon compliance

#### Reference Implementation Testing
- **Millipds Integration**: Compare our PDS against Python Millipds
- **Picopds Baseline**: Use minimal PDS as compliance reference
- **Hexpds Cross-Reference**: Validate against Elixir PDS implementation
- **TypeScript Reference**: Use @waverlyai/atproto-pds for API validation

#### XRPC Testing Framework
- **Custom XRPC Client**: Built-in atproto client for endpoint testing
- **Endpoint Test Suite**: Automated testing of all `com.atproto.*` endpoints
- **Response Validator**: Validate responses against atproto schemas
- **Error Handler Tests**: Test atproto error response format

#### Go Testing Framework
- **Standard Library**: `testing` package for unit tests
- **Testify**: Assertions and test suites
- **Testcontainers**: Database integration testing
- **Gomock**: Mock generation for external dependencies

#### Docker Compose Testing Workflows
- **Test Environment Setup**: `docker-compose -f docker-compose.test.yml up -d`
- **Test Execution**: Run tests against Docker Compose services
- **Environment Cleanup**: `docker-compose -f docker-compose.test.yml down -v`
- **Database Reset**: Fresh database state for each test run
- **Service Health Checks**: Verify all services are healthy before testing

#### CI/CD Integration
- **Automated Testing**: Run tests on every commit using Docker Compose
- **Test Reports**: Generate coverage and performance reports
- **Test Database**: Automated test database setup and teardown via Docker Compose
- **Parallel Testing**: Run tests in parallel for faster feedback
- **Docker Build**: Automated Docker image building for test environments

### Success Metrics

#### Protocol Compliance
- **100% Atproto Endpoint Coverage**: All required endpoints implemented and tested
- **Authentication Success Rate**: 99.9% successful authentication flows
- **Data Integrity**: Zero data corruption in repository operations
- **API Response Times**: < 100ms for simple operations, < 500ms for complex operations

#### Test Coverage
- **Code Coverage**: 90%+ coverage on PDS core functionality
- **Endpoint Coverage**: 100% coverage of atproto endpoints
- **Scenario Coverage**: All critical user journeys tested
- **Error Coverage**: All error conditions tested and handled

#### Performance
- **Load Testing**: PDS handles expected concurrent users
- **Memory Usage**: Stable memory usage under load
- **Database Performance**: Efficient queries and transactions
- **Response Times**: Meets performance requirements
- **Subforum Scalability**: Handles large numbers of subforum subscriptions

## Docker Compose Development Workflow

### Development Environment
- **Local Development**: Use Docker Compose for consistent development environment
- **Service Dependencies**: PDS, AppView, PostgreSQL, and Redis all managed via Docker Compose
- **Hot Reloading**: Development containers with hot reloading for faster iteration
- **Database Migrations**: Automated migration application on container startup
- **Environment Variables**: Configuration via Docker Compose environment files

### Development Commands
```bash
# Start development environment
docker-compose up -d

# Run tests against Docker Compose services
docker-compose -f docker-compose.test.yml up -d
go test ./...

# Reset test environment
docker-compose -f docker-compose.test.yml down -v

# View logs
docker-compose logs -f hashpost-pds
docker-compose logs -f hashpost-appview

# Execute commands in containers
docker-compose exec hashpost-pds hashpost migrate up
docker-compose exec postgres psql -U hashpost -d hashpost_dev
```

### Development Configuration
- **Environment Files**: `.env.dev`, `.env.test` for different environments
- **Volume Mounts**: Source code mounted for hot reloading
- **Port Mapping**: Exposed ports for local development access
- **Health Checks**: Service health monitoring for reliable testing

## Phase 2 Implementation Complete

### Infrastructure Setup
- **Project Structure**: Complete directory structure with cmd/, internal/, config/ directories
- **Cobra CLI**: PDS and AppView binaries created with cobra framework
- **Taskfile**: Comprehensive development workflow with targets for dev, test, migrate, generate, build
- **Docker Compose**: Development and test environments configured
- **Database Schema**: Complete HashPost schema supporting atproto records and HashPost-specific features
- **Dependencies**: All Bluesky Indigo packages added and working
- **Testing Infrastructure**: Docker Compose test environment ready for PDS validation

## Next Steps

1. **Implement PDS with Indigo**: Use Indigo packages for PDS implementation
2. **Design Custom Types**: Plan HashPost-specific atproto type definitions
3. **Create API Endpoints**: Atproto-compatible API using Indigo generated types
4. **Test with HashPost Clients**: Ensure compatibility with HashPost-specific clients
5. **Create Test Vectors**: Develop standardized test data for atproto operations

## Notes

- Focus on atproto protocol compliance using Bluesky Indigo libraries
- Leverage Indigo Go packages for PDS implementation
- Plan for HashPost-specific custom types using Indigo's type system
- Consider HashPost client application support
- Use Indigo's generated types for API endpoints

---

*Last Updated: [Current Date]*
*Status: Research Phase*
