package bet

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

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

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func writeFull(conn net.Conn, data []byte) error {
	totalWritten := 0
	for totalWritten < len(data) {
		n, err := conn.Write(data[totalWritten:])
		if err != nil {
			return err
		}
		totalWritten += n
	}
	return nil
}

func SendBatch(conn net.Conn, bets []Bet, agency string, lastBatch bool) error {
	
	var batchMessage strings.Builder
	
	if len(bets) == 0{
		return nil
	}
	
	for _, bet := range bets {
		batchMessage.WriteString(makeMsg(bet, agency))
	}
	
	if lastBatch {
		batchMessage.WriteString("END\n")
	}
	
	message := batchMessage.String()
	messageLength := len(message)
	
	if messageLength > 0xFFFF {
		return fmt.Errorf("batch message too large, exceeds maximum size of 65535 bytes")
	}
	
	buffer := new(bytes.Buffer)
	
	if err := binary.Write(buffer, binary.BigEndian, uint16(messageLength)); err != nil {
		return fmt.Errorf("failed to write message length: %v", err)
	}
	
	written, err := buffer.Write([]byte(message))
	
	if err != nil {
		return fmt.Errorf("failed to write batch message: %v", err)
	}
	
	if written != len(message) {
		return fmt.Errorf("failed to write full batch message: wrote %d bytes, expected %d", written, len(message))
	}
	
	if err := writeFull(conn, buffer.Bytes()); err != nil {
		return fmt.Errorf("failed to send batch: %v", err)
	}
	
	lenReceived, err := ReceiveConfirmation(conn)
	
	if err != nil || lenReceived != len(bets) {
		return fmt.Errorf("message length received (%d) is not equal to expected (%d) or error occurred: %v", lenReceived, len(bets), err)
	}
	
	if lastBatch {
		winners, err := ReceiveWinners(conn)
		if err != nil {
			return fmt.Errorf("failed to receive winners: %v", err)
		}
		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", len(winners))
	}
	return nil
}

func ProcessFileFromReader(conn net.Conn, agency string, r io.Reader, maxBatchSize int) error {
	scanner := bufio.NewScanner(r)
	var currentBatch []Bet
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) != 5 {
			return fmt.Errorf("invalid bet format")
		}
		currentBatch = append(currentBatch, Bet{
			Nombre:     parts[0],
			Apellido:   parts[1],
			DNI:        parts[2],
			Nacimiento: parts[3],
			Numero:     parts[4],
		})
		if len(currentBatch) == maxBatchSize {
			if err := SendBatch(conn, currentBatch, agency, false); err != nil {
				return err
			}
			currentBatch = currentBatch[:0]
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if len(currentBatch) > 0 {
		if err := SendBatch(conn, currentBatch, agency, true); err != nil {
			return err
		}
	}
	return nil
}

func ReceiveConfirmation(conn net.Conn) (int, error) {
	lengthBytes := make([]byte, 2)
	if sizeRead, err := io.ReadFull(conn, lengthBytes); err != nil || sizeRead != 2 {
		return 0, fmt.Errorf("failed to read confirmation length: %v", err)
	}
	
	var length uint16
	if err := binary.Read(bytes.NewReader(lengthBytes), binary.BigEndian, &length); err != nil {
		return 0, fmt.Errorf("failed to parse confirmation length: %v", err)
	}
	
	return int(length), nil
}


func ReceiveWinners(conn net.Conn) ([]int, error) {
	countBytes := make([]byte, 4)
	if sizeRead, err := io.ReadFull(conn, countBytes); err != nil || sizeRead != 4 {
		return []int{}, nil
	}
	
	var count uint32
	if err := binary.Read(bytes.NewReader(countBytes), binary.BigEndian, &count); err != nil {
		return nil, fmt.Errorf("failed to parse winners count: %v", err)
	}
	
	
	totalBytes := int(count) * 4
	documentBytes := make([]byte, totalBytes)
	if sizeRead, err := io.ReadFull(conn, documentBytes); err != nil || sizeRead != totalBytes {
		return nil, fmt.Errorf("failed to read winners documents: %v", err)
	}
	
	winners := make([]int, count)
	for i := 0; i < int(count); i++ {
		winners[i] = int(binary.BigEndian.Uint32(documentBytes[i*4 : (i+1)*4]))
	}
	
	return winners, nil
}
