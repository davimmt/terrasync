FROM alpine:3.10 AS terraform

RUN apk add --no-cache gnupg

ARG HASHICORP_PGP_FINGERPRINT=C874011F0AB405110D02105534365D9472D7468F
ARG TERRAFORM_PLATFORM=linux_amd64
ARG TERRAFORM_VERSION=1.5.7

WORKDIR /tmp

ADD https://keybase.io/hashicorp/pgp_keys.asc?fingerprint=${HASHICORP_PGP_FINGERPRINT} hashicorp.asc
ADD https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_SHA256SUMS .
ADD https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_SHA256SUMS.sig .
ADD https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_${TERRAFORM_PLATFORM}.zip .

RUN gpg --import hashicorp.asc
RUN gpg --verify terraform_${TERRAFORM_VERSION}_SHA256SUMS.sig terraform_${TERRAFORM_VERSION}_SHA256SUMS
RUN grep ${TERRAFORM_PLATFORM}.zip terraform_${TERRAFORM_VERSION}_SHA256SUMS | sha256sum -cs

WORKDIR /build/opt/local/bin

RUN unzip /tmp/terraform_${TERRAFORM_VERSION}_${TERRAFORM_PLATFORM}.zip

WORKDIR /build/opt/local/share/doc/terraform

ADD https://raw.githubusercontent.com/hashicorp/terraform/v${TERRAFORM_VERSION}/README.md .
ADD https://raw.githubusercontent.com/hashicorp/terraform/v${TERRAFORM_VERSION}/CHANGELOG.md .

FROM golang:1.21.1-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /terrasync

# Create non-root user
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/terraform" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "1000" \
    terrasync

# RUN mkdir -p /terraform/terraform && chown -R terrasync. /terraform && chmod -R 740 /terraform

# Run the tests in the container
# FROM build-stage AS tester
# RUN go test -v ./...

# Deploy the application binary into a lean image
FROM scratch

# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

USER terrasync

# Import terraform and terrasync
COPY --from=terraform /build/ /
ENV PATH="${PATH}:/opt/local/bin"
COPY --from=builder /terrasync /terrasync

# Run
ENTRYPOINT ["/terrasync"]
