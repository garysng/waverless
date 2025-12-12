# Waverless Web UI

Web-based user interface for managing Waverless serverless deployments.

## Environment Variables

Create a `.env` file in the `web-ui` directory to configure the application:

```bash
# API Backend URL
# The base URL for the Waverless API backend
# In production, set this to your actual backend URL
VITE_API_BACKEND_URL=http://localhost:8080
```

### Example Configurations

**Development (local):**
```bash
VITE_API_BACKEND_URL=http://localhost:8080
```

**Production:**
```bash
VITE_API_BACKEND_URL=https://api.yourcompany.com
```

**Docker deployment:**
```bash
VITE_API_BACKEND_URL=http://waverless-backend:8080
```

## Quick Start Feature

The Overview tab includes a collapsible Quick Start panel that allows you to test endpoints directly from the UI:

- **Collapsible**: Click to expand/collapse the panel to save screen space
- **Three test modes**:
  - `/run` (async) - Submit tasks asynchronously and get task ID immediately
  - `/runsync` (sync) - Submit tasks and wait for completion
  - `/status` (query) - Query task status by task ID
- **Left panel**: Submit test tasks with JSON input or query task status
- **Right panel**: View code examples in cURL, Python, and JavaScript
- Automatically uses the configured `VITE_API_BACKEND_URL` for API examples
- **Smart workflow**: When using `/run`, the returned task ID is auto-filled in the status query form

This makes it easy to:
1. Verify your endpoint is working
2. Test with sample inputs
3. Query task status for async operations
4. Copy code examples for integration
5. Keep the interface clean when not testing

## Development

```bash
# Install dependencies
pnpm install

# Start development server
pnpm dev

# Build for production
pnpm run build
```

## Production Build

The built static files will be in the `dist/` directory and can be served by any static file server or reverse proxy.
