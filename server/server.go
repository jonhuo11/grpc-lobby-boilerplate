package main

import (
	"context"
	"errors"
	"grpc-gamelobby-boilerplate/server/api"
	"io"
	"log"
	"reflect"
)

type server struct {
	api.UnimplementedLobbyServiceServer

	state serverState
}

func initServer() (*server, error) {
	return &server{
		state: initServerState(),
	}, nil
}

func (s *server) RegisterPlayer(ctx context.Context, req *api.RegisterPlayerRequest) (*api.RegisterPlayerResponse, error) {
	if s.state.isPlayerRegistered(req.Username) {
		return nil, errors.New("user already registered")
	}
	if len([]rune(req.Username)) < 3 {
		return nil, errors.New("username must be at least 3 characters long")
	}
	s.state.registerPlayerUnsafe(req.Username)
	log.Println("registered player " + req.Username)
	return &api.RegisterPlayerResponse{
		Username: req.Username,
	}, nil
}

func (s *server) LobbyStream(connection api.LobbyService_LobbyStreamServer) error {
	log.Println("client connected")
	var username string = ""
	for {
		req, err := connection.Recv()

		if err != nil {
			log.Println("client disconnecting...")
			if username != "" {
				if s.state.isPlayerInLinkedLobby(username) {
					s.state.leaveLobbyUnsafe(username)
				}
				s.state.logoutPlayerUnsafe(username)
				log.Println("player " + username + " logged out")
			}
			if err == io.EOF {
				return nil
			}
			return err
		}

		if req.Action == nil {
			connection.Send(failResponse("bad request format"))
			continue
		}

		log.Println("handling request of type " + reflect.TypeOf(req.Action).String())

		if s.state.isPlayerLoggedIn(username) { // assume player is registered and logged in
			switch req.Action.(type) {

			case *api.PlayerRequest_Login:
				connection.Send(failResponse("you already logged in"))

			case *api.PlayerRequest_CreateLobby:
				if s.state.isLobbyRegistered(req.GetCreateLobby().Name) {
					connection.Send(failResponse("lobby with this name already exists"))
				} else {
					if len([]rune(req.GetCreateLobby().Name)) < 3 {
						connection.Send(failResponse("lobby name must be at least 3 characters long"))
						continue
					}

					lobbyinfo := s.state.registerLobbyUnsafe(req.GetCreateLobby().Name, username)
					connection.Send(&api.LobbyResponse{Action: &api.LobbyResponse_LobbyDetails{LobbyDetails: lobbyinfo}})
					log.Println("player " + username + " created lobby " + req.GetCreateLobby().Name)
				}

			case *api.PlayerRequest_JoinLobby:
				if s.state.isLobbyRegistered(req.GetJoinLobby().Name) {
					lobbyinfo := s.state.joinLobbyUnsafe(req.GetJoinLobby().Name, username)
					connection.Send(&api.LobbyResponse{Action: &api.LobbyResponse_LobbyDetails{LobbyDetails: lobbyinfo}})
					log.Println("player " + username + " joined lobby " + req.GetJoinLobby().Name)
				} else {
					connection.Send(failResponse("no lobby with this name exists"))
				}

			case *api.PlayerRequest_LeaveLobby:
				if s.state.isPlayerInLinkedLobby(username) {
					lobbyinfo := s.state.leaveLobbyUnsafe(username)
					connection.Send(&api.LobbyResponse{Action: &api.LobbyResponse_LobbyDetails{LobbyDetails: lobbyinfo}})
					log.Println("player " + username + " left lobby " + lobbyinfo.Name)
				} else {
					connection.Send(failResponse("player not in lobby"))
				}

			default:
				connection.Send(failResponse("unimplemented"))

			}
		} else {
			switch req.Action.(type) {
			case *api.PlayerRequest_Login:
				if s.state.isPlayerRegistered(req.GetLogin().Username) {
					username = req.GetLogin().Username
					s.state.loginPlayerUnsafe(username, connection)
					connection.Send(&api.LobbyResponse{Action: &api.LobbyResponse_LobbyList{LobbyList: s.state.getLobbyListUnsafe()}})
					log.Println("player " + username + " logged in")
				} else {
					connection.Send(failResponse("player not registered"))
				}
			default:
				connection.Send(failResponse("user on connection not logged in"))
			}
		}
	}
}

func failResponse(info string) *api.LobbyResponse {
	return &api.LobbyResponse{
		Action: &api.LobbyResponse_FailEvent{
			FailEvent: &api.FailEvent{
				Info: info,
			},
		},
	}
}
