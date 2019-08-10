package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// HeroesServiceChaincode implementation of Chaincode
type HeroesServiceChaincode struct {
}

type GdSreBlock struct {
	ObjectType  string `json:"docType"`     //docType is used to distinguish the various types of objects in state database
	SreNo    string `json:"SreNo"`    //一物一码
	NoState     string `json:"NoState"`     //商品状态（0-未激活；1-激活（上链）；2-发货、物流、仓储；3-零售、4-购买、5-展示）
	FactoryName string `json:"FactoryName"` //公司名称
	FactoryAdd  string `json:"FactoryAdd"`  //公司地址	
	ShopName    string `json:"ShopName"`    //商品名称
	ShopType    string `json:"ShopType"`    //商品类型	
	ShopNote    string `json:"ShopNote"`    //商品介绍
	ShopValue    string `json:"ShopValue"`  //商品动态添加项
	OptMan    string `json:"OptMan"`    //操作人
	OptAdd   string `json:"OptAdd"`  //操作地址
	OptDev   string `json:"OptDev"`  //操作设备
}

// Init of the chaincode
// This function is called only one when the chaincode is instantiated.
// So the goal is to prepare the ledger to handle future requests.
func (t *HeroesServiceChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### HeroesServiceChaincode Init ###########")
	// Get the function and arguments from the request
	function, _ := stub.GetFunctionAndParameters()

	// Check if the request is the init function
	if function != "init" {
		return shim.Error("Unknown function call")
	}
	//err := stub.PutState("hello", []byte("world"))
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	// Return a successful message
	return shim.Success(nil)
}

// Invoke
// All future requests named invoke will arrive here.
func (t *HeroesServiceChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### HeroesServiceChaincode Invoke ###########")

	// Get the function and arguments from the request
	function, args := stub.GetFunctionAndParameters()

	// Check whether it is an invoke request
	if function != "invoke" {
		return shim.Error("Unknown function call")
	}

	// Check whether the number of arguments is sufficient
	if len(args) < 1 {
		return shim.Error("The number of arguments is insufficient.")
	}

	// In order to manage multiple type of request, we will check the first argument.
	// Here we have one possible argument: query (every query request will read in the ledger without modification)
	if args[0] == "query" {
		//return t.query(stub, args)
		return t.getHistoryForMarble(stub, args)
	}

	// The update argument will manage all update in the ledger
	if args[0] == "invoke" {
		return t.invoke(stub, args)
	}
	if args[0] == "init" {
		// setting an entity state
		return t.init(stub, args)
	}

	// If the arguments given don’t match any function, we return an error
	return shim.Error("Unknown action, check the first argument")
}
func (t *HeroesServiceChaincode) init(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	SreNo := args[1]
	cmd := strings.SplitN(args[2], "γ", -1)	
	NoState := cmd[0]	
	FactoryName := cmd[1]
	FactoryAdd := cmd[2]
	ShopName := cmd[3]	
	ShopType := cmd[4]	
	ShopNote := cmd[5]
	ShopValue := cmd[6]
	OptMan := ""
	OptAdd := ""
	OptDev := ""

	// ==== Check if marble already exists ====
	hoerAsBytes, err := stub.GetState(SreNo)
	if err != nil {
		return shim.Error("Failed to get sre: " + err.Error())
	} else if hoerAsBytes != nil {
		fmt.Println("This sre already exists: " + SreNo)
		return shim.Error("This sre already exists: " + SreNo)
	}
	// ==== Create marble object and marshal to JSON ====
	objectType := "GdSreBlock"
	GdSreBlock := &GdSreBlock{objectType, SreNo, NoState, ShopName, ShopType, FactoryName, FactoryAdd,ShopValue,ShopNote,OptMan,OptAdd,OptDev}
	GdSreJSONasBytes, err := json.Marshal(GdSreBlock)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(SreNo, GdSreJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

// invoke
// Every functions that read and write in the ledger will be here
func (t *HeroesServiceChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("########### HeroesServiceChaincode invoke ###########")

	if len(args) < 2 {
		return shim.Error("The number of arguments is insufficient.")
	}
//cmd := strings.SplitN(args[2], "γ", -1)	
//	NoState := cmd[0]
//	OptMan := cmd[1]
//	OptAdd := cmd[2]
//	OptDev := cmd[3]
	sreAsBytes, err := stub.GetState(args[1])
	if err != nil {
		return shim.Error("Failed to get sre: " + err.Error())
	} else if sreAsBytes != nil {	
		err = stub.PutState(args[1], []byte(args[2]))
		if err != nil {
			return shim.Error("Failed to update state of sre")
		}

		// Notify listeners that an event "eventInvoke" have been executed (check line 19 in the file invoke.go)
		err = stub.SetEvent("eventInvoke", []byte{})
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	}
	return shim.Error("This sreno not exists")

}
func (t *HeroesServiceChaincode) getHistoryForMarble(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	//marbleName := args[1]

	//fmt.Printf("- start getHistoryForMarble: %s\n", marbleName)

	resultsIterator, err := stub.GetHistoryForKey(args[1])
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString("￠")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForMarble returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// ===============================================
// readMarble - read a marble from chaincode state
// ===============================================
func (t *HeroesServiceChaincode) readHoer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the marble to query")
	}
	name = args[1]
	valAsbytes, err := stub.GetState(name) //get the hoer from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"hoer does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valAsbytes)
}
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

// query
// Every readonly functions in the ledger will be here
func (t *HeroesServiceChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("########### HeroesServiceChaincode query ###########")
	// Check whether the number of arguments is sufficient
	if len(args) < 2 {
		return shim.Error("The number of arguments is insufficient.")
	}
	// Get the state of the value matching the key hello in the ledger
	state, err := stub.GetState(args[1])
	if err != nil {
		return shim.Error("Failed to get state of hello")
	}
	return shim.Success(state)

	/*
	resultsIterator, err := stub.GetQueryResult(args[1])
	if err != nil {
		return shim.Error("Failed to get state of chaincode")
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Failed to get state of chaincode")
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
	*/
	//return buffer.Bytes(), nil
}

func main() {
	// Start the chaincode and make it ready for futures requests
	err := shim.Start(new(HeroesServiceChaincode))
	if err != nil {
		fmt.Printf("Error starting Heroes Service chaincode: %s", err)
	}
}
