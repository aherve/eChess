package main

import (
	"fmt"
	"log"
	"strings"

	"go.bug.st/serial"
)

type BoardEvent = [16]byte

type Board struct {
	Connected bool
	Port      serial.Port
	Squares   [64][64]string
}

func NewBoard() *Board {
	return &Board{
		Connected: false,
	}
}

func (board *Board) StreamEvents(c chan BoardEvent) {
	if !board.Connected {
		log.Fatal("Board is not connected!")
	}

	buff := make([]byte, 100)
	for {
		n, err := board.Port.Read(buff)
		if err != nil {
			log.Fatal(err)
			break
		}
		if n == 0 {
			break
		}

		if n != 19 {
			log.Println("Invalid data length:", n)
			continue
		}
		if (buff[0] != 0xFF) || (buff[18] != 0xFF) || buff[17] != 0xFF {
			log.Println("Invalid start or end byte")
			continue
		}
		msg := BoardEvent{}
		copy(msg[:], buff[1:17])
		c <- msg
	}
}

func (board *Board) Connect(c chan BoardEvent) {
	okPortPrefixes := []string{"/dev/ttyUSB0", "/dev/tty.usbserial", "/dev/cu.usbserial"}

	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	for _, port := range ports {

		for _, pref := range okPortPrefixes {
			if strings.HasPrefix(port, pref) {
				fmt.Println("Connecting to board on port:", port)

				port, err := serial.Open(port, &serial.Mode{
					BaudRate: 115200,
				})
				if err != nil {
					log.Fatal(err)
				}

				board.Connected = true
				board.Port = port
				log.Println("Connected to board on port", port)
				go board.StreamEvents(c)
				return
			}
		}

	}
}
