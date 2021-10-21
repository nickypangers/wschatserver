package server

import (
	"log"
	"strings"
)

const (
	ChangeAddress string = "/address"
)

func (c *Client) processCommand(message string) bool {
	messageArr := strings.Split(message, " ")
	if len(messageArr) == 0 {
		return false
	}
	command := messageArr[0]
	arg := strings.Trim(messageArr[1], " ")
	if arg == "" {
		return false
	}

	switch command {
	case ChangeAddress:
		log.Printf("%s change address to %s", c.Address, arg)
		c.Address = arg
		return true
	default:
		log.Printf("no such command: %s", command)
		return false
	}
}
