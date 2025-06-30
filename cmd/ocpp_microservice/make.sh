#!/usr/bin/env bash
#set -euo pipefail

REMOTE_USER="root"
REMOTE_HOST="31.172.64.145"
REMOTE_PATH="/home"
REMOTE_PASS="O76OsZIPOzvznqT2"

echo 'Building binary'
if sudo GOOS=linux GOARCH=amd64 go build -o ocpp main.go; then
    echo 'Build succeeded'
else
    echo 'Build failed'; exit 1
fi

cp ocpp /Users/antondamashkevich/Programming/golang/chargerdeveloper/allmicroservices/

echo 'stopping previous process on remote'
sshpass -p "${REMOTE_PASS}" ssh "${REMOTE_USER}@${REMOTE_HOST}" \
  "if pkill -f '${REMOTE_PATH}/ocpp'; then echo 'Killed remote process'; else echo 'No running process found'; fi"
echo 'remote process kill step completed'

# Отправка бинаря на удаленный сервер
echo 'begin upload ocpp binary to remote server'
sshpass -p "${REMOTE_PASS}" scp ocpp "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_PATH}"
echo 'end upload ocpp binary to remote server'

# Запуск приложения на сервере в фоне
sshpass -p "${REMOTE_PASS}" ssh "${REMOTE_USER}@${REMOTE_HOST}" "chmod +x ${REMOTE_PATH}/ocpp && ${REMOTE_PATH}/ocpp "

#export PATH="$PATH:$HOME/go/bin" && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/proto/control/control.proto
