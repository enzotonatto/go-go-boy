# Go-Go-Boy Multiplayer

Nós desenvolvemos um jogo de labirinto 2D executado diretamente no terminal usando Go. O desafio central do jogo é navegar pelo labirinto, coletando a estrela (•) enquanto evita o inimigo (Ω) que sempre rastreia o personagem principal. A localização da estrela é alterada periodicamente, tornando o jogo ainda mais desafiador. Ambas as entidades, o inimigo e a estrela, são controladas por rotinas independentes (goroutines).

## Elementos do jogo
* Personagem (☻): O protagonista que você controla pelo labirinto.
* Parede (█): Obstáculos fixos que bloqueiam o caminho do personagem.
* Inimigo (Ω): Uma entidade que segue o personagem pelo labirinto.
* Estrela (•): O objetivo principal, muda de localização periodicamente.
* Portal (0): Teleporta o personagem para outra parte do labirinto.

## Como jogar

### Objetivo: Navegue pelo labirinto para coletar a estrela enquanto evita o inimigo.
#### Movimentação/Interação:
#### Utilize as teclas:
* W: Mover para cima
* A: Mover para a esquerda
* S: Mover para baixo
* D: Mover para a direita
#### Interagir: Pressione a tecla 'e' para realizar ações especiais, como congelar o inimigo e a estrela.
- Vitória: Colete a estrela para completar o jogo.
- Derrota: Evite o inimigo, pois o contato resultará no fim do jogo.

## Instructions

### Building and Running

1. **Start the Server**

```sh
cd server
go run server.go
```

The server will start and listen for incoming client connections on port 1234.

2. **Start the Client**

```sh
cd client
go run client.go
```

The client will connect to the server, register itself, and begin sending commands and receiving game state updates.

### File Descriptions

- `server/server.go`: Contains the implementation of the GameServer.
- `client/client.go`: Contains the implementation of the GameClient.
- `mapa.txt`: Contains the map for the game.
- `Makefile`: Build automation tool.
- `.gitignore`: Git ignore file.
