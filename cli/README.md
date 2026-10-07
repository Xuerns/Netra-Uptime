# Pulsewatch

Official CLI to quickly initialize, run, and manage **PulseWatch** self-hosted uptime monitoring and status pages.

## Quick Start

### 1. Instant Run (No installation required)

You can run PulseWatch immediately using your preferred package runner:

```bash
# npm / npx
npx pulsewatch start

# pnpm
pnpm dlx pulsewatch start

# yarn
yarn dlx pulsewatch start

# bun
bunx pulsewatch start
```

### 2. Global Installation (Optional)

If you prefer having the `pulsewatch` command permanently available:

```bash
# npm
npm install -g pulsewatch

# pnpm
pnpm add -g pulsewatch

# yarn
yarn global add pulsewatch

# bun
bun add -g pulsewatch
```

Then run anywhere:

```bash
pulsewatch start
```

PulseWatch web dashboard will be available at: **http://localhost:8080**

---

## Commands

| Command | Description |
| :--- | :--- |
| `pulsewatch init` | Generate `.env` and `docker-compose.yml` with secure random credentials. |
| `pulsewatch start` | Initialize (if not already done) and start containers (`docker compose up -d`). |
| `pulsewatch stop` | Stop and remove running containers (`docker compose down`). |
| `pulsewatch status` | Show status and health checks of PulseWatch, Postgres, and Redis (`docker compose ps`). |
| `pulsewatch logs` | Stream live container logs (`docker compose logs -f`). |
| `pulsewatch update` | Pull the latest Docker images and restart containers. |
| `pulsewatch help` | Display available commands and options. |

---

## Options

### Custom Port
By default, PulseWatch runs on port `8080`. You can change the port with `--port`:

```bash
npx pulsewatch start --port 3000
```

### Custom Directory
By default, configuration files are stored globally in `~/.pulsewatch` so you can manage PulseWatch from any folder. If you want to store configuration in the current folder or a specific directory:

```bash
# In the current directory
npx pulsewatch start --dir .

# Or in a specific path
npx pulsewatch start --dir /opt/pulsewatch
```

### Force Overwrite
To re-generate `.env` and `docker-compose.yml` (overwriting existing files):

```bash
npx pulsewatch init --force
```

---

## Requirements

- **Docker** and **Docker Compose** plugin installed.
- **Node.js** >= 18.

---

## License

GNU AGPLv3 © [xuerns](https://github.com/xuerns)
