package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

type Elemento struct {
	simbolo rune
	cor termbox.Attribute
	corFundo termbox.Attribute
	tangivel bool
}
	
var personagem = Elemento{
	simbolo: '☻',
	cor: termbox.ColorWhite,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

var parede = Elemento{
	simbolo: '█',
	cor: termbox.ColorBlack | termbox.AttrBold | termbox.AttrDim,
	corFundo: termbox.ColorDarkGray,
	tangivel: true,
}

var vazio = Elemento{
	simbolo: ' ',
	cor: termbox.ColorDefault,
	corFundo: termbox.ColorDefault,
	tangivel: false,
}

var inimigo = Elemento{
	simbolo: 'Ω',
	cor: termbox.ColorRed,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

var estrela = Elemento{
	simbolo: '•',
	cor: termbox.ColorYellow,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

var portal = Elemento{
	simbolo: '0',
	cor: termbox.ColorGreen,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

var mapa [][]Elemento

var statusMsg string

var posX, posY int
var posXinimigo, posYinimigo int
var posXestrela, posYestrela int

func main(){
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	carregarMapa("mapa.txt")
	desenhaTudo()

	go moverInimigo()
	go moverEstrela()

	for {
		event := termbox.PollEvent();
		
		if event.Type == termbox.EventKey {
			if event.Key == termbox.KeyEsc {
				return
			}
			if (event.Ch == 'e') {
				interagir()
			} else {
				mover(event.Ch)
			}
			
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
				posX, posY = x, y
				elementoAtual = vazio
			case inimigo.simbolo:
				posXinimigo, posYinimigo = x, y
				elementoAtual = vazio
			case estrela.simbolo:
				posXestrela, posYestrela = x, y
				elementoAtual = vazio
			case portal.simbolo:
				elementoAtual = portal
			}
			linhaElementos = append(linhaElementos, elementoAtual)
		}
		mapa = append(mapa, linhaElementos)
		y++
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func desenhaTudo() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for y, linha := range mapa {
		for x, elem := range linha {
			termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
		}
	}

	desenhaBarraDeStatus()

	termbox.Flush()
}

func desenhaBarraDeStatus() {
	for i, c := range statusMsg {
		termbox.SetCell(i, len(mapa)+1, c, termbox.ColorBlack, termbox.ColorDefault)
	}
	msg := "Use WASD para mover e E para interagir. ESC para sair."
	for i, c := range msg {
		termbox.SetCell(i, len(mapa)+3, c, termbox.ColorBlack, termbox.ColorDefault)
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
	novaPosX, novaPosY := posX+dx, posY+dy

	// Fora dos limites
	if !dentroDosLimites(novaPosX, novaPosY) {
		return
	}

	// Conflito
	if mapa[novaPosY][novaPosX].tangivel {
		switch mapa[novaPosY][novaPosX].simbolo {
		case inimigo.simbolo: {
			encerrar(false)
		}
		case estrela.simbolo: {
			encerrar(true)
		}
		case portal.simbolo: {
			// muda de posicao
		}
		}
		return
	}

	mapa[posY][posX] = vazio
	posX, posY = novaPosX, novaPosY
	mapa[posY][posX] = personagem
}

func interagir() {
	statusMsg = fmt.Sprintf("Interagindo em (%d, %d)", posX, posY) // Pode remover
	// Congela por 5 segundos
}

func moverInimigo() {
	for {
		var dirX, dirY, novaPosX, novaPosY int

		if posXinimigo < posX {
			dirX = 1
		} else if posXinimigo > posX {
			dirX = -1
		}

		if posYinimigo < posY {
			dirY = 1
		} else if posYinimigo > posY {
			dirY = -1
		}

		novaPosX = posXinimigo + dirX
		novaPosY = posYinimigo + dirY

		if mapa[novaPosY][novaPosX].simbolo == personagem.simbolo {
			encerrar(false)
		}

		if !dentroDosLimites(novaPosX, posYinimigo) || mapa[posYinimigo][novaPosX].tangivel {
			novaPosX = posXinimigo
		}

		if !dentroDosLimites(posXinimigo, novaPosY) || mapa[novaPosY][posXinimigo].tangivel {
			novaPosY = posYinimigo
		}

		mapa[posYinimigo][posXinimigo] = vazio
		posXinimigo, posYinimigo = novaPosX, novaPosY
		mapa[posYinimigo][posXinimigo] = inimigo

		desenhaTudo()
		time.Sleep(800 * time.Millisecond)
	}
}

func moverEstrela() {
	for {
		var novaPosX, novaPosY int

		for {
			novaPosY = rand.Intn(len(mapa))
			novaPosX = rand.Intn(len(mapa[0]))

			if !mapa[novaPosY][novaPosX].tangivel {
				break
			}
		}

		mapa[posYestrela][posXestrela] = vazio
		posXestrela, posYestrela = novaPosX, novaPosY
		mapa[posYestrela][posXestrela] = estrela

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
	return y >= 0 && y < len(mapa) && x >= 0 && x < len(mapa[y])
}