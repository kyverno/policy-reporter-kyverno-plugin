FROM golang:1.16-buster as builder

WORKDIR /app
COPY . .

RUN go get -d -v \
    && go install -v

RUN make build

FROM alpine:latest
LABEL MAINTAINER "Frank Jogeleit <frank.jogeleit@gweb.de>"

WORKDIR /app

RUN apk add --update --no-cache ca-certificates

RUN addgroup -S kyverno-plugin && adduser -u 1234 -S kyverno-plugin -G kyverno-plugin

USER 1234

COPY --from=builder /app/LICENSE.md .
COPY --from=builder /app/build/kyverno-plugin /app/kyverno-plugin

EXPOSE 2112

ENTRYPOINT ["/app/kyverno-plugin", "run"]