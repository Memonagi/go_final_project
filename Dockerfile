FROM golang:1.21.6 as builder

COPY ../.. /src

WORKDIR /src

RUN CGO_ENABLED=1 GOOS=linux go build -o bin/scheduler-service cmd/scheduler-service/main.go

FROM debian:stable-slim

COPY --from=builder /src/bin/scheduler-service /app/bin/scheduler-service
COPY --from=builder /src/scheduler.db /app/scheduler.db
COPY --from=builder /src/web/. /app/web/.

WORKDIR /app

ENV TODO_PORT=7540
ENV TODO_DBFILE=scheduler.db

EXPOSE 7540

ENTRYPOINT ["./bin/scheduler-service"]