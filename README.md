# gRPC-examples

```bash
# kratos errors
protoc --proto_path=. --proto_path=./third_party --go_out=paths=source_relative:. --go-errors_out=paths=source_relative:. ${compile_path}
```

```bash
# gRPC
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ${compile_path}
```