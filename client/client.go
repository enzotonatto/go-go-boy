package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

type Elemento struct {
	simbolo  rune
	cor      termbox.Attribute
	corFundo termbox.Attribute
	tangivel bool
}

var personagem = Elemento{
	simbolo:  '☻',
	cor:      termbox.ColorWhite,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

var parede = Elemento{
	simbolo:  '█',
	cor:      termbox.ColorBlack | termbox.AttrBold | termbox.AttrDim,
	corFundo: termbox.ColorDarkGray,
	tangivel: true,
}

var vazio = Elemento{
	simbolo:  ' ',
	cor:      termbox.ColorDefault,
	corFundo: termbox.ColorDefault,
	tangivel: false,
}

var inimigo = Elemento{
	simbolo:  'Ω',
	cor:      termbox.ColorRed,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

var estrela = Elemento{
	simbolo:  '•',
	cor:      termbox.ColorYellow,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

var portal = Elemento{
	simbolo:  '0',
	cor:      termbox.ColorGreen,
	corFundo: termbox.ColorDefault,
	tangivel: true,
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
	statusMsg                 string
	interacted, whileInteract bool
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
	clientID = "client1"
	serverAddress = "localhost:1234"

	gameClient = NewGameClient(clientID, serverAddress)

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	carregarMapa("../mapa.txt")
	desenhaTudo()

	go moverInimigo()
	go moverEstrela()

	for {
		event := termbox.PollEvent()

		if event.Type == termbox.EventKey {
			if event.Key == termbox.KeyEsc {
				return
			}
			if event.Ch == 'e' {
				go interagir()
			} else {
				mover(event.Ch)
			}

			gameClient.UpdateGameState()
			desenhaTudo()
		}
	}
}

func carregarMapa(nomeArquivo string) {
	arquivo, err := os.Open(nomeArquivo)

	if err != nil {
		panic(err)
	}
	defer arquivo.Close()

	scanner := bufio.NewScanner(arquivo)
	y := 0

	for scanner.Scan() {
		linhaTexto := scanner.Text()
		var linhaElementos []Elemento
		for x, char := range linhaTexto {
			elementoAtual := vazio
			switch char {
			case parede.simbolo:
				elementoAtual = parede
			case personagem.simbolo:
				pos := gameClient.gameState.Players[clientID]
				pos.X = x
				pos.Y = y
				gameClient.gameState.Players[clientID] = pos
				elementoAtual = vazio
			case inimigo.simbolo:
				gameClient.gameState.Enemy.X = x
				gameClient.gameState.Enemy.Y = y
				elementoAtual = vazio
			case estrela.simbolo:
				gameClient.gameState.Star.X = x
				gameClient.gameState.Star.Y = y
				elementoAtual = vazio
			case portal.simbolo:
				elementoAtual = portal
			}
			linhaElementos = append(linhaElementos, elementoAtual)
		}
		gameClient.gameState.Map = append(gameClient.gameState.Map, linhaElementos)
		y++
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func desenhaTudo() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for y, linha := range gameClient.gameState.Map {
		for x, elem := range linha {
			termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
		}
	}

	desenhaBarraDeStatus()

	termbox.Flush()
}

func desenhaBarraDeStatus() {
	for i, c := range gameClient.gameState.statusMsg {
		termbox.SetCell(i, len(gameClient.gameState.Map)+1, c, termbox.ColorBlack, termbox.ColorDefault)
	}
	msg := "Use WASD para mover e E para interagir. ESC para sair."
	for i, c := range msg {
		termbox.SetCell(i, len(gameClient.gameState.Map)+3, c, termbox.ColorBlack, termbox.ColorDefault)
	}
}

func mover(comando rune) {
	dx, dy := 0, 0
	switch comando {
	case 'w':
		dy = -1
	case 'a':
		dx = -1
	case 's':
		dy = 1
	case 'd':
		dx = 1
	}
	novaPosX, novaPosY := gameClient.gameState.Players[clientID].X+dx, gameClient.gameState.Players[clientID].Y+dy

	// Fora dos limites
	if !dentroDosLimites(novaPosX, novaPosY) {
		return
	}

	// Conflito
	if gameClient.gameState.Map[novaPosY][novaPosX].tangivel {
		switch gameClient.gameState.Map[novaPosY][novaPosX].simbolo {
		case inimigo.simbolo:
			{
				encerrar(false)
			}
		case estrela.simbolo:
			{
				encerrar(true)
			}
		case portal.simbolo:
			{
				novaPosX, novaPosY = teleport(novaPosX, novaPosY)
				gameClient.gameState.Map[gameClient.gameState.Players[clientID].Y][gameClient.gameState.Players[clientID].X] = vazio

				pos := gameClient.gameState.Players[clientID]
				pos.X, pos.Y = novaPosX, novaPosY
				gameClient.gameState.Players[clientID] = pos

				gameClient.gameState.Map[gameClient.gameState.Players[clientID].Y][gameClient.gameState.Players[clientID].X] = personagem
			}
		}
		return
	}

	gameClient.gameState.Map[gameClient.gameState.Players[clientID].Y][gameClient.gameState.Players[clientID].X] = vazio

	pos := gameClient.gameState.Players[clientID]
	pos.X, pos.Y = novaPosX, novaPosY
	gameClient.gameState.Players[clientID] = pos

	gameClient.gameState.Map[gameClient.gameState.Players[clientID].Y][gameClient.gameState.Players[clientID].X] = personagem
}

func interagir() {
	if gameClient.gameState.interacted {
		return
	}

	gameClient.gameState.statusMsg = "Você congelou todos!"

	gameClient.gameState.whileInteract = true

	desenhaTudo()

	time.Sleep(2000 * time.Millisecond)

	gameClient.gameState.statusMsg = ""

	gameClient.gameState.interacted = true
	gameClient.gameState.whileInteract = false

	desenhaTudo()
}

func moverInimigo() {
	for {
		if gameClient.gameState.whileInteract {
			continue
		}

		var dirX, dirY, novaPosX, novaPosY int

		if gameClient.gameState.Enemy.X < gameClient.gameState.Players[clientID].X {
			dirX = 1
		} else if gameClient.gameState.Enemy.X > gameClient.gameState.Players[clientID].X {
			dirX = -1
		}

		if gameClient.gameState.Enemy.Y < gameClient.gameState.Players[clientID].Y {
			dirY = 1
		} else if gameClient.gameState.Enemy.Y > gameClient.gameState.Players[clientID].Y {
			dirY = -1
		}

		novaPosX = gameClient.gameState.Enemy.X + dirX
		novaPosY = gameClient.gameState.Enemy.Y + dirY

		if gameClient.gameState.Map[novaPosY][novaPosX].simbolo == personagem.simbolo {
			encerrar(false)
		}

		if !dentroDosLimites(novaPosX, gameClient.gameState.Enemy.Y) || gameClient.gameState.Map[gameClient.gameState.Enemy.Y][novaPosX].tangivel {
			novaPosX = gameClient.gameState.Enemy.X
		}

		if !dentroDosLimites(gameClient.gameState.Enemy.X, novaPosY) || gameClient.gameState.Map[novaPosY][gameClient.gameState.Enemy.X].tangivel {
			novaPosY = gameClient.gameState.Enemy.Y
		}

		gameClient.gameState.Map[gameClient.gameState.Enemy.Y][gameClient.gameState.Enemy.X] = vazio
		gameClient.gameState.Enemy.X, gameClient.gameState.Enemy.Y = novaPosX, novaPosY
		gameClient.gameState.Map[gameClient.gameState.Enemy.Y][gameClient.gameState.Enemy.X] = inimigo

		desenhaTudo()
		time.Sleep(800 * time.Millisecond)
	}
}

func moverEstrela() {
	for {
		if gameClient.gameState.whileInteract {
			continue
		}

		var novaPosX, novaPosY int

		for {
			novaPosY = rand.Intn(len(gameClient.gameState.Map))
			novaPosX = rand.Intn(len(gameClient.gameState.Map[0]))

			if !gameClient.gameState.Map[novaPosY][novaPosX].tangivel {
				break
			}
		}

		gameClient.gameState.Map[gameClient.gameState.Star.Y][gameClient.gameState.Star.X] = vazio
		gameClient.gameState.Star.X, gameClient.gameState.Star.Y = novaPosX, novaPosY
		gameClient.gameState.Map[gameClient.gameState.Star.Y][gameClient.gameState.Star.X] = estrela

		desenhaTudo()
		time.Sleep(3000 * time.Millisecond)
	}
}

func encerrar(ganhou bool) {
	termbox.Close()

	if ganhou {
		fmt.Println("Parabéns! Você ganhou o jogo :)")
	} else {
		fmt.Println("Você perdeu o jogo :(")
	}

	os.Exit(1)
}

func dentroDosLimites(x int, y int) bool {
	return y >= 0 && y < len(gameClient.gameState.Map) && x >= 0 && x < len(gameClient.gameState.Map[y])
}

func teleport(x int, y int) (int, int) {
	portalA := [2]int{79, 2}
	portalB := [2]int{0, 28}

	if x == portalA[0] && y == portalA[1] {
		return portalB[0] + 1, portalB[1]
	}

	if x == portalB[0] && y == portalB[1] {
		return portalA[0] - 1, portalA[1]
	}

	return x, y
}
