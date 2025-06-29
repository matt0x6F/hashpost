This is a [Next.js](https://nextjs.org) project bootstrapped with [`create-next-app`](https://nextjs.org/docs/app/api-reference/cli/create-next-app).

## Getting Started

**The UI is designed to be run as part of the standard Docker Compose development environment.**

### 1. Start the full development environment (API + UI + DB):

From the project root:

```bash
make dev
```

This will start the API server, database, and UI in Docker Compose. The UI will be available at [http://localhost:3000](http://localhost:3000).

**Do not run the UI with `npm run dev` or other standalone Next.js commands.** The UI expects the backend and database to be available via Docker Compose networking.

### 2. Editing the UI

You can edit UI files in `ui/` as usual. Docker Compose will hot-reload changes.

---

# HashPost API Client Generation

This document explains how to generate and use the TypeScript API client for HashPost.

## Overview

The API client is generated from the OpenAPI schema that Huma automatically creates from your Go API. This ensures type safety and automatic updates when the API changes.

## Setup

### 1. Install Dependencies

```bash
cd ui
npm install
```

This installs the OpenAPI Generator CLI tool.

### 2. Generate the API Client

Make sure your HashPost server is running (via Docker Compose):

```bash
# In the root directory
make dev
```

Then generate the API client:

```bash
# In the ui directory
npm run generate-api
```

This will:
1. Download the OpenAPI schema from your running server
2. Generate TypeScript types and API client code
3. Place the generated files in `src/generated/api/`

### 3. Use the Generated Client

The generated client is available in `src/lib/api-client.ts`. It provides:

- Type-safe API calls
- Automatic authentication handling (JWT cookies)
- Centralized configuration for all generated API classes

## Available Scripts

- `npm run download-openapi` - Download the OpenAPI schema from the server
- `npm run generate-api` - Download schema and generate the full API client
- `npm run generate-api-local` - Generate client from local `openapi.json` file

## Usage Examples

```typescript
import { getApi } from '@/lib/api-client';
import { AuthenticationApi } from '@/generated/api/src/apis/AuthenticationApi';

const authApi = getApi(AuthenticationApi);
const loginResult = await authApi.loginUser({
  email: 'user@example.com',
  password: 'password123'
});
```

## Generated Files Structure

```
src/generated/api/
├── index.ts              # Main exports
├── runtime.ts            # Runtime configuration
├── models/               # TypeScript interfaces for API models
├── apis/                 # Generated API classes
└── ...
```

## Customization

The generated client can be customized by modifying:

- `openapi-generator-config.json` - Configuration for the generator
- `src/lib/api-client.ts` - Centralized config and helper for generated clients

## Development Workflow

1. **API Changes**: When you modify your Go API endpoints, the OpenAPI schema automatically updates
2. **Regenerate Client**: Run `npm run generate-api` to update the TypeScript client
3. **Type Safety**: The generated client provides full type safety for all API calls

## Troubleshooting

### Server Not Running
If you get an error about the server not being available:
```bash
# Start the server first (in Docker Compose)
make dev

# Then generate the client
npm run generate-api
```

### Generated Files Not Updating
If the generated files don't seem to update:
```bash
# Clean and regenerate
rm -rf src/generated/
npm run generate-api
```

### TypeScript Errors
If you get TypeScript errors after generation:
```bash
# Restart the TypeScript server in your IDE
# Or run type checking
npm run type-check
```

## Integration with Existing Code

The generated client is designed to be used directly in your UI code. You can:

1. Import the API class you need from `@/generated/api/src/apis/`.
2. Use `getApi(YourApiClass)` from `@/lib/api-client` to get a pre-configured instance.

Example:
```typescript
import { getApi } from '@/lib/api-client';
import { UsersApi } from '@/generated/api/src/apis/UsersApi';

const usersApi = getApi(UsersApi);
const profile = await usersApi.getUserProfile();
```

## Best Practices

1. **Always regenerate** the client after API changes
2. **Commit the generated files** to version control (they're type-safe)
3. **Use the centralized config** in `api-client.ts` for all API calls
4. **Test the generated client** with your existing components

---

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
