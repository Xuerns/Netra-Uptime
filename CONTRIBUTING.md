# Contributing to Netra Uptime

Thank you for your interest in contributing to **Netra Uptime**! We appreciate all contributions, whether it is adding new features, fixing bugs, improving documentation, or suggesting enhancements.

---

## Getting Started

### 1. Fork & Clone the Repository

1. Fork the official repository on GitHub: [Xuerns/Netra-Uptime](https://github.com/Xuerns/Netra-Uptime).
2. Clone your forked repository to your local machine:
   ```bash
   git clone https://github.com/<your-username>/Netra-Uptime.git
   cd Netra-Uptime
   ```
3. Create a descriptive topic branch for your work:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/issue-description
   ```

---

## Development Setup

You can set up and run your development environment using either of the following two options:

### Option A: Docker Compose (All-in-One)

Build and spin up the complete application stack (app container and PostgreSQL) simultaneously:

```bash
docker compose up --build
```

The web interface will be accessible at: `http://localhost:27346`.

---

### Option B: Local Development Setup

Run the required infrastructure service (PostgreSQL) inside Docker, and run the backend and frontend servers locally on your host machine for faster live-reloading:

```bash
# 1. Start development PostgreSQL container in background
docker compose up -d postgres

# 2. Run the Go backend server (terminal 1)
cd server
go run ./cmd/netra-uptime

# 3. Run the SvelteKit frontend server (terminal 2)
cd client
pnpm dev
```

> **Requirements for local development:**
> - Go 1.24+
> - Node.js 20+
> - pnpm 9+
> - Docker & Docker Compose

---

## Code Quality & Validation

Before submitting your changes, please ensure that your code passes all formatting, linting, and tests:

### Backend (Go)
From the `server/` directory:
```bash
go fmt ./...
go vet ./...
gosec ./...
go test ./...
go build ./...
```

### Frontend (SvelteKit)
From the `client/` directory:
```bash
pnpm lint
pnpm run check
pnpm run format
pnpm run build
```

---

## Submitting a Pull Request

When your changes are tested and ready:

1. Commit your changes with clear, descriptive commit messages:
   ```bash
   git add .
   git commit -m "feat: add webhook retry logic"
   git push origin feature/your-feature-name
   ```
2. Open a Pull Request from your branch against the `main` branch of [Xuerns/Netra-Uptime](https://github.com/Xuerns/Netra-Uptime).

### Pull Request Format

To help reviewers understand and review your contribution efficiently, please structure your Pull Request description using the following template:

```markdown
### Title
[Category]: Brief description of what is being added or fixed
Example: "feat: add Telegram notification dispatcher" or "fix: resolve ICMP packet timeout"

### Description

#### 1. What is this feature / change?
A concise explanation of what this feature, improvement, or bug fix is.

#### 2. How does it work?
A technical overview of the workflow, code changes, and components involved.

#### 3. Why is this feature / change important?
The motivation and benefits this brings to Netra Uptime users and project stability.
```

---

## Reporting Issues

If you find a bug or have a feature request, please search existing issues first. If none exists, open a new issue on [GitHub Issues](https://github.com/Xuerns/Netra-Uptime/issues) with detailed steps to reproduce.
