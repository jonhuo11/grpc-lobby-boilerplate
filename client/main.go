package main

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	api "grpc-gamelobby-boilerplate/client/apicopy"
)

func main() {
	var addr string
	fmt.Print("Enter server address (empty for localhost:9000): ")
	fmt.Scanln(&addr)

	if addr == "" {
		addr = "localhost:9000"
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(addr, opts...)
	defer conn.Close()
	if err != nil {
		panic(err)
	}
	client := api.NewLobbyServiceClient(conn)

	var username string
	for {
		fmt.Print("Enter your username: ")
		fmt.Scanln(&username)

		if _, err := client.RegisterPlayer(context.Background(), &api.RegisterPlayerRequest{
			Username: username,
		}); err != nil {
			fmt.Println(err.Error())
			continue
		}
		break
	}

	var lobbyList []string = make([]string, 0)
	//inLobby := false
	lobbystream, err := client.LobbyStream(context.Background())

	lobbystream.Send(&api.PlayerRequest{Action: &api.PlayerRequest_Login{Login: &api.LoginPlayer{
		Username: username,
	}}})

	for {
		res, err := lobbystream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		switch res.Action.(type) {
		case *api.LobbyResponse_LobbyList:
			lobbyList = res.GetLobbyList().LobbyList
			fmt.Println(lobbyList)
		default:
			fmt.Println("not implemented")
		}
	}
}
