# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-11-10

### Added

#### Monitoring Engine & Protocols
- **Multi-Protocol Health Checks**:
  - **HTTP/HTTPS**: Support for custom request methods (`GET`, `POST`, `PUT`, `PATCH`, `HEAD`), scalar expected HTTP status code validation (default: 200), custom HTTP headers (JSON key-value, max 25 headers / 4KB with CRLF injection and host override protections), request body payloads (up to 16KB), and sanitized error response snippets (printable UTF-8 text, max 512 bytes captured on status >= 400 for rapid root-cause diagnosis).
  - **TCP Port Monitoring**: Raw TCP connection handshakes for verifying services like databases, mail servers, and custom network daemons.
  - **ICMP Ping**: Cross-platform ping engine supporting Linux/macOS sockets and unprivileged native Windows Win32 `IcmpSendEcho` (`iphlpapi.dll`) with packet loss and round-trip time (RTT) tracking.
  - **SSL/TLS Certificate Tracking**: Automatic certificate chain inspection, expiration countdown, certificate authority verification, and proactive renewal alerts.
  - **Domain Expiry Tracking**: Integrated WHOIS and RDAP checks to monitor domain expiration dates and alert before domain lapses occur.
- **On-Demand Diagnostic Check ("Check Now")**: Run instant checks from the dashboard to diagnose issues without waiting for scheduled intervals.
- **Fleet Overview & Degraded States**: Real-time fleet health aggregation, degraded performance indicators, sparkline trendlines, and latency percentile tracking (p50, p90, p99).

#### Concurrency & Scheduling
- **Go Goroutine Scheduler & Worker Pool**: High-throughput concurrent worker pool with buffered channels and non-blocking scheduling.
- **Immediate Retry (False-Positive Suppression)**: Automatic 1x retry with a 2-second sleep delay upon transient connection errors, timeouts, 5xx server errors, or status code mismatches (SSRF errors excluded) prior to alert rule evaluation and metric persistence.
- **Scheduled Maintenance Suppression**: Dynamic worker pool integration (`syncAndCheckMaintenance`) automatically advancing maintenance lifecycle (`scheduled` -> `in_progress` -> `completed`), tagging checks with `status = 'maintenance'`, suppressing incident generation, muting notification dispatches, and excluding maintenance windows from SLA availability calculations.
- **Jittered Execution**: Anti-thundering herd jitter algorithm to stagger concurrent checks across target domains.
- **Strict Invariant Validation**: Enforced rule ensuring `Timeout < Interval` (minimum interval: 30s, maximum timeout: 30s) to protect network resources.
- **Unbiased SLA Rollups & Retention**: Dedicated background workers for hourly and daily metric rollups. High-latency checks with valid HTTP status (< 500) remain counted in `up_checks` (availability decoupled from performance degradation), timeouts are counted in `down_checks`, and latency aggregations (AVG/MIN/MAX with weighted averages) exclude failed, maintenance, and timeout samples to prevent statistical distortion.

#### Alert Rules & Notifications
- **Configurable Alert Rule Engine**: Multi-threshold evaluations including consecutive failure counts, latency thresholds (ms), and recovery confirmations.
- **Multi-Channel Dispatchers**: First-class notification dispatchers for Email (SMTP / STARTTLS), Slack webhooks, Discord webhooks, and generic custom JSON webhooks.
- **Alert Suppression & Cooldowns**: Intelligent alert dampening to prevent alert storms and notification fatigue during flapping conditions, plus full suppression during scheduled maintenance.
- **Alert Templates & Delivery Logs**: Reusable message formatting templates and comprehensive notification logs detailing delivery status and error diagnostics.

#### Incident Management Lifecycle
- **Automated Incident Triage**: Automatic incident creation upon threshold breach, with dynamic severity calculation (`Critical`, `Major`, `Minor`).
- **Response Snippet Diagnostics**: Storage and rendering of sanitized error response snippets (max 512 bytes) and triggered rule expressions directly on incident records for immediate root-cause identification.
- **Real Incident Downtime Duration**: Accurate downtime tracking calculated from the delta between incident trigger timestamp (`StartedAt`) and recovery timestamp (`ResolvedAt`).
- **Incident Workspace & Timeline**: Collaborative incident notes, investigation timeline entries, and status transitions (`Investigating`, `Identified`, `Monitoring`, `Resolved`).
- **Automatic Resolution**: Seamless auto-resolution and downtime duration calculation as soon as the target passes consecutive recovery checks.
- **Pagination & Filters**: Filter incidents by severity, state, monitor, and date ranges.

