# Platform Administration

## Overview
Platform administration provides centralized management capabilities for system-wide settings, user management, and content moderation across all communities.

## Capabilities
- **system_admin**: Full platform access including system settings and analytics
- **user_management**: User account management and role assignment
- **moderation**: Platform-wide content moderation tools
- **compliance**: Legal and compliance request handling
- **legal_requests**: Legal team access to user correlation data
- **cross_user_correlation**: Access to cross-user identity correlation tools

## API Endpoints
- `PUT /platform/rules` - Update platform-wide rules (requires system_admin)
- `GET /platform/rules` - Retrieve platform rules
- `PUT /system/settings` - Update system settings (requires system_admin)
- `GET /system/settings` - Retrieve system settings (requires system_admin)

## System Settings
Platform rules and system configuration are stored as JSONB in the system_settings table, accessible only to users with system_admin capability.

## User Interface
The platform administration dashboard is accessible at `/admin` and provides:
- User management interface
- Content moderation tools
- System analytics and monitoring
- Platform rules configuration
- System settings management
- Cross-user correlation tools (for authorized roles)

## Access Control
All administrative functions require appropriate capabilities. Users without the necessary permissions will be redirected to the main platform.
