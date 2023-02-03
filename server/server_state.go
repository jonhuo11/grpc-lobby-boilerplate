package main

import (
	"grpc-gamelobby-boilerplate/server/api"
	"sync"
)

type serverState struct {
	mu         sync.Mutex
	playerList map[string]*Player
	lobbyList  map[string]*Lobby
}

func initServerState() serverState {
	return serverState{
		playerList: make(map[string]*Player),
		lobbyList:  make(map[string]*Lobby),
	}
}

type Player struct {
	Username   string
	Connection api.LobbyService_LobbyStreamServer
	Lobby      *Lobby
}

type Lobby struct {
	Name       string
	State      api.LobbyState
	PlayerList map[string]*Player
}

func (s *serverState) isPlayerRegistered(username string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, e := s.playerList[username]
	return e
}

func (s *serverState) isLobbyRegistered(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, e := s.lobbyList[name]
	return e
}

func (s *serverState) isPlayerInLobby(username, lobbyname string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, e := s.lobbyList[lobbyname].PlayerList[username]
	return e
}

func (s *serverState) isPlayerInLinkedLobby(username string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.playerList[username] == nil {
		return false
	} else {
		return s.playerList[username].Lobby != nil
	}
}

func (s *serverState) isPlayerLoggedIn(username string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.playerList[username] == nil {
		return false
	}
	return s.playerList[username].Connection != nil
}

// unsafe methods, verify if good in handler
func (s *serverState) registerPlayerUnsafe(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playerList[username] = &Player{
		Username:   username,
		Connection: nil,
	}
}

func (s *serverState) loginPlayerUnsafe(username string, connection api.LobbyService_LobbyStreamServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playerList[username].Connection = connection
}

func (s *serverState) logoutPlayerUnsafe(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playerList[username].Connection = nil
}

func (s *serverState) getLobbyListUnsafe() *api.LobbyList {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0)
	for k := range s.lobbyList {
		out = append(out, k)
	}
	return &api.LobbyList{
		LobbyList: out,
	}
}

func (s *serverState) registerLobbyUnsafe(name, creator string) *api.LobbyDetails {
	s.mu.Lock()
	defer s.mu.Unlock()
	l := make(map[string]*Player)
	l[creator] = s.playerList[creator]
	s.lobbyList[name] = &Lobby{
		Name:       name,
		PlayerList: l,
		State:      api.LobbyState_IN_LOBBY,
	}
	s.playerList[creator].Lobby = s.lobbyList[name]
	return &api.LobbyDetails{
		Name:       name,
		PlayerList: []string{creator},
		State:      api.LobbyState_IN_LOBBY,
	}
}

func (s *serverState) joinLobbyUnsafe(name, username string) *api.LobbyDetails {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lobbyList[name].PlayerList[username] = s.playerList[username]
	s.playerList[username].Lobby = s.lobbyList[name]

	l := make([]string, 0)
	for p := range s.lobbyList[name].PlayerList {
		l = append(l, p)
	}
	return &api.LobbyDetails{
		Name:       name,
		PlayerList: l,
		State:      s.lobbyList[name].State,
	}
}

func (s *serverState) leaveLobbyUnsafe(username string) *api.LobbyDetails {
	s.mu.Lock()
	defer s.mu.Unlock()
	lobbyname := s.playerList[username].Lobby.Name
	delete(s.lobbyList[lobbyname].PlayerList, username)
	l := make([]string, 0)
	for p := range s.lobbyList[lobbyname].PlayerList {
		l = append(l, p)
	}
	s.playerList[username].Lobby = nil

	return &api.LobbyDetails{
		Name:       lobbyname,
		PlayerList: l,
		State:      s.lobbyList[lobbyname].State,
	}
}
