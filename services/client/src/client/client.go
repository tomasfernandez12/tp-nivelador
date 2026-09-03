package client

import (
	"bufio"
	"net"
	"os"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200

type Bet struct {
	AgencyId  int
	FirstName string
	LastName  string
	Document  int
	Birthdate string
	Number    int
}

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
	BatchSize  int
}

type Client struct {
	conn   net.Conn
	config ClientConfig
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{conn: conn, config: config}
	return client, nil
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
			continue
		}

		logger.Info(action, logger.Success)
		break
	}

	return conn, err
}

func (client *Client) Run() error {
	const mainAction = "process-bets"
	defer client.conn.Close()

	inputFile, err := os.Open(client.config.InputFile)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(client.config.OutputFile)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	bets := make([]Bet, 0, client.config.BatchSize)
	for messageId := 0; scanner.Scan(); messageId++ {
		messageArgs := []any{"agency-id", client.config.AgencyId, "message-id", messageId}
		logger.Info(mainAction, logger.InProgress, messageArgs...)

		agencyId, err := parseInteger([]byte(client.config.AgencyId))
		if err != nil {
			return err
		}
		bet, err := deserializeBet(scanner.Bytes(), agencyId)
		if err != nil {
			return err
		}
		bets = append(bets, bet)
		if len(bets) == client.config.BatchSize {
			if err := sendBets(client.conn, bets); err != nil {
				logger.Error("send-message", logger.Fail, messageArgs...)
				return err
			}
			bets = bets[:0]
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if len(bets) > 0 {
		if err := sendBets(client.conn, bets); err != nil {
			return err
		}
	}
	if err := sendMessage(client.conn, nil); err != nil {
		return err
	}
	for {
		message, err := receiveMessage(client.conn)
		if err != nil {
			return err
		}
		if message.Size == 0 {
			break
		}
		bet, err := deserializeSerializedBet(message.Payload)
		if err != nil {
			return err
		}
		output := serializeOutput(bet)
		if _, err := outputFile.Write(output); err != nil {
			return err
		}
	}
	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)

	return nil
}
