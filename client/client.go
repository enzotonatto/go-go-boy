package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
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
	Map                       			[][]Elemento
	Players                   			map[string]Position
	Enemy                     			Position
	Star                      			Position
	StatusMsg                 			string
	Interacted, WhileInteract, Running	bool
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
	Command        rune
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

func (gc *GameClient) SendCommand(command rune) {
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
	if !gameStateReply.State.Running {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		termbox.Close()
		fmt.Println(gameStateReply.State.StatusMsg)
		os.Exit(1)
	}

	gc.gameState = gameStateReply.State
}

var gameClient *GameClient
var clientID string
var serverAddress string

func main() {
	ip := flag.String("ip", "localhost", "ip to connect")
	port := flag.String("port", "1234", "port to connect")

	flag.Parse()

	clientID = strconv.FormatInt(time.Now().Unix(), 10)
	serverAddress = *ip + ":" + *port

	gameClient = NewGameClient(clientID, serverAddress)

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	go updateScreen()

	for {
		event := termbox.PollEvent()

		if event.Type == termbox.EventKey {
			if event.Key == termbox.KeyEsc {
				return
			}
			gameClient.SendCommand(event.Ch)
		}
	}
}

func updateScreen() {
	for {
		gameClient.UpdateGameState()
		desenhaTudo()
		time.Sleep(250 * time.Millisecond)
	}
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
