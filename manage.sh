#!/bin/bash

set -e

SERVICE_NAME="app"
PROJECT_NAME="ristorante-api"
CONTAINER_NAME="ristorante-api-db"


# Carica variabili da .env
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
else
  echo "⚠️  Nessun file .env trovato nella directory corrente."
  exit 1
fi

usage() {
  echo "Uso: ./manage.sh {start|stop|rebuild|reset|logs|shell|seed-check}"
  exit 1
}

start() {
  echo "🚀 Avvio dell'applicazione..."
  docker compose up -d
}

stop() {
  echo "🛑 Arresto dei container..."
  docker compose down
}

rebuild() {
  echo "🔧 Rebuild completo (senza eliminare i volumi)..."
  docker compose down
  docker compose build
  docker compose up -d
}

reset() {
  echo "♻️  Reset completo (volumi inclusi) e seeding..."
  docker compose down -v
  docker compose build
  docker compose up -d

  echo "⏳ Attesa dell'avvio del servizio..."
  sleep 5

  echo "📂 Verifica presenza del file di seed..."
  if docker exec "$CONTAINER_NAME" test -f /app/database/migrations/complete_seed.sql; then
    echo "✅ File di seed trovato!"
  else
    echo "❌ ATTENZIONE: complete_seed.sql mancante nel container!"
  fi
}

logs() {
  echo "📜 Log del container $CONTAINER"
  docker logs -f "$CONTAINER_NAME"
}

shell() {
  echo "🧮 Accesso alla shell del container..."
  docker exec -it "$CONTAINER_NAME" sh
}

seed_check() {
  echo "🔎 Verifica contenuto tabella 'ristorante' su DB '${DB_NAME}'..."

  docker exec -e PGPASSWORD="$DB_PASSWORD" "$CONTAINER_NAME" \
    psql -U "$DB_USER" -d "$DB_NAME" \
    -c "SELECT * FROM ristorante LIMIT 5;"
}

# Dispatcher
case "$1" in
  start) start ;;
  stop) stop ;;
  rebuild) rebuild ;;
  reset) reset ;;
  logs) logs ;;
  shell) shell ;;
  seed-check) seed_check ;;
  *) usage ;;
esac