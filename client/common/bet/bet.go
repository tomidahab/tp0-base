package bet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")


// Bet representa una apuesta.
type Bet struct {
	Nombre     string
	Apellido   string
	DNI        string
	Nacimiento string
	Numero     string
}

// makeMsg arma el mensaje individual de una apuesta.
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

// SendBatch envía un batch de apuestas al servidor.
/*
func SendBatch(conn net.Conn, bets []Bet, agency string, lastBatch bool) error {
	const maxMessageSize = 8192 
	const endSize = len("END\n")

	var batchMessage strings.Builder
	var batchMessageFuture []Bet
	var betsSent []Bet
	currentSize := 0

	if len(bets) == 0{
		return nil
	}

	for _, bet := range bets {
		betMessage := makeMsg(bet, agency)
		betSize := len(betMessage)
		
		if currentSize+betSize+(endSize*boolToInt(lastBatch)) > maxMessageSize {
			batchMessageFuture = append(batchMessageFuture, bet)
		} else {
			batchMessage.WriteString(betMessage)
			betsSent = append(betsSent, bet)
			currentSize += betSize
		}
	}

	if lastBatch && len(batchMessageFuture) == 0 {
		batchMessage.WriteString("END\n")
	}
	

	message := batchMessage.String()
	messageLength := len(message)

	// Nunca deberia llegar aca pero lo dejo x las dudas
	if messageLength > 0xFFFF {
		return fmt.Errorf("batch message too large, exceeds maximum size of 65535 bytes")
	}

	buffer := new(bytes.Buffer)

	if err := binary.Write(buffer, binary.BigEndian, uint16(messageLength)); err != nil {
		return fmt.Errorf("failed to write message length: %v", err)
	}

	if _, err := buffer.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write batch message: %v", err)
	}

	if _, err := conn.Write(buffer.Bytes()); err != nil {
		return fmt.Errorf("failed to send batch: %v", err)
	}

	len_recieved, err := ReceiveConfirmation(conn)

	if  len_recieved != messageLength && err != nil {
		return fmt.Errorf("message lenght received is not equal to real one")
	}


	for _, bet := range betsSent {
		log.Infof("action: apuesta_enviada | result: success | dni: %s | numero: %s", bet.DNI, bet.Numero)
	}

	return SendBatch(conn, batchMessageFuture,agency,lastBatch)
}
*/
func SendBatch(conn net.Conn, bets []Bet, agency string, lastBatch bool) error {

	// log.Infof("DELETE sending batcg")


	var batchMessage strings.Builder

	if len(bets) == 0{
		return nil
	}

	for _, bet := range bets {
		betMessage := makeMsg(bet, agency)
		batchMessage.WriteString(betMessage)
	}

	if lastBatch {
		batchMessage.WriteString("END\n")
	}
	
	message := batchMessage.String()
	messageLength := len(message)

	// Nunca deberia llegar aca pero lo dejo x las dudas
	if messageLength > 0xFFFF {
		return fmt.Errorf("batch message too large, exceeds maximum size of 65535 bytes")
	}

	buffer := new(bytes.Buffer)

	//log.Infof("DELETE writing data to send in bigendian: %v", messageLength)


	if err := binary.Write(buffer, binary.BigEndian, uint16(messageLength)); err != nil {
		return fmt.Errorf("failed to write message length: %v", err)
	}

	//log.Infof("DELETE sent number in bigendian: %v", messageLength)


	if _, err := buffer.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write batch message: %v", err)
	}

	//log.Infof("DELETE sending buffer of : %v", len(bets))


	if _, err := conn.Write(buffer.Bytes()); err != nil {
		return fmt.Errorf("failed to send batch: %v", err)
	}

	//log.Infof("DELETE sended buffer of : %v", len(bets))

	len_recieved, err := ReceiveConfirmation(conn)

	//log.Infof("DELETE recieve conf that server received : %v", len_recieved)


	if err != nil || len_recieved != len(bets) {
		return fmt.Errorf("message length received (%d) is not equal to expected (%d) or error occurred: %v", len_recieved, messageLength, err)
	}

	if lastBatch {
		//log.Infof("DELETE this is the last batch, so we are waiting for winners")

		winners, err := ReceiveWinners(conn)
		//log.Infof("DELETE received winner")

		if err != nil {
			return fmt.Errorf("failed to receive winners: %v", err)

		}
		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", len(winners))
	}


	

	/*for _, bet := range bets {
		log.Infof("action: apuesta_enviada | result: success | dni: %s | numero: %s", bet.DNI, bet.Numero) COMMENT FOR NOW
	}*/

	return nil
}



// ProcessFile procesa el archivo de texto y envía las apuestas en batchs.
func ProcessFile(conn net.Conn, agency string, fileContent string, maxBatchSize int, loopTime time.Duration) error {
	lines := strings.Split(strings.TrimSpace(fileContent), "\n")
	totalBets := len(lines)

	//log.Infof("DELETE read file")

	if totalBets == 0 {
		return fmt.Errorf("file is empty or has no valid bets")
	}

	//log.Infof("DELETE totalBets: %v", totalBets)

	var currentBatch []Bet
	for i, line := range lines {
		//log.Infof("DELETE for n: %v, current batch size: %v", i, len(currentBatch))

		parts := strings.Split(line, ",")
		if len(parts) != 5 {
			return fmt.Errorf("invalid bet format on line %d", i+1)
		}

		bet := Bet{
			Nombre:     parts[0],
			Apellido:   parts[1],
			DNI:        parts[2],
			Nacimiento: parts[3],
			Numero:     parts[4],
		}

		currentBatch = append(currentBatch, bet)

		if len(currentBatch) == maxBatchSize || i == totalBets-1 {
			//log.Infof("DELETE reached maxbatchsize: %v", totalBets)
			lastBatch := (i == totalBets-1)
			//time.Sleep(time.Duration(loopTime) * time.Millisecond)
			if err := SendBatch(conn, currentBatch, agency, lastBatch); err != nil {
				return fmt.Errorf("failed to send batch: %v", err)
			}
			currentBatch = []Bet{}
		}
	}

	return nil
}

func ReceiveConfirmation(conn net.Conn) (int, error) {
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


func ReceiveWinners(conn net.Conn) ([]int, error) {
    // Leo 4 bytes del n de ganadores
    countBytes := make([]byte, 4)
	//log.Infof("DELETE waiting for n of winners")

    if _, err := conn.Read(countBytes); err != nil {
		return []int{}, nil
        //return nil, fmt.Errorf("failed to read winners count: %v", err)
    }

    var count uint32
    if err := binary.Read(bytes.NewReader(countBytes), binary.BigEndian, &count); err != nil {
        return nil, fmt.Errorf("failed to parse winners count: %v", err)
    }

	//log.Infof("DELETE read n of winners %v", count)


    // Leo los dni de los ganadores (4 bytes por documento)
    totalBytes := int(count) * 4
    documentBytes := make([]byte, totalBytes)
	//log.Infof("DELETE waiting for documents of winners")

    if _, err := conn.Read(documentBytes); err != nil {
        return nil, fmt.Errorf("failed to read winners documents: %v", err)
    }

    // Convertir los bytes de los dni a enteros
    winners := make([]int, count)
    for i := 0; i < int(count); i++ {
        winners[i] = int(binary.BigEndian.Uint32(documentBytes[i*4 : (i+1)*4]))
    }

    return winners, nil
}
