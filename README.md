# Sloth-Grpc

## Compile proto file to golang server

```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./sql_service/sql_service.proto
```

## Compile proto file to python client

```
python -m grpc_tools.protoc -I./sql_service --python_out=. --grpc_python_out=. ./sql_service/sql_service.proto
```

## Compile and run

```
go build server.go && server.exe
```
