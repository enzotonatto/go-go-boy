package main

import (
	"log"
	"net/rpc"
	"strconv"
	"time"

	"github.com/nsf/termbox-go"
)

type Elemento struct {
	Simbolo  rune
	Cor      termbox.Attribute
	CorFundo termbox.Attribute
	Tangivel bool
}

type Position struct {
	X int
	Y int
}

type GameState struct {
	Map                       [][]Elemento
	Players                   map[string]Position
	Enemy                     Position
	Star                      Position
	StatusMsg                 string
	Interacted, WhileInteract bool
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
	ClientID       string
	SequenceNumber int
	Command        string
}

type GameClient struct {
	clientID       string
	rpcClient      *rpc.Client
	sequenceNumber int
	gameState      GameState
}

func NewGameClient(clientID string, serverAddress string) *GameClient {
	client, err := rpc.Dial("tcp", serverAddress)
	if err != nil {
		log.Fatal("Dialing:", err)
	}

	gameClient := &GameClient{
		clientID:       clientID,
		rpcClient:      client,
		sequenceNumber: 0,
	}

	// Register the client with the server
	registerArgs := RegisterArgs{ClientID: clientID}
	var registerReply RegisterReply
	err = client.Call("GameServer.RegisterClient", &registerArgs, &registerReply)
	if err != nil {
		log.Fatal("Registering client:", err)
	}

	gameClient.gameState = registerReply.InitialState
	return gameClient
}

func (gc *GameClient) SendCommand(command string) {
	gc.sequenceNumber++
	commandArgs := CommandArgs{
		ClientID:       gc.clientID,
		SequenceNumber: gc.sequenceNumber,
		Command:        command,
	}
	var reply struct{}
	err := gc.rpcClient.Call("GameServer.SendCommand", &commandArgs, &reply)
	if err != nil {
		log.Println("Sending command:", err)
	}
}

func (gc *GameClient) UpdateGameState() {
	gameStateArgs := GameStateArgs{ClientID: gc.clientID}
	var gameStateReply GameStateReply
	err := gc.rpcClient.Call("GameServer.GetGameState", &gameStateArgs, &gameStateReply)
	if err != nil {
		log.Println("Getting game state:", err)
		return
	}

	gc.gameState = gameStateReply.State
}

var gameClient *GameClient
var clientID string
var serverAddress string

func main() {
	clientID = strconv.FormatInt(time.Now().Unix(), 10)
	serverAddress = "localhost:1234"

	gameClient = NewGameClient(clientID, serverAddress)

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	desenhaTudo()

	go updateScreen()
	for {
		event := termbox.PollEvent()

		if event.Type == termbox.EventKey {
			gameClient.SendCommand(strconv.QuoteRune(event.Ch))
		}
	}
}

func updateScreen() {
	gameClient.UpdateGameState()
	time.Sleep(1000 * time.Millisecond)
}

func desenhaTudo() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for y, linha := range gameClient.gameState.Map {
		for x, elem := range linha {
			termbox.SetCell(x, y, elem.Simbolo, elem.Cor, elem.CorFundo)
		}
	}

	desenhaBarraDeStatus()

	termbox.Flush()
}

func desenhaBarraDeStatus() {
	for i, c := range gameClient.gameState.StatusMsg {
		termbox.SetCell(i, len(gameClient.gameState.Map)+1, c, termbox.ColorBlack, termbox.ColorDefault)
	}
	msg := "Use WASD para mover e E para interagir. ESC para sair."
	for i, c := range msg {
		termbox.SetCell(i, len(gameClient.gameState.Map)+3, c, termbox.ColorBlack, termbox.ColorDefault)
	}
}
