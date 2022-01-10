#!/usr/bin/env sh
set -e

# TODO This file is here due to api-infra not supporting Viper entirely yet. Once the following PR
#      is merged, this can be removed: https://github.com/fluidshare/api-infra/pull/75
if [ -f /vault/secrets/.env ]; then
  echo "==> Loading environment variables from /vault/secrets/.env"
  source /vault/secrets/.env
else
  echo "==> Vault .env file not present, skipping."
fi

exec "$@"
