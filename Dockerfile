FROM alpine AS dependencies

# terraform
RUN apk add --no-cache gnupg

WORKDIR /tmp

ARG HASHICORP_PGP_FINGERPRINT=C874011F0AB405110D02105534365D9472D7468F
ARG TERRAFORM_PLATFORM=linux_amd64
ARG TERRAFORM_VERSION=1.5.7

ADD https://keybase.io/hashicorp/pgp_keys.asc?fingerprint=${HASHICORP_PGP_FINGERPRINT} hashicorp.asc
ADD https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_SHA256SUMS .
ADD https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_SHA256SUMS.sig .
ADD https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_${TERRAFORM_PLATFORM}.zip .

RUN gpg --import hashicorp.asc
RUN gpg --verify terraform_${TERRAFORM_VERSION}_SHA256SUMS.sig terraform_${TERRAFORM_VERSION}_SHA256SUMS
RUN grep ${TERRAFORM_PLATFORM}.zip terraform_${TERRAFORM_VERSION}_SHA256SUMS | sha256sum -cs
RUN unzip terraform_${TERRAFORM_VERSION}_${TERRAFORM_PLATFORM}.zip

# git
RUN apk add --no-cache git

WORKDIR /usr/libexec/git-core
RUN rm -rf mergetools
RUN find . -type l -exec sh -c '[[ $1 != ./git-remote-https && $1 != ./git-pull ]] && unlink $1' _ {} \;
RUN find . -type f -exec sh -c '[[ $1 != ./git-remote-http ]] && rm -f $1' _ {} \;

WORKDIR /usr/lib
RUN rm -rf mdev modules-load.d ossl-modules engines-3
RUN find . -type l -exec sh -c '[[ $(echo $1 | grep -Ev "pcre|curl|http|libidn2|libbrotlidec|libunistring|libbrotlicommon") ]] && unlink $1' _ {} \;
RUN find . -type f -exec sh -c '[[ $(echo $1 | grep -Ev "pcre|curl|http|libidn2|libbrotlidec|libunistring|libbrotlicommon") ]] && rm -f $1' _ {} \;

WORKDIR /lib
RUN rm -rf apk firmware mdev sysctl.d modules-load.d libapk*

# terrasync
FROM golang:1.21.1-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
ADD src ./src

RUN CGO_ENABLED=0 GOOS=linux go build -o /terrasync

# Create non-root user for scratch final image
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/terraform" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "1000" \
    terrasync

# FROM builder AS tester
# RUN go test -v ./...

# Final clean image
FROM scratch

# terraform
COPY --from=dependencies /tmp/terraform /usr/bin/terraform

# git
COPY --from=dependencies /lib /lib
COPY --from=dependencies /usr/lib /usr/lib
COPY --from=dependencies /usr/libexec/git-core /usr/libexec/git-core
COPY --from=dependencies /usr/bin/git /usr/bin/git
COPY --from=dependencies /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# terrasync
COPY --from=builder /terrasync /terrasync

# Run
ENTRYPOINT ["/terrasync"]
