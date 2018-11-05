package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/thomasks/eju-chaincodes/cryptoutils"
)

// ToChaincodergs comment
func ToChaincodergs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

// Chaincode comment
type Chaincode struct {
}

//{"Args":["attr", "name"]}'
func (t *Chaincode) attr(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("get attr: ", args[0])
	value, ok, err := cid.GetAttributeValue(stub, args[0])
	if err != nil {
		return shim.Error("get attr error: " + err.Error())
	}

	if ok == false {
		value = "not found"
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return shim.Error("json marshal error: " + err.Error())
	}
	return shim.Success(bytes)
}

//{"Args":["query_chaincode","chaincode","key"]}'
func (t *Chaincode) queryChaincode(stub shim.ChaincodeStubInterface, chaincode, key string) pb.Response {
	fmt.Printf("query %s in %s\n", key, chaincode)
	return stub.InvokeChaincode(chaincode, ToChaincodergs("query", key), stub.GetChannelID())
}

//{"Args":["write_chaincode","chaincode","key","value"]}'
func (t *Chaincode) writeChaincode(stub shim.ChaincodeStubInterface, chaincode, key, value string) pb.Response {
	fmt.Printf("write %s to %s, value is %s\n", key, chaincode, value)
	return stub.InvokeChaincode(chaincode, ToChaincodergs("write", key, value), stub.GetChannelID())
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

//{"Args":["history","key"]}'
func (t *Chaincode) history(stub shim.ChaincodeStubInterface, key string) pb.Response {
	fmt.Printf("history %s\n", key)
	iter, err := stub.GetHistoryForKey(key)
	defer iter.Close()
	if err != nil {
		return shim.Error("query fail " + err.Error())
	}

	values := make(map[string]string)

	for iter.HasNext() {
		fmt.Printf("next\n")
		if kv, err := iter.Next(); err == nil {
			fmt.Printf("id: %s value: %s\n", kv.TxId, kv.Value)
			values[kv.TxId] = string(kv.Value)
		}
		if err != nil {
			return shim.Error("iterator history fail: " + err.Error())
		}
	}

	bytes, err := json.Marshal(values)
	if err != nil {
		return shim.Error("json marshal fail: " + err.Error())
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

	var cds []cryptoutils.CryptoDescriptor
	if err := json.Unmarshal([]byte(cryptoDescriptor), &cds); err != nil {
		return shim.Error("unmarshal cryptoDescriptor error: " + err.Error())
	}

	var rawDataMap map[string]interface{}
	if err := json.Unmarshal([]byte(value), &rawDataMap); err != nil {
		return shim.Error("unmarshal value error: " + err.Error())
	}

	if len(cds) >= 1 {
		cryptoutils.CryptoDataByDescriptor(stub, rawDataMap, cds)
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

//{"Args":["readMultiSegData","key","value",SegDescriptor]}
func (t *Chaincode) readMultiSegData(stub shim.ChaincodeStubInterface, key string) pb.Response {
	fmt.Printf("key %s\n", key)
	bytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("query fail " + err.Error())
	}

	/*var readTo = new(HeadBodyBlock)
	if err := json.Unmarshal(bytes, readTo); err != nil {
		return shim.Error("unmarshal rawBytes error: " + err.Error())
	}
	var cds []cryptoutils.CryptoDescriptor
	if err := json.Unmarshal([]byte(readTo.Head.CryptoDescriptor), &cds); err != nil {
		return shim.Error("unmarshal cryptoDescriptor error: " + err.Error())
	}*/

	//crypto value in each level
	/*for _, rawDataMap := range readTo.Body {
		cryptoutils.DecryptoDataByDescriptor(stub, rawDataMap, cds)
		//attris := cd.CryptoFields
		//level := cd.Level
		//getCryptoKey4ChannelLevel(level, stub.GetChannelID)
	}
	ret, err2 := json.Marshal(readTo.Body)
	if err2 != nil {
		return shim.Error("json marshal error: " + err.Error())
	}*/
	return shim.Success(bytes)
}

func parseMultiSegData(stub shim.ChaincodeStubInterface, jsonValue string) (string, error) {
	fmt.Printf("@@parseMultiSegData jsonValue is: [%s]\n", jsonValue)
	var bytes = []byte(jsonValue)
	var readTo = make(map[string]interface{}, 128)
	if err := json.Unmarshal(bytes, &readTo); err != nil {
		fmt.Printf("@@parseMultiSegData readTo mett error [%s]\n.", err.Error())
		return jsonValue, err
	}
	var headMap, ok = readTo["head"].(map[string]interface{})
	if !ok {
		fmt.Println("head is not a map!")
	}

	var cdsJSON, ok2 = headMap["cryptoDescriptor"].(string)
	if !ok2 {
		fmt.Printf("cryptoDescriptor is not a string\n,cryptoDescriptor type is %T\ncryptoDescriptor value is %v\n", headMap["cryptoDescriptor"], headMap["cryptoDescriptor"])
	}
	var cds []cryptoutils.CryptoDescriptor
	if err := json.Unmarshal([]byte(cdsJSON), &cds); err != nil {
		fmt.Printf("@@parseMultiSegData CryptoDescriptor mett error [%s]\nraw json string is [%s]\n", err.Error(), cdsJSON)
		return jsonValue, err
	}
	cryptoutils.DecryptoDataByDescriptor(stub, readTo, cds)
	ret, err2 := json.Marshal(readTo)
	if err2 != nil {
		fmt.Printf("@@parseMultiSegData Marshal mett error [%s]\n.", err2.Error())
		return jsonValue, err2
	}
	return string(ret), nil
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
		// 获取每个需要解密的数据
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
		if err != nil { // 如果解密失败，则返回加密数据
			fmt.Printf("parseMultiSegData meet error [%s]\n", err.Error())
			decryptBuffer.WriteString(string(queryResponse.Value))
		} else { // 解密成功，则返回解密后数据
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
	fmt.Printf("query %s\n", key)
	err := stub.DelState(key)
	if err != nil {
		return shim.Error("query fail " + err.Error())
	}
	return shim.Success(nil)
}

//Invoke {"write":["key","value"]}
//Invoke {"writeMultiSegData":["key","value","SegDataDescriptor"]}
//
func (t *Chaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "attr":
		if len(args) != 1 {
			return shim.Error("parametes's number is wrong")
		}
		return t.attr(stub, args)
	case "query": //查询
		if len(args) != 1 {
			return shim.Error("parametes's number is wrong")
		}
		return t.query(stub, args[0])
	case "queryByParam": //查询
		if len(args) != 1 {
			return shim.Error("parametes's number is wrong")
		}
		return t.queryByParam(stub, args)
	case "sync": //写入
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
