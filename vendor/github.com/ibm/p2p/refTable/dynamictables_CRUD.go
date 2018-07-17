/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package refTable

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	//"github.com/op/go-logging"
)

//var myLogger = logging.MustGetLogger("Procure-To-Pay")

type DynamicTables struct {
	CompanyCode   string `json:"companycode"`
	DocType       string `json:"doctype"`
	DocSource     string `json:"docsource"`
	StageName     string `json:"stagename"`
	AttributeName string `json:"attributename"`
	Value         string `json:"value"`
}

func AddDynamicTablesRecords(stub shim.ChaincodeStubInterface, dynamictablesRecArr string) pb.Response {
	var dynamictables []DynamicTables
	err := json.Unmarshal([]byte(dynamictablesRecArr), &dynamictables)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input sapcountry array:", err)
	}

	for _, dynamictable := range dynamictables {
		db.TableStruct{Stub: stub, TableName: util.TAB_DYNAMICTABLES, PrimaryKeys: []string{dynamictable.CompanyCode, dynamictable.DocType, dynamictable.DocSource, dynamictable.StageName, dynamictable.AttributeName}, Data: string(util.MarshalToBytes(dynamictable))}.Add()
	}
	return shim.Success(nil)
}

/*
  Get company data from blockchain
*/
func GetDynamicTables(stub shim.ChaincodeStubInterface, erpsystem string, doctype string, docsource string, stagename string, attributename string) (DynamicTables, string) {

	ccRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_DYNAMICTABLES, PrimaryKeys: []string{erpsystem, doctype, docsource, stagename, attributename}, Data: ""}.Get()
	var dynamictable DynamicTables

	err := json.Unmarshal([]byte(ccRecord), &dynamictable)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input sapcountry:", err, ccRecord)
		return dynamictable, "ERROR in parsing input sapcountry"
	}
	return dynamictable, ""
}

// GetALL method to get all data
func GetAllDynamicTables(stub shim.ChaincodeStubInterface) []DynamicTables {
	var allDynamicTabless []DynamicTables
	DynamicTablessRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_DYNAMICTABLES, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, grnRow := range DynamicTablessRec {
		var currentDynamicTables DynamicTables
		json.Unmarshal([]byte(grnRow), &currentDynamicTables)
		allDynamicTabless = append(allDynamicTabless, currentDynamicTables)
	}

	return allDynamicTabless
}
