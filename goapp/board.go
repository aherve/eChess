package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
)

type BoardEvent = [16]byte
type Squares = [8][8]string

type Board struct {
	Connected bool
	Port      serial.Port
	Squares   Squares
}

func NewBoard() *Board {
	return &Board{
		Connected: false,
	}
}

func buildSquares(evt BoardEvent) Squares {
	board := buildEmptyBoard()
	for i := 0; i < 16; i += 2 {
		whiteByte := evt[i]
		blackByte := evt[i+1]

		for j := 0; j < 8; j++ {

			whiteBit := whiteByte & (1 << j)
			blackBit := blackByte & (1 << j)

			if whiteBit > 0 {
				board[i/2][j] = "W"
			} else if blackBit > 0 {
				board[i/2][j] = "B"
			}
		}
	}
	return board
}

func buildEmptyBoard() Squares {
	var squares Squares
	for i := 0; i < 8; i++ {
		thisRow := [8]string{}
		for j := 0; j < 8; j++ {
			thisRow[j] = "_"
		}
		squares[i] = thisRow
	}
	return squares
}

func (board *Board) StreamEvents(c chan Squares) {
	if !board.Connected {
		log.Fatal("Board is not connected!")
	}

	buff := []byte{}
	time.Sleep(1 * time.Second)
	for {
		newData := make([]byte, 128)
		n, err := board.Port.Read(newData)
		if err != nil {
			log.Fatalf("Error while reading from port: %v", err)
			break
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
				c <- squares
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

func (board *Board) Connect(c chan Squares) {
	okPortPrefixes := []string{"/dev/ttyUSB0", "/dev/tty.usbserial", "/dev/cu.usbserial"}

	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	for _, portName := range ports {

		for _, pref := range okPortPrefixes {
			if strings.HasPrefix(portName, pref) {
				fmt.Println("Connecting to board on port:", portName)

				port, err := serial.Open(portName, &serial.Mode{
					BaudRate: 115200,
				})
				if err != nil {
					log.Fatalf("Error while opening port: %v", err)
				}

				board.Connected = true
				board.Port = port
				log.Println("Connected to board on port", portName)
				go board.StreamEvents(c)
				return
			}
		}

	}
}
