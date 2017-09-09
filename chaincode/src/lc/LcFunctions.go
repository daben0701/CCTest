package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"

	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"
)

type SimpleChaincode struct {
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	stub.PutState("LCSequence", []byte(strconv.Itoa(1)))
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "issueLc" { //create a new Lc
		return t.issueLc(stub, args)
	} else if function == "getLcByNo" { //read a Lc
		return t.getLcByNo(stub, args)
	} else if function == "transferLc" { //change owner of a specific Lc
		return t.transferLc(stub, args)
	} else if function == "getLcByOwner" { //find Lcs for owner X using rich query
		return t.getLcListByOwner(stub, args)
	}
	//} else if function == "transferMarblesBasedOnColor" { //transfer all Lcs of a certain color
	//	return t.transferMarblesBasedOnColor(stub, args)
	//} else if function == "delete" { //delete a Lc
	//	return t.delete(stub, args)
	//} else if function == "getLcByNo" { //read a Lc
	//	return t.getLcByNo(stub, args)
	//} else if function == "queryMarblesByOwner" { //find Lcs for owner X using rich query
	//	return t.queryMarblesByOwner(stub, args)
	//} else if function == "queryMarbles" { //find Lcs based on an ad hoc rich query
	//	return t.queryMarbles(stub, args)
	//} else if function == "getHistoryForMarble" { //get history of values for a Lc
	//	return t.getHistoryForMarble(stub, args)
	//} else if function == "getMarblesByRange" { //get Lcs based on range query
	//	return t.getMarblesByRange(stub, args)
	//}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

//发信用证给指定客户
func (t *SimpleChaincode) issueLc(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//   0       1
	// "apply corp code", "amount"
	if len(args) < 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}

	lcNumStr := getLcNumber(stub)
	sendBank, err := decodeBank(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	recvBank, err := decodeBank(args[1])
	if err != nil {
		return shim.Error(err.Error())
	}
	applyCorp, err := decodeCorp(args[2])
	if err != nil {
		return shim.Error(err.Error())
	}
	benefCorp, err := decodeCorp(args[3])
	if err != nil {
		return shim.Error(err.Error())
	}
	amount, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		return shim.Error("2nd argument must be a numeric string")
	}
	issueDate, err := time.Parse("20060102", args[5])
	if err != nil {
		return shim.Error("6nd argument must be a date string yyyyMMdd")
	}
	expireDate, err := time.Parse("20060102", args[6])
	if err != nil {
		return shim.Error("6nd argument must be a date string yyyyMMdd")
	}
	lc := &LCLetter{lcNumStr, sendBank, recvBank, applyCorp, benefCorp, amount,
		issueDate, expireDate, recvBank.LegalEntity} //开证的时候直接给了接收行

	creatorByte, err := stub.GetCreator()
	if err != nil {
		return shim.Error("Error stub.GetCreator")
	}
	//fmt.Println(creatorByte)
	fmt.Println(string(creatorByte))
	certStart := bytes.IndexAny(creatorByte, "-----") // Devin:I don't know why sometimes -----BEGIN is invalid, so I use -----
	if certStart == -1 {
		return shim.Error("No certificate found")
	}
	certText := creatorByte[certStart:]
	fmt.Println("certStart:" + strconv.Itoa(certStart))
	//fmt.Println(certText)
	bl, _ := pem.Decode(certText)
	if bl == nil {
		return shim.Error("Could not decode the PEM structure")
	}

	cert, err := x509.ParseCertificate(bl.Bytes)
	if err != nil {
		return shim.Error("ParseCertificate failed")
	}
	fmt.Println(cert)

	//fmt.Println("Orgs:"+strings.Join( cert.Subject.Organization,","))
	//fmt.Println("DNS:"+ strings.Join( cert.DNSNames,","))
	fmt.Println("Issuer:" + strings.Join(cert.Issuer.Organization, ","))
	lcJSONasBytes, err := json.Marshal(lc)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println(lcJSONasBytes)
	// === Save LC to state ===
	err = stub.PutState(lcNumStr, lcJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(lcNumStr))
}

//获得LC的编号，并且将序列+1
func getLcNumber(stub shim.ChaincodeStubInterface) string {
	return "LC" + time.Now().Format("20060102") + strconv.Itoa(getNextSequence(stub, "LC"))
}

func getNextSequence(stub shim.ChaincodeStubInterface, formPrefix string) int {
	key := formPrefix + "Sequence"
	lcSeqAsBytes, err := stub.GetState(key)
	if err != nil {
		shim.Error("Failed to get Sequence: " + err.Error())
	}
	seq, _ := strconv.Atoi(string(lcSeqAsBytes))
	stub.PutState(key, []byte(strconv.Itoa(seq+1)))
	return seq
}

func decodeCorp(jsonStr string) (Corporation, error) {
	var corp Corporation
	err := json.Unmarshal([]byte(jsonStr), &corp)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to decode JSON of: " + jsonStr + "\" to Corporation}"
		return Corporation{}, errors.New(jsonResp)
	}
	return corp, nil
}
func decodeBank(jsonStr string) (Bank, error) {
	var bank Bank
	err := json.Unmarshal([]byte(jsonStr), &bank)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to decode JSON of: " + jsonStr + "\" to Bank}"
		return Bank{}, errors.New(jsonResp)
	}
	return bank, nil
}
func decodeLegalEntity(jsonStr string) (LegalEntity, error) {
	var bank LegalEntity
	err := json.Unmarshal([]byte(jsonStr), &bank)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to decode JSON of: " + jsonStr + "\" to LegalEntity}"
		return LegalEntity{}, errors.New(jsonResp)
	}
	return bank, nil
}

//根据信用证号查询信用证信息
func (t *SimpleChaincode) getLcByNo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	lcNum := args[0]
	lcBytes, err := stub.GetState(lcNum)
	if err != nil {
		return shim.Error("query Letter of Credit fail. Number:" + lcNum)
	}
	return shim.Success(lcBytes)
}

//查看某Entity下有哪些信用证
func (t *SimpleChaincode) getLcListByOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	entityNo := args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"Owner.No\":\"%s\"}}", entityNo)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
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

//将信用证递交给下一方
func (t *SimpleChaincode) transferLc(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2. lcNum,newOwnerEntityJson")
	}
	lcNum := args[0]
	lcBytes, err := stub.GetState(lcNum)
	if err != nil {
		return shim.Error("query Letter of Credit fail. Number:" + lcNum)
	}
	entity, err := decodeLegalEntity(args[1])
	if err != nil {
		return shim.Error(err.Error())
	}
	lc := LCLetter{}
	err = json.Unmarshal(lcBytes, &lc) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("Transfer LC " + lcNum + " from owner:" + lc.Owner.No + " to new owner:" + entity.No)
	lc.Owner = entity
	jsonB, _ := json.Marshal(lc)
	err = stub.PutState(lcNum, jsonB) //rewrite the marble
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end transfer LC (success)")
	return shim.Success(nil)
}

//获得Bill of Landing的编号，并且将序列+1
func getBlNumber(stub shim.ChaincodeStubInterface) string {
	return "BL" + time.Now().Format("20060102") + strconv.Itoa(getNextSequence(stub, "BL"))
}
