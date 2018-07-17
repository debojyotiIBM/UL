/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package companyCode

import (
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	logging "github.com/op/go-logging"
)

type CompanyCode struct {
	ErpSystem                  string `json:"erpsystem"`
	CompanyCode                string `json:"companycode"`
	Name                       string `json:"companyname"`
	PriceAbsoluteUpperThresold int64  `json:"priceabsoluteupperthresold"`
	PriceRelativeUpperThresold int64  `json:"pricerelativeupperthresold"`
	PriceAbsoluteLowerThresold int64  `json:"priceabsolutelowerthresold"`
	PriceRelativeLowerThresold int64  `json:"pricerelativelowerthresold"`
	ACAbsoluteThreshold        int64  `json:"acabsolutethreshold"`
	ACRelativeThreshold        int64  `json:"acrelativethreshold"`
}

type CompanyCodeDays struct {
	CompanyCode string `json:"companyCode"`
	CompanyName string `json:"companyName"`
	Days        string `json:"Days"`
}

type CompanyCodeController struct {
	CompanyCode string `json:"companyCode"`
	CompanyName string `json:"companyName"`
	Controller  string `json:"Controller"`
	EmailId     string `json:"emailId"`
}

var myLogger = logging.MustGetLogger("Procure-To-Pay CompanyCode")

/*
  Adds company data to blockchain
*/

func AddCompanyRecords(stub shim.ChaincodeStubInterface, companyRecArr string) pb.Response {
	var companyCodes []CompanyCode
	err := json.Unmarshal([]byte(companyRecArr), &companyCodes)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input company code array:", err)
	}
	for _, companyCode := range companyCodes {
		db.TableStruct{Stub: stub, TableName: util.TAB_COMPANYCODE, PrimaryKeys: []string{companyCode.ErpSystem, companyCode.CompanyCode}, Data: string(util.MarshalToBytes(companyCode))}.Add()
	}
	return shim.Success(nil)
}

/*
  Get company data from blockchain
*/
func GetCompanyCode(stub shim.ChaincodeStubInterface, erpsystem string, cc string) (CompanyCode, string) {
	ccRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_COMPANYCODE, PrimaryKeys: []string{erpsystem, cc}, Data: ""}.Get()
	var companyCode CompanyCode
	err := json.Unmarshal([]byte(ccRecord), &companyCode)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input company code:", err, ccRecord)
		return companyCode, "ERROR in parsing input company code"
	}
	return companyCode, ""
}
func AddCompanyCodeDays(stub shim.ChaincodeStubInterface, companyDaysRecArr string) pb.Response {

	var companyCodeDaysArr []CompanyCodeDays
	err := json.Unmarshal([]byte(companyDaysRecArr), &companyCodeDaysArr)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input company code array:", err)
	}
	for _, companyCodeObj := range companyCodeDaysArr {
		db.TableStruct{Stub: stub, TableName: util.TAB_COMPANYCODE_DAYS, PrimaryKeys: []string{companyCodeObj.CompanyCode}, Data: string(util.MarshalToBytes(companyCodeObj))}.Add()
	}
	return shim.Success(nil)

}
func GetCompanyCodeDays(stub shim.ChaincodeStubInterface, code string) (CompanyCodeDays, string) {
	ccDaysRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_COMPANYCODE_DAYS, PrimaryKeys: []string{code}, Data: ""}.Get()
	var companyCodeDays CompanyCodeDays
	err := json.Unmarshal([]byte(ccDaysRecord), &companyCodeDays)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input company code Days:", err, ccDaysRecord)
		return companyCodeDays, "ERROR in parsing input company code Days"
	}
	return companyCodeDays, ""
}

func GetAllCompanyCodeDays(stub shim.ChaincodeStubInterface) pb.Response {

	ccDaysRecords, _ := db.TableStruct{Stub: stub, TableName: util.TAB_COMPANYCODE_DAYS, PrimaryKeys: []string{}, Data: ""}.GetAll()
	var compDaysArr []CompanyCodeDays
	var compDays CompanyCodeDays

	for _, ccdt := range ccDaysRecords {
		err := json.Unmarshal([]byte(ccdt), &compDays)
		if err != nil {
			myLogger.Debugf("ERROR parsing events", ccdt, err)
			return shim.Error("ERROR parsing events")
		}
		compDaysArr = append(compDaysArr, compDays)
	}
	return shim.Success(util.MarshalToBytes(compDaysArr))
}

func AddControllers(stub shim.ChaincodeStubInterface, controllerRecArr string) pb.Response {

	var companyCodeContrArr []CompanyCodeController
	err := json.Unmarshal([]byte(controllerRecArr), &companyCodeContrArr)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input company code array:", err)
	}
	for _, companyCodeContrObj := range companyCodeContrArr {
		db.TableStruct{Stub: stub, TableName: util.TAB_COMPANYCODE_CONTROLLER, PrimaryKeys: []string{companyCodeContrObj.CompanyCode}, Data: string(util.MarshalToBytes(companyCodeContrObj))}.Add()
	}
	return shim.Success(nil)
}

func GetAllControllers(stub shim.ChaincodeStubInterface) pb.Response {

	ccContrRecords, _ := db.TableStruct{Stub: stub, TableName: util.TAB_COMPANYCODE_CONTROLLER, PrimaryKeys: []string{}, Data: ""}.GetAll()
	var compContArr []CompanyCodeController
	var compController CompanyCodeController

	for _, ccdt := range ccContrRecords {
		err := json.Unmarshal([]byte(ccdt), &compController)
		if err != nil {
			myLogger.Debugf("ERROR parsing events", ccdt, err)
			return shim.Error("ERROR parsing events")
		}
		compContArr = append(compContArr, compController)
	}
	return shim.Success(util.MarshalToBytes(compContArr))
}
