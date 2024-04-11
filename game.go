package main

import (
    "bufio"
    "github.com/nsf/termbox-go"
    "os"
    "fmt"
)

type Elemento struct {
    simbolo rune
    cor termbox.Attribute
    corFundo termbox.Attribute
    tangivel bool
}
	
var personagem = Elemento{
	simbolo: '☺',
	cor: termbox.ColorWhite,
	corFundo: termbox.ColorDefault,
	tangivel: true,
}

var parede = Elemento{
	simbolo: '▤',
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
    simbolo: '#',
    cor: termbox.ColorRed,
    corFundo: termbox.ColorDefault,
    tangivel: true,
}

var estrela = Elemento{
    simbolo: '☆',
    cor: termbox.ColorYellow,
    corFundo: termbox.ColorDefault,
    tangivel: false,
}

var portal = Elemento{
    simbolo: '0',
    cor: termbox.ColorGreen,
    corFundo: termbox.ColorDefault,
    tangivel: false,
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

    // Muda estrela de lugar
    // Movimenta inimigo

	for {
        event := termbox.PollEvent();
		
		if event.Type == termbox.EventKey {
			switch event.Key {
			case termbox.KeyEsc:
				return
			case 'e':
				interagir()
			default:
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
                elementoAtual = personagem
			case inimigo.simbolo:
                posXinimigo, posYinimigo = x, y
                elementoAtual = inimigo
            case estrela.simbolo:
                posXestrela, posYestrela = x, y
                elementoAtual = estrela
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
    if novaPosY < 0 || novaPosY > len(mapa) || novaPosX < 0 || novaPosX > len(mapa[novaPosY]) {
        return
    }

    // Conflito
    if mapa[novaPosY][novaPosX].tangivel == true {

        if mapa[novaPosY][novaPosX].simbolo == '#' {
            // morrer()
        }
        return
    }

    mapa[posY][posX] = vazio
    posX, posY = novaPosX, novaPosY
    mapa[posY][posX] = personagem
}

func interagir() {
    statusMsg = fmt.Sprintf("Interagindo em (%d, %d)", posX, posY)
}

/*
func moverInimigo(mapa [][]Elemento, inimigo *Elemento, posX, posY int) {
    for {
        // Espera um tempo aleatório antes de mover o inimigo
        time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

        // Direções possíveis
        direcoes := []int{-1, 0, 1}

        // Escolhe uma direção aleatória
        dirX := direcoes[rand.Intn(len(direcoes))]
        dirY := direcoes[rand.Intn(len(direcoes))]

        // Nova posição do inimigo
        novaPosX := posX + dirX
        novaPosY := posY + dirY

        // Verifica se a nova posição é válida
        if novaPosY >= 0 && novaPosY < len(mapa) && novaPosX >= 0 && novaPosX < len(mapa[novaPosY]) &&
            mapa[novaPosY][novaPosX].tangivel == false {

            // Move o inimigo
            mapa[posY][posX].inimigo = false
            mapa[novaPosY][novaPosX].inimigo = true
            posX = novaPosX
            posY = novaPosY
        }
    }
}
*/