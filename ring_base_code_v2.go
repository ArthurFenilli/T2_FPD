package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmação da eleição)
	corpo [4]int // conteúdo da mensagem para colocar os ids (usar um tamanho compatível com o número de processos no anel)
}

var (
	chans = []chan mensagem{ // vetor de canais para formar o anel de eleição
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	controle = make(chan int)
	wg       sync.WaitGroup // wg é usado para esperar que o programa termine
)

func ElectionControler(in chan int) {
	defer wg.Done()

	var temp mensagem

	rand.Seed(time.Now().UnixNano())

	failedProcess := rand.Intn(len(chans))

	electionProcess := rand.Intn(len(chans))
	for electionProcess == failedProcess {
		electionProcess = rand.Intn(len(chans))
	}

	temp.tipo = 2
	chans[(failedProcess+3)%4] <- temp
	fmt.Printf("Controle: mudar o processo %d para falho\n", failedProcess)

	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// Iniciar eleição
	temp.tipo = 1
	chans[(electionProcess+3)%4] <- temp
	fmt.Printf("Controle: iniciar eleição no processo %d\n", electionProcess)

	fmt.Println("\n   Processo controlador concluído\n")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()

	// Variáveis locais que indicam se este processo é o líder e se está ativo
	var actualLeader int
	var bFailed bool = false // todos iniciam sem falha

	actualLeader = leader // indicação do líder veio por parâmetro

	for true {
		temp := <-in // ler mensagem
		fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.corpo[3])

		switch temp.tipo {
		case 1:
			if temp.corpo[(TaskId+3)%4] == 1 {
				temp.tipo = 4
				for i := 0; i < len(temp.corpo); i++ {
					if temp.corpo[(i+3)%4] == 1 {
						actualLeader = i
						break
					}
				}
				temp.corpo[0] = actualLeader
				out <- temp
			} else {
				if !bFailed {
					temp.corpo[(TaskId+3)%4] = 1
					out <- temp
				} else {
					temp.corpo[(TaskId+3)%4] = -1
					out <- temp
				}
			}
		case 2:
			bFailed = true
			fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
			fmt.Printf("%2d: líder atual %d\n", TaskId, actualLeader)
			controle <- -5
		case 3:
			bFailed = false
			fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
			fmt.Printf("%2d: líder atual %d\n", TaskId, actualLeader)
			controle <- -5

		case 4:
			actualLeader = temp.corpo[0]
			out <- temp
			wg.Done()
		default:
			fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
			fmt.Printf("%2d: líder atual %d\n", TaskId, actualLeader)
		}
	}

	fmt.Printf("%2d: terminei \n", TaskId)
}

func main() {
	wg.Add(5) // Adicionar uma contagem de cinco, um para cada goroutine

	// Criar os processos do anel de eleição
	go ElectionStage(0, chans[3], chans[0], 0) // este é o líder
	go ElectionStage(1, chans[0], chans[1], 0) // não é líder, é o processo 0
	go ElectionStage(2, chans[1], chans[2], 0) // não é líder, é o processo 0
	go ElectionStage(3, chans[2], chans[3], 0) // não é líder, é o processo 0

	fmt.Println("\n   Anel de processos criado")

	// Criar o processo controlador
	go ElectionControler(controle)

	fmt.Println("\n   Processo controlador criado\n")

	wg.Wait() // Esperar as goroutines terminarem
}
