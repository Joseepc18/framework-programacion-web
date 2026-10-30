#!/usr/bin/env bash
# Carga los datos de demostración de las Fases 1 y 2 llamando a la API (ver scripts/datos-demo/main.go).
# Uso: API_URL=http://localhost:3001 ./scripts/datos-demo.sh   (o contra la API desplegada)
set -euo pipefail
cd "$(dirname "$0")/.."
exec go run ./scripts/datos-demo
