# syntax=docker/dockerfile:1.7
FROM node:20.19

WORKDIR /usr/src/e2e

# 日本語レンダリング（任意）
RUN apt-get update && apt-get install -y --no-install-recommends \
    fonts-noto-cjk \
    && rm -rf /var/lib/apt/lists/*

# devDependencies も入れる
COPY package.json yarn.lock ./
RUN --mount=type=cache,target=/usr/local/share/.cache/yarn/v6 \
    yarn install --frozen-lockfile --production=false

# ブラウザを “両方” 同梱（←ここがポイント）
#   - chromium
#   - chromium-headless-shell  ← これが今のログで要求されているもの
RUN npx playwright install --with-deps chromium chromium-headless-shell

# テスト資材
COPY tests ./tests
COPY playwright.config.ts ./