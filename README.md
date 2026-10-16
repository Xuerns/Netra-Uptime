<div align="center">
  <img src="assets/Logo.png" alt="Netra Uptime Logo" width="80" />
  <h1 style="border:none;">Netra Uptime</h1>
  <p><strong>Modern, Lightweight, Self-Hosted Uptime & Status Monitoring</strong></p>

  <a href="https://kit.svelte.dev/" target="_blank"><img src="https://img.shields.io/badge/SvelteKit-5-brightgreen?logo=svelte&logoColor=orange" alt="SvelteKit"></a>
  <a href="https://go.dev/" target="_blank"><img src="https://img.shields.io/badge/Go-1.24+-brightgreen?logo=go&logoColor=blue" alt="Go"></a>
  <a href="https://www.docker.com/" target="_blank"><img src="https://img.shields.io/badge/Docker-Ready-blue?logo=docker" alt="Docker"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-AGPL_3.0-blue.svg" alt="License: AGPL-3.0"></a>
</div>

## Overview

**Netra Uptime** is a high-performance, open-source uptime monitoring and status page platform designed for self-hosting. Built with a concurrent Go backend worker pool and a reactive SvelteKit interface, Netra Uptime provides low-latency telemetry, automated incident lifecycle tracking, and customizable alert rules.

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

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0) - see the [LICENSE](LICENSE) file for details.