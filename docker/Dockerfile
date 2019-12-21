FROM golang:1.13-alpine AS builder

WORKDIR /go/src/github.com/drasko/fluxm2m
COPY . .
RUN make $SVC_NAME \
    && mv build/fluxm2m /exe

FROM scratch
COPY --from=builder /exe /
ENTRYPOINT ["/exe"]
