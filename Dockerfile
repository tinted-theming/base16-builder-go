FROM debian:buster-slim
COPY entrypoint.sh /bin
COPY base16-builder-go /bin
ENTRYPOINT ["/bin/entrypoint.sh"]
