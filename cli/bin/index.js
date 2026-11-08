#!/usr/bin/env node

import { existsSync, mkdirSync, writeFileSync, readFileSync } from 'node:fs';
import { homedir } from 'node:os';
import { join, resolve } from 'node:path';
import { randomBytes } from 'node:crypto';
import { spawnSync } from 'node:child_process';
import { generateEnv, generateCompose } from '../src/templates.js';

const HELP = `
Netra Uptime CLI

Usage:
  netra-uptime <command> [options]

Commands:
  init      Generate .env and docker-compose.yml (random secrets)
  start     Init (if needed) and run: docker compose up -d
  stop      Run: docker compose down
  status    Show container status
  logs      Follow logs
  credentials  Display database credentials from .env (alias: creds)
  update    Pull the latest image and restart
  help      Show this help

Options:
  --port <n>   Host port for the web UI (default: 27346)
  --dir <path> Install directory (default: ~/.netra-uptime)
  --force      Overwrite existing .env / docker-compose.yml (init)
`;

function parseArgs(argv) {
  const opts = { _: [] };
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a === '--force') opts.force = true;
    else if (a === '--port') opts.port = argv[++i];
    else if (a === '--dir') opts.dir = argv[++i];
    else if (a === '-h' || a === '--help') opts._.push('help');
    else opts._.push(a);
  }
  return opts;
}

const secret = (bytes = 16) => randomBytes(bytes).toString('hex');

function run(cmd, args, cwd) {
  const r = spawnSync(cmd, args, { cwd, stdio: 'inherit' });
  if (r.error) {
    console.error(`✖ Failed to run "${cmd}": ${r.error.message}`);
    process.exit(1);
  }
  if (r.status !== 0) process.exit(r.status ?? 1);
}

function ensureDocker() {
  const r = spawnSync('docker', ['compose', 'version'], { stdio: 'ignore' });
  if (r.error || r.status !== 0) {
    console.error('✖ Docker with the Compose plugin is required. Install it from https://docs.docker.com/get-docker/');
    process.exit(1);
  }
}

function init(dir, opts) {
  const port = Number(opts.port || 27346);
  if (!Number.isInteger(port) || port < 1 || port > 65535) {
    console.error('✖ Invalid --port value');
    process.exit(1);
  }

  mkdirSync(dir, { recursive: true });
  const envPath = join(dir, '.env');
  const composePath = join(dir, 'docker-compose.yml');

  if (existsSync(envPath) && !opts.force) {
    console.log(`• .env already exists, keeping it (${envPath})`);
  } else {
    writeFileSync(
      envPath,
      generateEnv({ port, jwtSecret: secret(32), dbPassword: secret() }),
      { mode: 0o600 }
    );
    console.log(`✔ Created ${envPath}`);
  }

  if (existsSync(composePath) && !opts.force) {
    console.log(`• docker-compose.yml already exists, keeping it`);
  } else {
    writeFileSync(composePath, generateCompose());
    console.log(`✔ Created ${composePath}`);
  }

  return { dir, port };
}

function getPortFromEnv(dir, fallback = '27346') {
  try {
    const envContent = readFileSync(join(dir, '.env'), 'utf-8');
    const match = envContent.match(/^HOST_PORT=(\d+)/m);
    if (match) return match[1];
  } catch {}
  return fallback;
}

function readCredentials(dir) {
  const envPath = join(dir, '.env');
  if (!existsSync(envPath)) return null;

  try {
    const content = readFileSync(envPath, 'utf-8');
    const parse = (key) => {
      const match = content.match(new RegExp(`^${key}=(.*)$`, 'm'));
      return match ? match[1].trim() : '';
    };

    return {
      envPath,
      hostPort: parse('HOST_PORT') || '27346',
      dbUser: parse('POSTGRES_USER') || 'netra',
      dbPassword: parse('POSTGRES_PASSWORD'),
      dbName: parse('POSTGRES_DB') || 'netra_uptime',
    };
  } catch {
    return null;
  }
}

function formatCredentialsSummary(creds) {
  if (!creds) {
    return '✖ No .env file found. Run "netra-uptime init" first.';
  }

  return `[Configuration & Credentials]
  Config file : ${creds.envPath}
  PostgreSQL  : localhost:5432
    • Database: ${creds.dbName}
    • User    : ${creds.dbUser}
    • Password: ${creds.dbPassword || '(not set)'}`;
}

function main() {
  const opts = parseArgs(process.argv.slice(2));
  const command = opts._[0] || 'help';
  const dir = resolve(opts.dir || join(homedir(), '.netra-uptime'));
  const dirFlag = opts.dir ? ` --dir ${opts.dir}` : '';

  switch (command) {
    case 'init': {
      const { port } = init(dir, opts);
      const creds = readCredentials(dir);
      console.log(`
⪢ Netra Uptime initialized successfully! ⪡

${formatCredentialsSummary(creds)}

Next steps:
  1. Start Netra Uptime:
     netra-uptime start${dirFlag}

  2. View credentials anytime:
     netra-uptime credentials${dirFlag}

  3. Check container status:
     netra-uptime status${dirFlag}

  4. Need more info? Run:
     netra-uptime help
`);
      break;
    }
    case 'start': {
      ensureDocker();
      init(dir, opts);
      run('docker', ['compose', 'up', '-d'], dir);
      const activePort = opts.port || getPortFromEnv(dir, '27346');
      const creds = readCredentials(dir);
      console.log(`
⫸ Netra Uptime is up and running! ⫷
   Web Dashboard: http://localhost:${activePort}

${formatCredentialsSummary(creds)}

Useful commands:
  • netra-uptime credentials${dirFlag} - View database credentials
  • netra-uptime status${dirFlag}      - Check container health
  • netra-uptime logs${dirFlag}        - Stream live logs
  • netra-uptime stop${dirFlag}        - Stop all containers
  • netra-uptime help              - View all commands
`);
      break;
    }
    case 'credentials':
    case 'creds': {
      const creds = readCredentials(dir);
      console.log(`
${formatCredentialsSummary(creds)}
`);
      break;
    }
    case 'stop':
      ensureDocker();
      run('docker', ['compose', 'down'], dir);
      break;
    case 'status':
      ensureDocker();
      run('docker', ['compose', 'ps'], dir);
      break;
    case 'logs':
      ensureDocker();
      run('docker', ['compose', 'logs', '-f', '--tail', '100'], dir);
      break;
    case 'update':
      ensureDocker();
      run('docker', ['compose', 'pull'], dir);
      run('docker', ['compose', 'up', '-d'], dir);
      break;
    case 'help':
      console.log(HELP);
      break;
    default:
      console.error(`✖ Unknown command: ${command}`);
      console.log(HELP);
      process.exit(1);
  }
}

main();
