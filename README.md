<div align="center">
  <a href="https://wavespeed.ai">
    <img src="docs/images/wavespeed-logo.png" alt="Wavespeed.ai Logo" width="200"/>
  </a>

  <h1>Waverless</h1>

  <p>
    <strong>High-performance Serverless GPU task orchestration system</strong>
  </p>

  <p>
    <a href="https://wavespeed.ai">🌐 Visit Wavespeed.ai</a> •
    <a href="docs/USER_GUIDE.md">📖 Documentation</a> •
    <a href="https://github.com/wavespeedai/waverless/issues">💬 Issues</a>
  </p>
</div>

---

## Overview

Waverless is a high-performance Serverless GPU task orchestration system designed for AI inference and training workloads, powered by [Wavespeed.ai](https://wavespeed.ai).

## Core Features

- 🚀 **Pull-based Architecture** - Workers actively pull tasks for better load balancing and fault tolerance
- 🔌 **RunPod Compatible** - Fully compatible with runpod-python SDK, no code modification needed
- ☸️ **Kubernetes Native** - Built-in K8s application management, supports deploying GPU workloads via API
- 📊 **Multi-Endpoint Routing** - Supports multiple independent task queues and worker pools
- 🌐 **Web Management Interface** - React-based modern UI for visual deployment and monitoring
- ⚡ **Auto Scaling** - Automatically adjusts worker count based on queue depth

## Quick Start

```bash
# Clone repository
git clone https://github.com/wavespeedai/waverless.git
cd waverless

# Deploy complete environment
./deploy.sh install

# Access Web UI
kubectl port-forward -n wavespeed svc/waverless-web-svc 3000:80
# Visit http://localhost:3000 (default: admin/admin)
```

**For detailed deployment, configuration, and usage**, see [User Guide](docs/USER_GUIDE.md).

## Architecture

```
┌─────────────┐         ┌──────────────────┐         ┌─────────────┐
│   Client    │ submit  │   Waverless      │  pull   │   Worker    │
│             ├────────>│   API Server     │<────────┤  (RunPod)   │
│  (V1 API)   │         │                  │         │ Endpoint: A │
└─────────────┘         │  - Task Queue    │         └─────────────┘
                        │  - Worker Mgmt   │
┌─────────────┐  API    │  - K8s Manager   │         ┌─────────────┐
│  Web UI     │ Request │                  │  pull   │   Worker    │
│(React+Nginx)├────────>│  Redis + MySQL   │<────────┤  (RunPod)   │
│             │         │                  │         │ Endpoint: B │
└─────────────┘         └──────────────────┘         └─────────────┘
```

**See [System Architecture](docs/ARCHITECTURE.md) for detailed design.**

## API Usage

Waverless provides RunPod-compatible V1/V2 APIs and K8s management APIs.

**Quick Example**:
```bash
# Submit task
curl -X POST http://localhost:8080/v1/wan22/run \
  -H "Content-Type: application/json" \
  -d '{"input": {"prompt": "a beautiful landscape"}}'

# Query status
curl http://localhost:8080/v1/status/{task_id}
```

**See [User Guide](docs/USER_GUIDE.md) for complete API documentation and usage examples.**

## Documentation

Waverless documentation has been streamlined into 3 core documents:

| Document | Description | Audience |
|----------|-------------|----------|
| [User Guide](docs/USER_GUIDE.md) | Quick start, configuration, autoscaling, Web UI, and troubleshooting | Users, Operators |
| [Architecture](docs/ARCHITECTURE.md) | System architecture, components, data models, statistics, and GPU tracking | Architects, System Designers |
| [Developer Guide](docs/DEVELOPER_GUIDE.md) | Advanced topics, graceful shutdown, concurrency safety, task tracking internals | Developers, Contributors |

### Quick Links by Role

**New Users**: Start with [User Guide](docs/USER_GUIDE.md) → Quick Start section

**Operators**: [User Guide](docs/USER_GUIDE.md) → Configuration & Troubleshooting sections

**Developers**: [Architecture](docs/ARCHITECTURE.md) → [Developer Guide](docs/DEVELOPER_GUIDE.md)

**Architects**: [Architecture](docs/ARCHITECTURE.md) for complete system design

## License

MIT License

## Contact

- GitHub: https://github.com/wavespeedai/waverless
- Issues: https://github.com/wavespeedai/waverless/issues
