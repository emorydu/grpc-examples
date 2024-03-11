server:
	go run cmd/server/main.go --port 8080

server1:
	go run cmd/server/main.go --port 50051

server2:
	go run cmd/server/main.go --port 50052

server1-tls:
	go run cmd/server/main.go --port 50051 -tls

server2-tls:
	go run cmd/server/main.go --port 50052 -tls

client:
	go run cmd/client/main.go --addr 0.0.0.0:8080

client-tls:
	go run cmd/client/main.go --addr 0.0.0.0:8080 --tls

test:
	go test -cover -race ./...

clean:
	rm pb/*.go

cert:
	cd cert; ./gen.sh; cd ..

.PHONY: clean server client test cert