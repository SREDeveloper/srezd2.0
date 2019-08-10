package blockchain

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"strings"
)
// InitSre query the chaincode to get the state of Sre
func (setup *FabricSetup) InitSre(value string)(string,error) {

	// Prepare arguments
	cmd := strings.SplitN(value,"γ", 2)
	var args []string
	args = append(args, "invoke")
	args = append(args, "init")
	args = append(args, cmd[0])
	args = append(args, cmd[1])
	// Create a request (proposal) and send it
	//response, err := setup.client.Execute(channel.Request{ChaincodeID: setup.ChainCodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1]), []byte(args[2]), []byte(args[3])}, TransientMap: transientDataMap})
	response, err := setup.client.Execute(channel.Request{ChaincodeID: setup.ChainCodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1]), []byte(args[2]), []byte(args[3])}})
	if err != nil {
		return "", fmt.Errorf("failed to move funds: %v", err)
	}
	return string(response.Payload), nil
}