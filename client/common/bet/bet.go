package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type Bet struct {
	Nombre     string
	Apellido   string
	DNI        string
	Nacimiento string
	Numero     string
}

func makeMsg(bet Bet, agency string) string {
	return fmt.Sprintf(
		"%s,%s,%s,%s,%s,%s\n",
		agency,
		bet.Nombre,
		bet.Apellido,
		bet.DNI,
		bet.Nacimiento,
		bet.Numero,
	)
}

func sendBet(conn net.Conn, bet Bet, agency string) (string, error) {
	message := makeMsg(bet, agency)
	messageLength := len(message)

	buffer := new(bytes.Buffer)

	fmt.Printf("Sending message length: %d\n", messageLength)

	// manda el long del msg en big endian
	if err := binary.Write(buffer, binary.BigEndian, uint16(messageLength)); err != nil {
		return "", fmt.Errorf("failed to write message length: %v", err)
	}

	// escribe lo que va a mandar en el buffer
	if _, err := buffer.Write([]byte(message)); err != nil {
		return "", fmt.Errorf("failed to write message data: %v", err)
	}

	// manda lo que esta en el buffer
	if _, err := conn.Write(buffer.Bytes()); err != nil {
		return "", fmt.Errorf("failed to send data: %v", err)
	}

	return messageLength, nil
}


func receiveConfirmation(conn net.Conn) (int, error) {
	lengthBytes := make([]byte, 2)
	if _, err := conn.Read(lengthBytes); err != nil {
		return 0, fmt.Errorf("failed to read confirmation length: %v", err)
	}

	var length uint16
	if err := binary.Read(bytes.NewReader(lengthBytes), binary.BigEndian, &length); err != nil {
		return 0, fmt.Errorf("failed to parse confirmation length: %v", err)
	}

	return int(length), nil
}