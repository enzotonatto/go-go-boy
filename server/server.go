package main

import (
	"bufio"
	"errors"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

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

type Elemento struct {
	Simbolo  rune
	Cor      termbox.Attribute
	CorFundo termbox.Attribute
	Tangivel bool
}

var personagem = Elemento{
	Simbolo:  '☻',
	Cor:      termbox.ColorWhite,
	CorFundo: termbox.ColorDefault,
	Tangivel: true,
}

var vazio = Elemento{
	Simbolo:  ' ',
	Cor:      termbox.ColorDefault,
	CorFundo: termbox.ColorDefault,
	Tangivel: false,
}

var parede = Elemento{
	Simbolo:  '█',
	Cor:      termbox.ColorBlack | termbox.AttrBold | termbox.AttrDim,
	CorFundo: termbox.ColorDarkGray,
	Tangivel: true,
}

var inimigo = Elemento{
	Simbolo:  'Ω',
	Cor:      termbox.ColorRed,
	CorFundo: termbox.ColorDefault,
	Tangivel: true,
}

var estrela = Elemento{
	Simbolo:  '•',
	Cor:      termbox.ColorYellow,
	CorFundo: termbox.ColorDefault,
	Tangivel: true,
}

var portal = Elemento{
	Simbolo:  '0',
	Cor:      termbox.ColorGreen,
	CorFundo: termbox.ColorDefault,
	Tangivel: true,
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

func NewGameServer() *GameServer {
	return &GameServer{
		clients: make(map[string]*Client),
		gameState: GameState{
			Map:     [][]Elemento{},
			Players: make(map[string]Position),
			Running: true,
		},
	}
}

func (gs *GameServer) RegisterClient(args *RegisterArgs, reply *RegisterReply) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if _, exists := gs.clients[args.ClientID]; exists {
		return errors.New("client already registered")
	}

	position := Position{ X: 4, Y: 12 }
	gs.clients[args.ClientID] = &Client{ID: args.ClientID, Position: position }

	gs.gameState.Players[args.ClientID] = position
	gs.gameState.Map[position.Y][position.X] = personagem

	reply.InitialState = gs.gameState
	return nil
}

func (gs *GameServer) SendCommand(args *CommandArgs, reply *struct{}) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	_, exists := gs.clients[args.ClientID]
	if !exists {
		return errors.New("client not registered")
	}

	if args.Command == 'e' {
		go gs.interagir()
	} else {
		gs.mover(args.Command, args.ClientID)
	}

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

func (gs *GameServer) mover(comando rune, clientId string) {
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
	novaPosX, novaPosY := gs.gameState.Players[clientId].X+dx, gs.gameState.Players[clientId].Y+dy

	// Fora dos limites
	if !gs.dentroDosLimites(novaPosX, novaPosY) {
		return
	}

	// Conflito
	if gs.gameState.Map[novaPosY][novaPosX].Tangivel {
		switch gs.gameState.Map[novaPosY][novaPosX].Simbolo {
		case inimigo.Simbolo:
			{
				gs.encerrar(false)
			}
		case estrela.Simbolo:
			{
				gs.encerrar(true)
			}
		case portal.Simbolo:
			{
				novaPosX, novaPosY = gs.teleport(novaPosX, novaPosY)
				gs.gameState.Map[gs.gameState.Players[clientId].Y][gs.gameState.Players[clientId].X] = vazio

				pos := gs.gameState.Players[clientId]
				pos.X, pos.Y = novaPosX, novaPosY
				gs.gameState.Players[clientId] = pos

				gs.gameState.Map[gs.gameState.Players[clientId].Y][gs.gameState.Players[clientId].X] = personagem
			}
		}
		return
	}

	gs.gameState.Map[gs.gameState.Players[clientId].Y][gs.gameState.Players[clientId].X] = vazio

	pos := gs.gameState.Players[clientId]
	pos.X, pos.Y = novaPosX, novaPosY
	gs.gameState.Players[clientId] = pos

	gs.gameState.Map[gs.gameState.Players[clientId].Y][gs.gameState.Players[clientId].X] = personagem
}

func (gs *GameServer) interagir() {
	if gs.gameState.Interacted {
		return
	}

	gs.gameState.StatusMsg = "Você congelou todos!"

	gs.gameState.WhileInteract = true

	time.Sleep(2000 * time.Millisecond)

	gs.gameState.StatusMsg = ""

	gs.gameState.Interacted = true
	gs.gameState.WhileInteract = false
}

func (gs *GameServer) encerrar(ganhou bool) {
	gs.gameState.Running = false

	if ganhou {
		gs.gameState.StatusMsg = "Parabéns! Você ganhou o jogo :)";
	} else {
		gs.gameState.StatusMsg = "Você perdeu o jogo :(";
	}
}

func (gs *GameServer) dentroDosLimites(x int, y int) bool {
	return y >= 0 && y < len(gs.gameState.Map) && x >= 0 && x < len(gs.gameState.Map[y])
}

func (gs *GameServer) teleport(x int, y int) (int, int) {
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

func (gs *GameServer) moverInimigo() {
	for {
		if gs.gameState.WhileInteract {
			continue
		}

		var dirX, dirY, novaPosX, novaPosY int

		for clientID := range gs.gameState.Players {
			if gs.gameState.Enemy.X < gs.gameState.Players[clientID].X {
				dirX = 1
			} else if gs.gameState.Enemy.X > gs.gameState.Players[clientID].X {
				dirX = -1
			}

			if gs.gameState.Enemy.Y < gs.gameState.Players[clientID].Y {
				dirY = 1
			} else if gs.gameState.Enemy.Y > gs.gameState.Players[clientID].Y {
				dirY = -1
			}
		}

		novaPosX = gs.gameState.Enemy.X + dirX
		novaPosY = gs.gameState.Enemy.Y + dirY

		if gs.gameState.Map[novaPosY][novaPosX].Simbolo == personagem.Simbolo {
			gs.encerrar(false)
		}

		if !gs.dentroDosLimites(novaPosX, gs.gameState.Enemy.Y) || gs.gameState.Map[gs.gameState.Enemy.Y][novaPosX].Tangivel {
			novaPosX = gs.gameState.Enemy.X
		}

		if !gs.dentroDosLimites(gs.gameState.Enemy.X, novaPosY) || gs.gameState.Map[novaPosY][gs.gameState.Enemy.X].Tangivel {
			novaPosY = gs.gameState.Enemy.Y
		}

		gs.gameState.Map[gs.gameState.Enemy.Y][gs.gameState.Enemy.X] = vazio
		gs.gameState.Enemy.X, gs.gameState.Enemy.Y = novaPosX, novaPosY
		gs.gameState.Map[gs.gameState.Enemy.Y][gs.gameState.Enemy.X] = inimigo

		time.Sleep(1000 * time.Millisecond)
	}
}

func (gs *GameServer) moverEstrela() {
	for {
		if gs.gameState.WhileInteract {
			continue
		}

		var novaPosX, novaPosY int

		for {
			novaPosY = rand.Intn(len(gs.gameState.Map))
			novaPosX = rand.Intn(len(gs.gameState.Map[0]))

			if !gs.gameState.Map[novaPosY][novaPosX].Tangivel {
				break
			}
		}

		gs.gameState.Map[gs.gameState.Star.Y][gs.gameState.Star.X] = vazio
		gs.gameState.Star.X, gs.gameState.Star.Y = novaPosX, novaPosY
		gs.gameState.Map[gs.gameState.Star.Y][gs.gameState.Star.X] = estrela

		time.Sleep(3000 * time.Millisecond)
	}
}

func (gs *GameServer) carregarMapa(nomeArquivo string) {
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
			case parede.Simbolo:
				elementoAtual = parede
			case inimigo.Simbolo:
				gs.gameState.Enemy.X = x
				gs.gameState.Enemy.Y = y
				elementoAtual = vazio
			case estrela.Simbolo:
				gs.gameState.Star.X = x
				gs.gameState.Star.Y = y
				elementoAtual = vazio
			case portal.Simbolo:
				elementoAtual = portal
			}
			linhaElementos = append(linhaElementos, elementoAtual)
		}
		gs.gameState.Map = append(gs.gameState.Map, linhaElementos)
		y++
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func main() {
	gameServer := NewGameServer()

	gameServer.carregarMapa("mapa.txt")

	go gameServer.moverInimigo()
	go gameServer.moverEstrela()

	rpc.Register(gameServer)
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Listener error:", err)
	}
	defer listener.Close()

	log.Println("Serving RPC server on port 1234")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Erro ao aceitar conexão: ", err)
			continue
		}

		go rpc.ServeConn(conn)
	}
}
