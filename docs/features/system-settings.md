# System Settings

## Overview
System settings provide a centralized way to manage platform-wide configuration and rules.

## Setting Types
- **string**: Plain text values
- **boolean**: true/false values
- **number**: Numeric values
- **json**: Structured JSON data

## Platform Rules
Platform rules are stored as JSON in the system_settings table with the key "platform_rules". These rules apply to all communities and content across the platform.

## Access Control
All system settings operations require the system_admin capability. Settings are automatically validated based on their declared type.

## API Usage
```http
PUT /system/settings
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "body": {
    "setting_key": "platform_rules",
    "setting_value": "[{\"code\": \"no_hate_speech\", \"title\": \"No Hate Speech\", \"description\": \"Hate speech is prohibited\"}]",
    "setting_type": "json"
  }
}
```

## Validation
- String values are always accepted
- Boolean values must be "true" or "false"
- Number values must be valid numeric format
- JSON values must be valid JSON syntax

## Storage
Settings are stored in the `system_settings` table with:
- `setting_key`: Unique identifier for the setting
- `setting_value`: The actual setting value
- `setting_type`: Data type for validation
- `created_by`: User ID who created/modified the setting
- `created_at`: Timestamp of creation/modification
