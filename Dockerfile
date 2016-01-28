FROM gliderlabs/alpine

RUN mkdir -p /app/bin /app/etc /app/schemas /app/templates /app/logs

COPY bin/etcdrest /app/bin/etcdrest

COPY templates/print.tmpl /app/templates/print.tmpl

EXPOSE 8080
ENTRYPOINT ["/app/bin/etcdrest"]
