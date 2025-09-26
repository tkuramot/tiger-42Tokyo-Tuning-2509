#!/bin/bash

JOB_ID=$1

if [ -z "$JOB_ID" ]; then
  echo -e "Usage: $0 <job_id>"
  exit 1
fi

PREVIOUS_REMAINING=""

while true; do
    RESPONSE=$(curl -s -G "https://intermediate.ftt2508.dabaas.net/get_status" --data-urlencode "jobId=$JOB_ID")

    STATUS=$(echo "$RESPONSE" | jq -r '.status')
    MESSAGE=$(echo "$RESPONSE" | jq -r '.message')

    DOTS+="."
    # "." を順に増やして表示し、3つ以上になったらリセットする
    if [ ${#DOTS} -gt 3 ]; then
        DOTS="."
    fi

    case "$STATUS" in
        "waiting")
            REMAINING=$(echo "$RESPONSE" | jq -r '.remaining')
            
            printf "現在キューイング中ですのでしばらくお待ち下さい。詳細は下記のコマンドで確認できます%-5s\n" "$DOTS"
            printf '%s\n' 'curl -N --http1.1 -H "Accept: text/event-stream" "https://intermediate.ftt2509.dabaas.net/waiting-stream"'
            ;;
        "running")
            PROGRESS=$(echo "$RESPONSE" | jq -r '.progress')
            if [ "$PROGRESS" -eq 100 ]; then
                printf "\r\033[Kスコアを計算中です%-5s" "$DOTS"
            else
                printf "\r\033[K負荷試験実行中です%-5s (%d%%)" "$DOTS" "$PROGRESS"
            fi
            ;;
        "finished")
            RESPONSE_SCORE=$(curl -s -G "https://intermediate.ftt2508.dabaas.net/score/$JOB_ID")

            TEST_ID=$(echo "$RESPONSE_SCORE" | jq -r '.test_id')
            SCORE=$(echo "$RESPONSE_SCORE" | jq -r '.score')

            echo -e "\n\n===================================================\n\n"
            echo -e "負荷試験が完了しました！！！"
            echo -e "あなたのスコア: $SCORE"
            echo -e "テストID: $TEST_ID"
            echo -e "\n\n===================================================\n\n"
            break
            ;;
        "failed")
            MESSAGE=$(echo "$RESPONSE" | jq -r '.message')
            FILE_KEY=$(echo "$RESPONSE" | jq -r '.fileKey')
            LOG=$(echo "$RESPONSE" | jq -r '.log')
            RAW_DATA=$(echo "$RESPONSE" | jq -r '.rawData')

            echo $LOG > ./benchmarker/logs/$FILE_KEY.json
            echo $RAW_DATA > ./benchmarker/scores/raw-data-$FILE_KEY.json

            echo -e "\n\n===================================================\n\n"
            echo -e "負荷試験が失敗しました。メンターに報告してください。"
            echo -e "ファイルキー：$FILE_KEY"
            echo -e $MESSAGE
            echo -e "\n\n===================================================\n\n"
            break
            ;;
        *)
            echo -e "\n\n===================================================\n\n"
            echo -e "不明なステータスです。メンターに報告してください。"
            echo -e $STATUS
            echo -e "\n\n===================================================\n\n"
            break
            ;;
    esac

    sleep 10
done
