package main

import (
	"errors"
	"log"
	"net"
	"net/rpc"
	"sync"
)

/**

var mapa [][]Elemento
var interacted, whileInteract bool
var statusMsg string

var posX, posY int
var posXinimigo, posYinimigo int
var posXestrela, posYestrela int

*/

type GameServer struct {
    mu        sync.Mutex
    clients   map[string]*Client
    gameState GameState
}

type Client struct {
    ID       string
    Position Position
}

type Position struct {
    X int
    Y int
}

type GameState struct {
    Map      [][]int
    Players  map[string]Position
}

type RegisterArgs struct {
    ClientID string
}

type RegisterReply struct {
    InitialState GameState
}

type GameStateArgs struct {
    ClientID string
}

type GameStateReply struct {
    State GameState
}

type CommandArgs struct {
    ClientID      string
    SequenceNumber int
    Command       string
}

func NewGameServer() *GameServer {
    return &GameServer{
        clients:  make(map[string]*Client),
        gameState: GameState{
            Map:     [][]int{}, // Initialize with the map
            Players: make(map[string]Position),
        },
    }
}

func (gs *GameServer) RegisterClient(args *RegisterArgs, reply *RegisterReply) error {
    gs.mu.Lock()
    defer gs.mu.Unlock()

    if _, exists := gs.clients[args.ClientID]; exists {
        return errors.New("client already registered")
    }

    gs.clients[args.ClientID] = &Client{ID: args.ClientID, Position: Position{X: 0, Y: 0}} // Default position

    gs.gameState.Players[args.ClientID] = Position{X: 0, Y: 0}

    reply.InitialState = gs.gameState
    return nil
}

func (gs *GameServer) SendCommand(args *CommandArgs, reply *struct{}) error {
    gs.mu.Lock()
    defer gs.mu.Unlock()

    client, exists := gs.clients[args.ClientID]
    if !exists {
        return errors.New("client not registered")
    }

    // Process the command and update the game state accordingly
    // Example: simple movement command
    switch args.Command {
    case "move_up":
        client.Position.Y -= 1
    case "move_down":
        client.Position.Y += 1
    case "move_left":
        client.Position.X -= 1
    case "move_right":
        client.Position.X += 1
    }

    gs.gameState.Players[args.ClientID] = client.Position
    return nil
}

func (gs *GameServer) GetGameState(args *GameStateArgs, reply *GameStateReply) error {
    gs.mu.Lock()
    defer gs.mu.Unlock()

    _, exists := gs.clients[args.ClientID]
    if !exists {
        return errors.New("client not registered")
    }

    reply.State = gs.gameState
    return nil
}

func main() {
    gameServer := NewGameServer()
    rpc.Register(gameServer)
    listener, err := net.Listen("tcp", ":1234")
    if err != nil {
        log.Fatal("Listener error:", err)
    }
    defer listener.Close()
    log.Println("Serving RPC server on port 1234")
    rpc.Accept(listener)
}
