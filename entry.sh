#!/bin/bash

#==================================
# VM環境のみforkしたリポジトリをcloneし、初期化するスクリプト。
# 一度だけ実行可能。
#==================================

if [[ $HOSTNAME != app-* ]]; then
  echo "ローカル環境ではこのスクリプトは利用できません。"
  exit 1
fi

if [[ -e /.da/cloneUrl ]]; then
  echo "既にリポジトリをcloneしているため、処理を中断しました。"
  exit 1
fi

echo -n "forkしたリポジトリのURLを入力してください: "
read -r repoUrl

repoName=$(basename "$repoUrl" | sed -e 's/.git//')

if git clone "$repoUrl"; then
    if (cd "./${repoName}" && bash init.sh); then
        echo -n "$repoUrl" > /.da/cloneUrl
        echo "セットアップが完了しました。"
    else
        echo "初期化スクリプト（init.sh）の実行に失敗しました。"
        exit 1
    fi
else
    echo "リポジトリのcloneに失敗しました。URLが正しいかご確認ください。"
    exit 1
fi
