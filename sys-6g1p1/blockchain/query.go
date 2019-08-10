package blockchain

import (
	"fmt"
	"strings"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

// QuerySre query the chaincode to get the state of sre
func (setup *FabricSetup) QuerySre(value string) (string, error) {
	// Prepare arguments
	var args []string
	args = append(args, "invoke")
	args = append(args, "query")
	args = append(args, value)
	response, err := setup.client.Query(channel.Request{ChaincodeID: setup.ChainCodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1]), []byte(args[2])}})
	if err != nil {
		return "", fmt.Errorf("failed to srezd-service query: %v", err)
	}
	return string(response.Payload), nil
}

// QuerySreCC query the chaincode to get the state of sreCC
func (setup *FabricSetup) QuerySreCC(value string) (string, error) {
	cmd := strings.SplitN(value,"γ", 2)
	var args []string
	args = append(args, "invoke")
	args = append(args, "query")
	args = append(args, cmd[1])
	response, err := setup.client.Query(channel.Request{ChaincodeID: strings.TrimSpace(cmd[0]), Fcn: args[0], Args: [][]byte{[]byte(args[1]), []byte(args[2])}})
	if err != nil {
		return "", fmt.Errorf("failed to srezd-service query: %v", err)
	}
	return string(response.Payload), nil
}