#### Status Pages & Public Telemetry
- **Public & Authenticated Status Pages**: Flexible visibility controls with custom subdomain and slug routing.
- **Component Groups & Monitor Nesting**: Organize complex architectures into structured component groups with individual operational statuses.
- **Uptime History & Metrics**: 90-day interactive uptime history bars, active incidents, and response time visualizations. Telemetry during active scheduled maintenance windows is excluded from 90-day uptime calculations to preserve accurate operational SLAs.
- **Scheduled Maintenance**: Announce planned maintenance windows with countdown banners, automated status overrides, and incident suppression.
- **Embeddable SVG Badges & JSON Feeds**: Dynamic badge endpoints (`/api/v1/status-pages/:slug/badge`) for GitHub repositories and public JSON feeds for external monitoring aggregators.
- **Branding & Theming**: Customizable page titles, descriptions, logos, and custom color accents.

#### Multi-Tenancy & Access Control (RBAC)
- **Tenant Isolation**: Complete data and operational isolation per organization across queries, cache, and WebSocket broadcasts.
- **Role-Based Access Control**: Multi-tier permission architecture (`owner`, `admin`, `member`, `viewer`) governing operational privileges and access boundaries.
- **Admin Fleet Management**: Extended monitor creation permissions (`POST /monitors`) to the `admin` role (`RequireRole("owner", "admin")`), allowing team administrators full autonomy over monitors, alert rules, notification channels, and status pages while reserving billing and organization ownership transfers to the `owner`.
- **Role-Based UI Gating & Dynamic Error Handling**: Centralized frontend permission guards (`currentUserRole`, `canManage`, `isOwner`) paired with dynamic API error handling (`handleApiError`) that preserves granular server error feedback (e.g., 403 Forbidden).
- **Team Invitations**: Secure tokenized email invitation workflow with expiration lifecycle and seat limits.
- **Audit Logging**: Comprehensive, immutable audit trail recording user and administrative activities with filtering and export capabilities.

#### Windows Desktop Standalone App
- **Native System Tray Mode**: Dedicated Windows executable (`cmd/netra-uptime-desktop`) built with `fyne.io/systray` running silently without a console window (`-H windowsgui`).
- **Single-Instance Mutex Guard**: Named Win32 mutex ensuring only one background instance runs; launching additional shortcuts brings the existing dashboard to focus.
- **System Tray Controls**: Quick tray menu featuring "Open Netra" and "Quit" (performs graceful teardown, SQLite WAL checkpoint, and worker flush).
- **Silent Background Boot**: CLI flags `--background` and `--no-browser` for silent startup via the Windows Startup folder.
- **Automated LAN Mode**: Automatic assignment of `ALLOW_PRIVATE_NETWORK_MONITORING=true` when running SQLite on loopback (`127.0.0.1`), enabling out-of-the-box local homelab, printer, and NAS monitoring.

#### Dual-Mode Database Architecture
- **PostgreSQL & Pure-Go SQLite**: Dynamic database engine factory supporting PostgreSQL (`pgx/v5`) for distributed enterprise deployments and pure-Go SQLite (`glebarez/sqlite`, zero CGO) for single-binary self-hosting.
- **Write-Ahead Logging (WAL)**: Optimized SQLite WAL mode with automatic checkpoints for reliable multi-threaded concurrency.
- **Dialect-Agnostic ANSI SQL**: Standardized queries with in-application percentile math ensuring 100% feature parity between PostgreSQL and SQLite.
- **UUID v4 Standardization**: Uniform UUID v4 primary keys across all 23 database models with automatic fallback generators.

#### Developer API & UI Experience
- **Public REST API & API Keys**: Cryptographically hashed SHA-256 API tokens with scoped permissions, rate limiting, and automated lifecycle binding to creator accounts.
- **Modern SvelteKit 5 Frontend**: Built with SvelteKit 5 Runes (`$state`, `$derived`, `$effect`, `$props`), Tailwind CSS, Lucide icons, responsive navigation, and cohesive dark/light mode themes.
- **Live Real-Time Telemetry**: Real-time dashboard synchronization powered by secure WebSockets.
- **CLI Runner**: Instant zero-config execution via `npx netra-uptime start`.

