<div align="center">
  <a href="https://pulsewatch.dev" target="_blank">
    <img src="assets/Logo.png" alt="Netra Uptime Logo" width="84" />
  </a>
  <h1 style="border:none; margin-top: 10px; margin-bottom: 6px;">Netra Uptime</h1>
  <p><strong>Modern, Lightweight, Self-Hosted Uptime & Status Monitoring</strong></p>

  <p>
    <a href="https://github.com/Xuerns/PulseWatch/actions/workflows/ci.yml"><img src="https://github.com/Xuerns/PulseWatch/actions/workflows/ci.yml/badge.svg" alt="CI Status"></a>
    <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white" alt="Go Version"></a>
    <a href="https://kit.svelte.dev/"><img src="https://img.shields.io/badge/SvelteKit-5-FF3E00?logo=svelte&logoColor=white" alt="SvelteKit 5"></a>
    <a href="https://github.com/Xuerns/PulseWatch/pkgs/container/pulsewatch"><img src="https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white" alt="Docker Ready"></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/License-AGPL_3.0-blue.svg" alt="License: AGPL-3.0"></a>
  </p>

  <p>
    <a href="https://pulsewatch.dev/docs"><strong>Documentation</strong></a> •
    <a href="#quick-start"><strong>Quick Start</strong></a> •
    <a href="#key-features"><strong>Key Features</strong></a> •
    <a href="#contributing"><strong>Contributing</strong></a> •
    <a href="https://github.com/Xuerns/Netra-Uptime/issues"><strong>Report Issue</strong></a>
  </p>
</div>


## Overview

**Netra Uptime** is a high-performance, open-source uptime monitoring and status page platform designed for self-hosting. Built with a concurrent Go backend worker pool and a reactive SvelteKit interface, Netra Uptime provides low-latency telemetry, automated incident lifecycle tracking, and customizable alert rules.

## Quick Start

You can launch Netra Uptime in less than a minute using either the official CLI runner or Docker Compose.

### Option 1: Instant Run via CLI (Zero-Config)

Run Netra Uptime immediately without cloning using your favorite package runner:

```bash
# npx
npx netra-uptime start
```

### Option 2: Docker Compose (Self-Hosting)

Download the pre-configured production Docker Compose file and start the services:

```bash
# Download docker-compose.prod.yml
curl -fsSL https://raw.githubusercontent.com/Xuerns/Netra-Uptime/main/docker-compose.prod.yml -o docker-compose.yml

# Start containers in detached mode
docker compose up -d
```

Once running, access the dashboard at: **[http://localhost:8080](http://localhost:8080)**.

## Key Features

- **Multi-Protocol Monitoring**: Check uptime via HTTP/HTTPS, TCP ports, and ICMP Ping.
- **SSL Certificate & Domain Expiry**: Proactive tracking of SSL certificate validity and WHOIS/RDAP domain expiration.
- **Concurrent Worker Pool**: Non-blocking scheduler powered by Go goroutines and worker channels.
- **Manual Trigger ("Check Now")**: Run on-demand diagnostic checks directly from the monitor detail dashboard without waiting for cron intervals.
- **Custom Alert Rules & Multi-Channel Notifications**: Fine-grained breach thresholds (consecutive failures, response time limits) with notification dispatchers (Email/SMTP, Slack, Discord, Webhooks).
- **Public & Private Status Pages**: Branded status pages with uptime history bars, maintenance schedules, component grouping, and embeddable SVG badges.
- **Automated Incident Management**: Auto-generation of incident records on downtime, collaborative incident triage updates, and automatic resolution when monitors recover.
- **Multi-Tenancy & Team Roles**: Tenant isolation per organization with role-based access control (`owner`, `admin`, `member`) and email invitations.
- **System Health Checks**: Built-in `/healthz` and `/api/v1/health` endpoints verifying PostgreSQL and Redis connectivity for Docker and Kubernetes probes.
- **Public REST API & API Keys**: Secure external automation via rate-limited API keys.

## Contributing

Contributions are always welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on setting up your development environment and submitting pull requests.

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0) - see the [LICENSE](LICENSE) file for details.