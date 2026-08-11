set -euo pipefail

GO_DIR="./backend"        
GO_MAIN="./cmd/server/"              
FRONTEND_DIR="./frontend" 
COMPOSE_FILE="docker-compose.yml"

PIDS=()

cleanup() {
  echo ""
  echo "Stop All Process..."
  for pid in "${PIDS[@]:-}"; do
    if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
    fi
  done
  echo "Stop Docker Compose"
  docker compose -f "$COMPOSE_FILE" down
  exit 0
}

trap cleanup SIGINT SIGTERM

echo "Run Docker Compose"
docker compose -f "$COMPOSE_FILE" up postgres redis -d 

echo "Run Backend..."
(
  cd "$GO_DIR"
  go run "$GO_MAIN"
) &
PIDS+=($!)

echo "Run Frontend..."
(
  cd "$FRONTEND_DIR"
  pnpm run dev --open
) &
PIDS+=($!)

echo "Successfully Run All"

wait