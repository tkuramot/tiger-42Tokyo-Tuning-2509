FROM node:22-bullseye

RUN apt-get update && apt-get install -y --no-install-recommends \
    net-tools telnet less git ca-certificates tzdata python3 make g++ curl && rm -rf /var/lib/apt/lists/*
