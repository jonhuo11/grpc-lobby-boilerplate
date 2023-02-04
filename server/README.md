player actions
- create lobby
- join lobby
- leave lobby
- ring bell

server actions
- sound bell (notify all connections bell has rung)

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    lobby_api.proto