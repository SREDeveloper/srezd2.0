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
		return "", fmt.Errorf("failed to query: %v", err)
	}
	return string(response.Payload), nil
}

// QuerySreCC query the chaincode to get the state of sreCC
func (setup *FabricSetup) QuerySreCC(value string) (string, error) {
	// Prepare arguments
	var args []string
	args = append(args, "invoke")
	args = append(args, "query")
	args = append(args, value)
	response, err := setup.client.Query(channel.Request{ChaincodeID: setup.ChainCodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1]), []byte(args[2])}})
	if err != nil {
		return "", fmt.Errorf("failed to srezd-service query: %v", err)
	}
	if strings.EqualFold(string(response.Payload),"[]") {
		for _, ccid := range setup.ChainCodeListID {
			if strings.EqualFold(ccid,setup.ChainCodeID) {
				continue
			}else{
				response, err = setup.client.Query(channel.Request{ChaincodeID: "srezd1-service", Fcn: args[0], Args: [][]byte{[]byte(args[1]), []byte(args[2])}})
				if err != nil {
					return "", fmt.Errorf("failed to srezd1-service query: %v", err)
				}
			}
		}
	}
	return string(response.Payload), nil
}
