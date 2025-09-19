FROM quay.io/scylladb/scylla-operator-images:golang-1.25 AS builder
WORKDIR /go/src/github.com/scylladb/scylla-operator
COPY . .
RUN make build --warn-undefined-variables

FROM quay.io/scylladb/scylla-operator-images:base-ubi-9.6-minimal

LABEL org.opencontainers.image.title="Scylla Operator" \
      org.opencontainers.image.description="ScyllaDB Operator for Kubernetes" \
      org.opencontainers.image.authors="ScyllaDB Operator Team" \
      org.opencontainers.image.source="https://github.com/scylladb/scylla-operator/" \
      org.opencontainers.image.documentation="https://operator.docs.scylladb.com" \
      org.opencontainers.image.url="https://hub.docker.com/r/scylladb/scylla-operator" \
      org.opencontainers.image.vendor="ScyllaDB"

RUN microdnf install -y procps-ng && \
    microdnf clean all && \
    rm -rf /var/cache/dnf/*

COPY --from=builder /go/src/github.com/scylladb/scylla-operator/scylla-operator /usr/bin/
COPY --from=builder /go/src/github.com/scylladb/scylla-operator/scylla-operator-tests /usr/bin/
ENTRYPOINT ["/usr/bin/scylla-operator"]
