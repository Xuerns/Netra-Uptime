# Netra Uptime CLI

Official CLI to quickly initialize, run, and manage **Netra Uptime** self-hosted uptime monitoring and status pages.

## Quick Start

### 1. Instant Run (No installation required)

You can run Netra Uptime immediately using your preferred package runner:

```bash
# npm / npx
npx netra-uptime start

# pnpm
pnpm dlx netra-uptime start

# yarn
yarn dlx netra-uptime start

# bun
bunx netra-uptime start
```

### 2. Global Installation (Optional)

If you prefer having the `netra-uptime` command permanently available:

```bash
# npm
npm install -g netra-uptime

# pnpm
pnpm add -g netra-uptime

# yarn
yarn global add netra-uptime

# bun
bun add -g netra-uptime
```

Then run anywhere:

```bash
netra-uptime start
```

Netra Uptime web dashboard will be available at: **http://localhost:27346**

---

## Commands

| Command | Description |
| :--- | :--- |
| `netra-uptime init` | Generate `.env` and `docker-compose.yml` with secure random credentials. |
| `netra-uptime start` | Initialize (if not already done) and start containers (`docker compose up -d`). |
| `netra-uptime stop` | Stop and remove running containers (`docker compose down`). |
| `netra-uptime status` | Show status and health checks of Netra Uptime and Postgres (`docker compose ps`). |
| `netra-uptime credentials` | Display database connection credentials from `.env` (alias: `creds`). |
| `netra-uptime logs` | Stream live container logs (`docker compose logs -f`). |
| `netra-uptime update` | Pull the latest Docker images and restart containers. |
| `netra-uptime help` | Display available commands and options. |

---

## Options

### Custom Port
By default, Netra Uptime runs on port `27346`. You can change the port with `--port`:

```bash
npx netra-uptime start --port 3000
```

### Custom Directory
By default, configuration files are stored globally in `~/.netra-uptime` so you can manage Netra Uptime from any folder. If you want to store configuration in the current folder or a specific directory:

```bash
# In the current directory
npx netra-uptime start --dir .

# Or in a specific path
npx netra-uptime start --dir /opt/netra-uptime
```

### Force Overwrite
To re-generate `.env` and `docker-compose.yml` (overwriting existing files):

```bash
npx netra-uptime init --force
```

---

## Requirements

- **Docker** and **Docker Compose** plugin installed.
- **Node.js** >= 18.

---

## License

GNU AGPLv3 © [xuerns](https://github.com/xuerns)
