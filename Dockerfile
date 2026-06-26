ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

FROM alpine:3.23 AS prep
RUN apk add --no-cache ca-certificates tzdata \
  && echo "user:x:10001:10001::/nonexistent:/sbin/nologin" > /etc/passwd.custom \
  && echo "user:x:10001:" > /etc/group.custom

FROM scratch
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

COPY --from=prep /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=prep /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=prep /etc/passwd.custom /etc/passwd
COPY --from=prep /etc/group.custom /etc/group

COPY ${TARGETOS}/${TARGETARCH}${TARGETVARIANT:+/$TARGETVARIANT}/aniliberty-indexer /aniliberty-indexer

USER user:user
EXPOSE 3649
ENTRYPOINT ["/aniliberty-indexer"]
