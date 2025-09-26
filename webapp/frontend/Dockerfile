# ===== Runtime（最終ステージ：production）=====
FROM 42tokyo2508.azurecr.io/frontend:base AS production

WORKDIR /usr/src/frontend

COPY package.json yarn.lock ./

# production=false で next 等を必ずインストール（build 時に必要なパッケージが devDependencies にある場合があるため）
RUN yarn install --frozen-lockfile --production=false

# ソースコードをコピー
COPY . .

# Next ビルド
RUN yarn build

EXPOSE 3000