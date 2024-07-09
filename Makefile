build:
	 CGO_ENABLED=1 go build -o bin/scheduler-service cmd/scheduler-service/main.go

tidy:
	go mod tidy

fmt:
	gofumpt -w .
	gci write . --skip-generated -s standard -s default

lint: tidy fmt build
	golangci-lint run

build-docker:
	docker build --tag scheduler-service:v1 .
