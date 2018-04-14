ARG BASE_IMAGE=ubuntu:14.04
FROM ${BASE_IMAGE}

WORKDIR /service/go-sync-mongo
COPY bin/go-sync-mongo-linux-amd64 ./

ENTRYPOINT ["./go-sync-mongo-linux-amd64"]
