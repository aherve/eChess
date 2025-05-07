package main

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/notnil/chess"
	"go.bug.st/serial"
)

type BoardEvent = [16]byte
type BoardState = [8][8]chess.Color

type Board struct {
	Connected bool
	Port      serial.Port
	State     BoardState
	mu        *sync.Mutex
}

func NewBoard() *Board {
	return &Board{
		Connected: false,
		mu:        &sync.Mutex{},
	}
}

func (board *Board) Update(squares BoardState) {
	board.mu.Lock()
	defer board.mu.Unlock()
	for i := range squares {
		for j := range squares[i] {
			board.State[i][j] = squares[i][j]
		}
	}
}

func buildSquares(evt BoardEvent) BoardState {
	board := BoardState{}
	for i := 0; i < 16; i += 2 {
		whiteByte := evt[i]
		blackByte := evt[i+1]

		for j := range 8 {

			whiteBit := whiteByte & (1 << j)
			blackBit := blackByte & (1 << j)

			if whiteBit > 0 {
				board[j][i/2] = chess.White
			} else if blackBit > 0 {
				board[j][i/2] = chess.Black
			} else {
				board[j][i/2] = chess.NoColor
			}
		}
	}
	return board
}

func (board *Board) Listen(c chan bool) {
	if !board.Connected {
		log.Fatal("Board is not connected!")
	}

	const maxBufferSize = 1024 // Maximum size of the buffer in bytes
	buff := make([]byte, 0, maxBufferSize)
	time.Sleep(1 * time.Second)
	for {
		newData := make([]byte, 128)
		n, err := board.Port.Read(newData)
		if err != nil {
			log.Fatalf("Error while reading from port: %v", err)
			break
		}

		// If adding new data would exceed maxBufferSize, remove old data
		if len(buff)+n > maxBufferSize {
			// Keep only the last maxBufferSize/2 bytes
			keepFrom := len(buff) - maxBufferSize/2
			if keepFrom < 0 {
				keepFrom = 0
			}
			buff = buff[keepFrom:]
			log.Printf("Buffer overflow, trimming to %d bytes", len(buff))
		}

		buff = append(buff, newData[:n]...)
		if len(buff) < 19 {
			continue
		}

		for i := len(buff) - 1; i > 3; i-- {
			if buff[i] == 0xFF && buff[i-1] == 0xFF && buff[i-2] == 0xFF {
				if i < 16 {
					log.Println("discarding incomplete message")
					buff = buff[i+1:]
					break
				}
				msg := BoardEvent{}
				copy(msg[:], buff[i-18:i-2])
				squares := buildSquares(msg)

				board.Update(squares)

				c <- true
				if len(buff) > i+1 {
					buff = buff[i+1:]
				} else {
					buff = []byte{}
				}
				break
			}
		}
	}
}

func (board *Board) Connect(c chan bool) {
	board.mu.Lock()
	defer board.mu.Unlock()

	okPortPrefixes := []string{"/dev/ttyUSB0", "/dev/tty.usbserial", "/dev/cu.usbserial"}

	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Println("No serial ports found!")
		return
	}
	for _, portName := range ports {

		for _, pref := range okPortPrefixes {
			if strings.HasPrefix(portName, pref) {
				log.Println("Connecting to board on port:", portName)

				port, err := serial.Open(portName, &serial.Mode{
					BaudRate: 115200,
				})
				if err != nil {
					log.Fatalf("Error while opening port: %v", err)
				}

				board.Connected = true
				board.Port = port
				log.Println("Connected to board on port", portName)
				go board.Listen(c)
				return
			}
		}

	}
}

func (board *Board) sendLEDCommand(litSquares map[int8]bool) {
	board.mu.Lock()
	defer board.mu.Unlock()

	command := make([]byte, len(litSquares)+2)
	command[0] = 0xFE
	command[len(command)-1] = 0xFF

	pos := 0
	for k := range litSquares {
		i, j := getCoordinatesFromIndex(k)
		command[pos+1] = byte((j << 4) + i)
		pos++
	}

	_, err := board.Port.Write(command)
	if err != nil {
		log.Fatalf("Error while writing to port: %v", err)
	}

}
