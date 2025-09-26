#!/bin/bash

# ==================================
# 負荷試験スクリプト。
# ==================================
FILE_NAME=${1:-main}  # 引数があれば使い、なければ main

if [[ $HOSTNAME == benchmarker ]];
then
  LOG_FILE_PATH="./logs/${FILE_NAME}.json"
  RAW_DATA_FILE_PATH="./scores/raw-data-${FILE_NAME}.json"
  SUMMARY_FILE_PATH="./scores/summary-${FILE_NAME}.json"
else
  LOG_FILE_PATH="./logs/${FILE_NAME}.json"
  RAW_DATA_FILE_PATH="./scores/raw-data-${FILE_NAME}.json"
  SUMMARY_FILE_PATH="./scores/summary-${FILE_NAME}.json"
fi

# ディレクトリ確保
for f in "$LOG_FILE_PATH" "$RAW_DATA_FILE_PATH" "$SUMMARY_FILE_PATH"; do
  dir="$(dirname "$f")"
  [[ -d "$dir" ]] || mkdir -p "$dir"
done

# TARGET_IP は外から入ってくる想定（無ければ localhost）
TARGET_IP="${TARGET_IP:-127.0.0.1}"
echo "TARGET_IP: $TARGET_IP" 1>&2

# E2E経由のURL（https/localhostは好みで調整可）

if [[ $HOSTNAME == benchmarker ]];
then
  CLIENT_ORIGIN_URL="https://localhost"
  [[ -n "$TARGET_IP" ]] && CLIENT_ORIGIN_URL="https://${TARGET_IP}"
else
	CLIENT_ORIGIN_URL="http://tuning-nginx"
fi

echo "CLIENT_ORIGIN_URL: $CLIENT_ORIGIN_URL" 1>&2
echo "LOG_FILE_PATH: $LOG_FILE_PATH" 1>&2
echo "SUMMARY_FILE_PATH: $SUMMARY_FILE_PATH" 1>&2
echo "RAW_DATA_FILE_PATH: $RAW_DATA_FILE_PATH" 1>&2
echo "Running scenario: $FILE_NAME" 1>&2
k6 run --insecure-skip-tls-verify \
  --out json="$LOG_FILE_PATH" \
  --summary-export "$SUMMARY_FILE_PATH" \
  scenarios/main.js \
  -e CLIENT_ORIGIN_URL="$CLIENT_ORIGIN_URL" \
  -e RAW_DATA_FILE_PATH="$RAW_DATA_FILE_PATH" \
  1>&2
