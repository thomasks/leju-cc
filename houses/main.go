package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Chaincode comment
type Chaincode struct {
}

//CryptoDescriptor comment
type CryptoDescriptor struct {
	Level        string   `json:"level"`
	CryptoFields []string `json:"cryptoFields"`
}

//{"Args":["query","key"]}'
func (t *Chaincode) query(stub shim.ChaincodeStubInterface, key string) pb.Response {
	fmt.Printf("query %s\n", key)
	bytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("query fail " + err.Error())
	}
	return shim.Success(bytes)
}

//{"Args":["write","key","value"]}'
func (t *Chaincode) write(stub shim.ChaincodeStubInterface, key, value string) pb.Response {
	fmt.Printf("write %s, value is %s\n", key, value)
	if err := stub.PutState(key, []byte(value)); err != nil {
		return shim.Error("write fail " + err.Error())
	}
	return shim.Success(nil)
}

//{"Args":["writeMultiSegData","key","value",SegDescriptor]}
func (t *Chaincode) writeMultiSegData(stub shim.ChaincodeStubInterface, key, value, cryptoDescriptor string) pb.Response {
	fmt.Printf("write %s,value is %s,SegDescriptor is %s\n", key, value, cryptoDescriptor)

	var cds []CryptoDescriptor
	if err := json.Unmarshal([]byte(cryptoDescriptor), &cds); err != nil {
		return shim.Error("unmarshal cryptoDescriptor error: " + err.Error())
	}

	var rawDataMap map[string]interface{}
	if err := json.Unmarshal([]byte(value), &rawDataMap); err != nil {
		return shim.Error("unmarshal value error: " + err.Error())
	}

	blockHead := BlockHead{
		CryptoDescriptor: cryptoDescriptor,
		Key:              key,
	}
	var writeTo = make(map[string]interface{}, 128)
	writeTo["head"] = blockHead
	for key, value := range rawDataMap {
		writeTo[key] = value
	}
	bytes, err := json.Marshal(writeTo)
	if err != nil {
		return shim.Error("json marshal error: " + err.Error())
	}
	if err := stub.PutState(key, bytes); err != nil {
		return shim.Error("write fail " + err.Error())
	}

	var ret = make(map[string]interface{}, 4)

	txID := stub.GetTxID()

	ret["transactionId"] = txID

	bytes2, err2 := json.Marshal(ret)
	if err2 != nil {
		return shim.Error("json marshal error: " + err2.Error())
	}
	return shim.Success(bytes2)
}

func parseMultiSegData(stub shim.ChaincodeStubInterface, jsonValue string) (string, error) {
	return jsonValue, nil
}

func (t *Chaincode) queryByParam(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		fmt.Printf("@@queryByParam mett error [%s]\n.", err.Error())
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
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
		var decryptBuffer bytes.Buffer
		decryptBuffer.WriteString("{\"Key\":")
		decryptBuffer.WriteString("\"")
		decryptBuffer.WriteString(queryResponse.Key)
		//fmt.Printf("queryResponse.Key is[%s]\n", queryResponse.Key)
		decryptBuffer.WriteString("\"")

		decryptBuffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		//fmt.Printf("queryResponse.Value is[%s]\n", queryResponse.Value)
		decryptString, err := parseMultiSegData(stub, string(queryResponse.Value))
		if err != nil {
			fmt.Printf("parseMultiSegData meet error [%s]\n", err.Error())
			decryptBuffer.WriteString(string(queryResponse.Value))
		} else {
			decryptBuffer.WriteString(decryptString)
		}
		decryptBuffer.WriteString("}")
		buffer.WriteString(decryptBuffer.String())
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

//BlockHead descr
type BlockHead struct {
	CryptoDescriptor string `json:"cryptoDescriptor"`
	Key              string `json:"key"`
}

//HeadBodyBlock desc
type HeadBodyBlock struct {
	Head BlockHead              `json:"head"`
	Body map[string]interface{} `json:"body"`
}

//Init {"Args":["init"]}
func (t *Chaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Init Chaincode Chaincode")
	return shim.Success(nil)
}

func (t *Chaincode) delByKey(stub shim.ChaincodeStubInterface, key string) pb.Response {
	fmt.Printf("del %s\n", key)
	err := stub.DelState(key)
	if err != nil {
		return shim.Error("query fail " + err.Error())
	}
	return shim.Success(nil)
}

//Invoke {"writeMultiSegData":["key","value","SegDataDescriptor"]}
//
func (t *Chaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "query":
		if len(args) != 1 {
			return shim.Error("parametes's number is wrong")
		}
		return t.query(stub, args[0])
	case "queryByParam":
		if len(args) != 1 {
			return shim.Error("parametes's number is wrong")
		}
		return t.queryByParam(stub, args)
	case "sync":
		if len(args) != 3 {
			return shim.Error("parametes's number is wrong")
		}
		return t.writeMultiSegData(stub, args[0], args[1], args[2])
	default:
		return shim.Error("Invalid invoke function name.")
	}
}

func main() {
	err := shim.Start(new(Chaincode))
	if err != nil {
		fmt.Printf("Error starting Chaincode chaincode: %s", err)
	}
}
