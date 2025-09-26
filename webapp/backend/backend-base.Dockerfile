FROM golang:1.25 

RUN apt-get update && apt-get install -y --no-install-recommends \
    telnet less ca-certificates tzdata curl git \
    && rm -rf /var/lib/apt/lists/*