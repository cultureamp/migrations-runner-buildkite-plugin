ARG BUILDER_IMAGE=public.ecr.aws/docker/library/golang:1.24.5-alpine
ARG DISTROLESS_IMAGE=gcr.io/distroless/static

FROM ${BUILDER_IMAGE} as builder

ARG NETSKOPE_CERT
RUN if [ "${NETSKOPE_CERT}z" != "z" ];  then \
      echo "Installing Netskope MitM certificates" && \
      mkdir -p /usr/local/share/ca-certificates || true && \
      echo "${NETSKOPE_CERT}" >> /etc/ssl/certs/ca-certificates.crt && \
      echo "${NETSKOPE_CERT}" >> /usr/local/share/ca-certificates/netskope.crt && \
      apk --no-cache add ca-certificates && \
      update-ca-certificates; \
      fi

# Ensure ca-certficates are up to date
RUN update-ca-certificates

WORKDIR /src

COPY go.mod .

RUN go mod download
RUN go mod verify

COPY . ./

# build as a static binary without debug symbols
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags='-w -s -extldflags "-static"' -a \
      -o /dist/migrations-runner .

# runtime image using static distroless base
# using static nonroot image
# user:group is nobody:nobody, uid:gid = 65534:65534
FROM ${DISTROLESS_IMAGE}

COPY --from=builder /dist/migrations-runner /migrations-runner

ENTRYPOINT ["/migrations-runner"]
