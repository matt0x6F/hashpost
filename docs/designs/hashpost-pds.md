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
- **Follows**: Custom HashPost social graph types
- **Likes**: Custom HashPost engagement types
- **Reposts**: Custom HashPost content sharing types
- **Lists**: Custom HashPost curated content types

### Authentication System
- **DID Resolution**: Handle decentralized identifiers
- **Handle Resolution**: Username to DID mapping
- **Session Management**: Atproto session tokens
- **Identity Verification**: Cryptographic identity verification

### API Endpoints
- **Authentication**: Login, session management
- **Profile Management**: Create, update, delete profiles
- **Content Management**: Create, update, delete posts
- **Social Features**: Follow, unfollow, like, repost
- **Timeline**: Feed generation and management
- **Search**: Content and user search

### PDS Database Access
- **Database Layer**: Only PDS component accesses PostgreSQL database
- **Transaction Boundaries**: PDS handles all transaction boundaries for atproto data and business logic
- **Query Patterns**: PDS uses sqlc for all queries - atproto protocol, business logic, and RBAC
- **Data Consistency**: PDS ensures both atproto data consistency and business rule consistency

## Implementation Plan

### Phase 1: Indigo Research
- [ ] Study Bluesky Indigo Go libraries
- [ ] Understand Indigo package structure
- [ ] Research Indigo PDS implementation patterns
- [ ] Analyze Indigo authentication and data handling

### Phase 2: Core Implementation
- [ ] Set up Indigo dependencies
- [ ] Implement PDS using Indigo packages
- [ ] Set up DID and handle resolution using `atproto/identity`
- [ ] Implement authentication using `atproto/crypto`

### Phase 3: Feature Implementation
- [ ] Implement profile management using Indigo types
- [ ] Implement post creation and management
- [ ] Implement social features (follow, like, repost)
- [ ] Implement timeline generation

### Phase 4: Client Integration
- [ ] Test with HashPost-specific atproto clients
- [ ] Ensure custom type compatibility
- [ ] Implement advanced features
- [ ] Performance optimization

## Key Components

### Atproto Server
- **PDS Implementation**: Built using Bluesky Indigo libraries
- **Authentication**: DID-based authentication system using Indigo packages
- **Data Management**: Atproto data structure handling with Indigo types
- **API Endpoints**: Atproto-compatible API using Indigo generated types

### Data Structures
- **Profiles**: User profile data
- **Posts**: Content posts
- **Social Graph**: Follow relationships
- **Engagement**: Likes, reposts, replies

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

## Next Steps

1. **Research Bluesky Indigo**: Deep dive into Indigo Go libraries and packages
2. **Design Custom Types**: Plan HashPost-specific atproto type definitions
3. **Implement PDS with Indigo**: Use Indigo packages for PDS implementation
4. **Create API Endpoints**: Atproto-compatible API using Indigo generated types
5. **Test with HashPost Clients**: Ensure compatibility with HashPost-specific clients

## Notes

- Focus on atproto protocol compliance using Bluesky Indigo libraries
- Leverage Indigo Go packages for PDS implementation
- Plan for HashPost-specific custom types using Indigo's type system
- Consider HashPost client application support
- Use Indigo's generated types for API endpoints

---

*Last Updated: [Current Date]*
*Status: Research Phase*
