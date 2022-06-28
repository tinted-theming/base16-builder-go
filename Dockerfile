FROM debian:buster-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY entrypoint.sh /bin
COPY base16-builder-go /bin
ENTRYPOINT ["/bin/entrypoint.sh"]
