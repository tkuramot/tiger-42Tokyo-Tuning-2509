#!/bin/bash

# ==================================
# アプリのコンテナ再起動スクリプト。
# ==================================

# アプリのコンテナ再起動
echo "アプリのコンテナの再起動を開始します。"

# ネットワークの存在確認と作成
NETWORK_NAME="webapp-network"
if ! docker network ls | grep -q "$NETWORK_NAME"; then
    docker network create $NETWORK_NAME > /dev/null 2>&1
fi

if [[ $HOSTNAME == ftt2508-* ]]; then
    HOSTNAME=$HOSTNAME docker compose down --volumes --rmi local
    HOSTNAME=$HOSTNAME docker compose up --build -d
else
    echo "ローカル環境でのコンテナ再起動を開始します。"
    # init.sh実行時には実行しない
    docker compose -f docker-compose.local.yml down db backend --volumes --rmi local
	docker compose -f docker-compose.local.yml up --build -d
fi

if [ $? -ne 0 ]; then
    echo "コンテナの再起動に失敗しました。"
    exit 1
else
    echo "コンテナの再起動に成功しました。"
fi