---

### Changed

- **High-Performance In-Memory Cache**: Migrated from external Redis (`go-redis/v9`) to an embedded, thread-safe Go in-memory cache (`pkg/cache/memory.go`) with TTL eviction and singleflight locking, eliminating network overhead for cache lookups, rate limiting, and OTP tokens.
- **Default Application Port**: Updated default HTTP port from `8080` to `27346` across Docker configs, CLI runner, server bindings, and client proxies to prevent port conflicts with common web servers.
- **Embedded Single-Binary Distribution**: Bundled compiled SvelteKit frontend assets directly into the Go binary (`server/internal/static/embed.go`), enabling true zero-dependency single-executable deployments.
- **100% Open-Source Transition**: Removed legacy commercial billing modules, subscription gates, and third-party SaaS hooks in favor of operator-controlled configuration limits via `.env` (`MAX_MONITORS_PER_ORG`, `MAX_STATUS_PAGES_PER_ORG`, etc.).
- **Shared Application Lifecycle**: Extracted server startup, worker initialization, and graceful shutdown orchestration into a unified `internal/app` package shared between the headless server and desktop GUI executables.

---

### Removed

- **External Redis Dependency**: Eliminated Redis requirement from Docker Compose and production deployment stacks, resulting in a significantly lighter footprint and simpler maintenance.
- **Legacy Payment & Billing Stack**: Removed Midtrans integrations, subscription database models, tier checks, and billing interfaces.
- **External OAuth Module**: Removed Google OAuth handlers and `golang.org/x/oauth2` dependencies in favor of self-contained local authentication and tokenized invite flows.
- **Legacy Environment Variables**: Deprecated and purged obsolete `APP_MODE` and `PUBLIC_APP_MODE` configurations.

---

### Security

- **Comprehensive SSRF Protection**:
  - Fail-closed IP dialer and DNS validation blocking private IP ranges (RFC 1918, RFC 4193), loopback addresses (`127.0.0.0/8`, `::1`), link-local addresses (`169.254.0.0/16`), and internal Docker container bridge subnets (`172.16.0.0/12`).
  - Dial-time DNS resolution enforcement mitigating DNS rebinding and TOCTOU vulnerabilities.
  - Restrictive HTTP redirect handling preventing redirect-based bypasses into internal infrastructure.
  - Strict opt-in environment flag (`ALLOW_PRIVATE_NETWORK_MONITORING`) required to probe local networks.
- **JWT Authentication Hardening**:
  - Rejection of default or placeholder JWT secrets (e.g., `your-super-secret-jwt-key`, `change-me`) at application startup in production mode.
  - Enforced `token_type` claim validation distinguishing between `access` and `refresh` tokens to prevent token-confusion attacks (CWE-287).
  - Automated credential lifecycle management with refresh token rotation and revocation.
- **Rate Limiting & Abuse Prevention**:
  - Sliding-window rate limiters across sensitive endpoints (authentication, password resets, OTP requests, and public APIs).
  - Target domain quota caps evaluated at eTLD+1 to prevent excessive target polling and distributed abuse.
  - Enforced minimum check interval floors (30s) to prevent unintentional Denial of Service (DoS) against monitored endpoints.
- **Input Sanitization & Injection Defense**:
  - Comprehensive HTML stripping and `NoHTML` validation across monitor targets, status page details, and notification payloads to block Stored XSS.
  - Parameterized ANSI SQL queries preventing SQL injection vulnerabilities.
  - URL scheme validation in non-Windows desktop launchers resolving command execution risks (CWE-78 / Gosec G204).
- **Tenant Isolation & Session Revocation**:
  - Strict organization ID scoping across all database lookups, preventing Insecure Direct Object References (IDOR).
  - Immediate revocation of API keys upon creator demotion or organization expulsion.
  - Active WebSocket session eviction when a user's organization membership is terminated.

[1.0.0]: https://github.com/Xuerns/Netra-Uptime/releases/tag/v1.0.0
