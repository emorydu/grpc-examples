server:
	go run cmd/server/main.go --port 8080

client:
	go run cmd/client/main.go --addr 0.0.0.0:8080

test:
	go test -cover -race ./...