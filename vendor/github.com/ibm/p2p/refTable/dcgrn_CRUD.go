/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package refTable

import (
	"encoding/json"
	//"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	//"github.com/op/go-logging"
	"strconv"
)

//var myLogger = logging.MustGetLogger("Procure-To-Pay")

type DCGRN struct {
	ERPSystem        string `json:"erpsystem"`
	Client           string `json:"client"`
	PONumber         string `json:"ponumber"`
	POItem           int64  `json:"poitem"`
	Sequence         int64  `json:"sequence"`
	TransType        string `json:"transtype"`
	MatDocYear       int64  `json:"matdocyear"`
	DCGRNumber       string `json:"dcgrnumber"`
	MatDocItem       int64  `json:"matdocitem"`
	StepNumber       string `json:"stepnumber"`
	ConditionCounter string `json:"conditioncounter"`
	Quantity         string `json:"quantity"`
	RefDocNumber     string `json:"refdocnumber"`
	Currency         string `json:"currency"`
	GRNValue         string `json:"grnvalue"`
}

func AddDCGRNRecords(stub shim.ChaincodeStubInterface, dcgrnRecArr string) pb.Response {
	var dcgrns []DCGRN
	err := json.Unmarshal([]byte(dcgrnRecArr), &dcgrns)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input dcgrn array:", err)
	}

	for _, dcgrn := range dcgrns {
		db.TableStruct{Stub: stub, TableName: util.TAB_DCGRN, PrimaryKeys: []string{dcgrn.ERPSystem, dcgrn.Client, dcgrn.PONumber, strconv.Itoa(int(dcgrn.POItem)), strconv.Itoa(int(dcgrn.Sequence)), dcgrn.TransType, strconv.Itoa(int(dcgrn.MatDocYear)), dcgrn.DCGRNumber, strconv.Itoa(int(dcgrn.MatDocItem)), dcgrn.StepNumber, dcgrn.ConditionCounter}, Data: string(util.MarshalToBytes(dcgrn))}.Add()
	}
	return shim.Success(nil)
}

/*
  Get company data from blockchain
*/
func GetDCGRN(stub shim.ChaincodeStubInterface, erpsystem string, client string, ponumber string, poitem string, sequence string, transtype string, matdocyear string, dcgrnumber string, matdocitem string, stepnumber string, conditioncounter string) (DCGRN, string) {

	ccRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_DCGRN, PrimaryKeys: []string{erpsystem, client, ponumber, poitem, sequence, transtype, matdocyear, dcgrnumber, matdocitem, stepnumber, conditioncounter}, Data: ""}.Get()

	var dcgrn DCGRN

	err := json.Unmarshal([]byte(ccRecord), &dcgrn)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input dcgrn:", err, ccRecord)
		return dcgrn, "ERROR in parsing input dcgrn"
	}
	return dcgrn, ""
}

func GetAllDCGRN(stub shim.ChaincodeStubInterface) []DCGRN {
	var allDCGRNs []DCGRN
	dcGRNsRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_DCGRN, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, grnRow := range dcGRNsRec {
		var currentDCGRN DCGRN
		json.Unmarshal([]byte(grnRow), &currentDCGRN)
		allDCGRNs = append(allDCGRNs, currentDCGRN)
	}

	return allDCGRNs
}
