package client

import (
	"encoding/binary"
	"io"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const (
	messageHeaderSize = 4
	integerSize       = 4
	maxMessageSize    = 1024 * 1024
)

type Message struct {
	Size    uint32
	Payload []byte
}

func serializeBet(bet Bet) ([]byte, error) {
	payload := make([]byte, 0)
	payload = appendInteger(payload, bet.AgencyId)
	payload = appendString(payload, bet.FirstName)
	payload = appendString(payload, bet.LastName)
	payload = appendInteger(payload, bet.Document)
	payload = appendString(payload, bet.Birthdate)
	payload = appendInteger(payload, bet.Number)
	return payload, nil
}

func serializeBets(bets []Bet) ([]byte, error) {
	payload := make([]byte, 0)
	for _, bet := range bets {
		serializedBet, err := serializeBet(bet)
		if err != nil {
			return nil, err
		}
		payload = append(payload, serializedBet...)
	}
	return payload, nil
}

func sendBets(socket io.Writer, bets []Bet) error {
	payload, err := serializeBets(bets)
	if err != nil {
		return err
	}
	return sendMessage(socket, payload)
}

func parseInteger(value []byte) (int, error) {
	if len(value) == 0 {
		return 0, io.ErrUnexpectedEOF
	}
	number := 0
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return 0, io.ErrUnexpectedEOF
		}
		number = number*10 + int(digit-'0')
	}
	return number, nil
}

func splitBetLine(line []byte) ([][]byte, error) {
	fields := make([][]byte, 0, 5)
	start := 0
	for index, character := range line {
		if character == ',' {
			fields = append(fields, line[start:index])
			start = index + 1
		}
	}
	fields = append(fields, line[start:])
	if len(fields) != 5 {
		return nil, io.ErrUnexpectedEOF
	}
	return fields, nil
}

func deserializeBet(line []byte, agencyId int) (Bet, error) {
	fields, err := splitBetLine(line)
	if err != nil {
		return Bet{}, err
	}
	document, err := parseInteger(fields[2])
	if err != nil {
		return Bet{}, err
	}
	number, err := parseInteger(fields[4])
	if err != nil {
		return Bet{}, err
	}
	return Bet{agencyId, string(fields[0]), string(fields[1]), document, string(fields[3]), number}, nil
}

func serializeOutput(bet Bet) []byte {
	output := make([]byte, 0)
	output = append(output, []byte(bet.FirstName)...)
	output = append(output, ',')
	output = append(output, []byte(bet.LastName)...)
	output = append(output, ',')
	output = appendDecimal(output, bet.Document)
	output = append(output, ',')
	output = append(output, []byte(bet.Birthdate)...)
	output = append(output, ',')
	output = appendDecimal(output, bet.Number)
	return append(output, '\n')
}

func appendDecimal(output []byte, value int) []byte {
	if value == 0 {
		return append(output, '0')
	}
	digits := make([]byte, 0)
	for value > 0 {
		digits = append(digits, byte(value%10)+'0')
		value /= 10
	}
	for index := len(digits) - 1; index >= 0; index-- {
		output = append(output, digits[index])
	}
	return output
}

func appendInteger(payload []byte, value int) []byte {
	encoded := make([]byte, integerSize)
	binary.BigEndian.PutUint32(encoded, uint32(value))
	return append(payload, encoded...)
}

func appendString(payload []byte, value string) []byte {
	encoded := []byte(value)
	payload = appendInteger(payload, len(encoded))
	return append(payload, encoded...)
}

func readInteger(payload []byte, position *int) (int, error) {
	if len(payload)-*position < integerSize {
		return 0, io.ErrUnexpectedEOF
	}
	value := binary.BigEndian.Uint32(payload[*position : *position+integerSize])
	*position += integerSize
	return int(value), nil
}

func readString(payload []byte, position *int) (string, error) {
	length, err := readInteger(payload, position)
	if err != nil || len(payload)-*position < length {
		return "", io.ErrUnexpectedEOF
	}
	value := string(payload[*position : *position+length])
	*position += length
	return value, nil
}

func deserializeSerializedBet(payload []byte) (Bet, error) {
	position := 0
	agencyId, err := readInteger(payload, &position)
	if err != nil {
		return Bet{}, err
	}
	firstName, err := readString(payload, &position)
	if err != nil {
		return Bet{}, err
	}
	lastName, err := readString(payload, &position)
	if err != nil {
		return Bet{}, err
	}
	document, err := readInteger(payload, &position)
	if err != nil {
		return Bet{}, err
	}
	birthdate, err := readString(payload, &position)
	if err != nil {
		return Bet{}, err
	}
	number, err := readInteger(payload, &position)
	if err != nil {
		return Bet{}, err
	}
	return Bet{agencyId, firstName, lastName, document, birthdate, number}, nil
}

func sendMessage(socket io.Writer, payload []byte) error {
	header := make([]byte, messageHeaderSize)
	binary.BigEndian.PutUint32(header, uint32(len(payload)))
	if err := safe_socket.SendAll(socket, header); err != nil {
		return err
	}
	return safe_socket.SendAll(socket, payload)
}

func receiveMessage(socket io.Reader) (Message, error) {
	header, err := safe_socket.RecvAll(socket, messageHeaderSize)
	if err != nil {
		return Message{}, err
	}
	size := binary.BigEndian.Uint32(header)
	if size > maxMessageSize {
		return Message{}, io.ErrUnexpectedEOF
	}
	payload, err := safe_socket.RecvAll(socket, int(size))
	return Message{Size: size, Payload: payload}, err
}