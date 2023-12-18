FROM alpine:3.19

ARG HOME=/app

RUN apk add --update --no-cache ca-certificates

RUN addgroup -g 1616 -S upgrademanager \
    && adduser --home ${HOME} -u 1616 -S upgrademanager -G upgrademanager \
    && mkdir -p /app \
    && chown upgrademanager: -R /app

USER 1616

WORKDIR ${HOME}

COPY upgrade-manager /app/

EXPOSE 10000

ENTRYPOINT ["/app/upgrade-manager"]
CMD ["start"]