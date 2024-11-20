FROM chainguard/go AS builder

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache \
    go build -o /app/jira-bot

FROM chainguard/wolfi-base

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/jira-bot /app/jira-bot

ENTRYPOINT ["/app/jira-bot"]
CMD ["server"]

EXPOSE 8000