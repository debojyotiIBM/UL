/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package taxcode

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	"github.com/op/go-logging"
)

var myLogger = logging.MustGetLogger("Procure-To-Pay")

type TAXCODE struct {
	ERPSystem      string  `json:"erpsystem"`
	TaxCode        string  `json:"taxCode"`
	TaxDescription string  `json:"taxdescription"`
	TaxPercentage  float64 `json:"taxpercentage"`
	Procedure      string  `json:"procedure"`
	CountryCode    string  `json:"countrycode"`
	UPDCode        string  `json:"updcode"`
	Client         string  `json:"client"`
}

// Tax Code

func AddTaxRecords(stub shim.ChaincodeStubInterface, taxRecArr string) pb.Response {
	var taxCodes []TAXCODE
	err := json.Unmarshal([]byte(taxRecArr), &taxCodes)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input tax code array:", err)
	}
	for _, taxCode := range taxCodes {
		db.TableStruct{Stub: stub, TableName: util.TAB_TAXCODE, PrimaryKeys: []string{taxCode.ERPSystem, taxCode.TaxCode, taxCode.Procedure, taxCode.CountryCode, taxCode.Client}, Data: string(util.MarshalToBytes(taxCode))}.Add()
	}
	return shim.Success(nil)
}

/*
  Get company data from blockchain
*/
func GetTaxCode(stub shim.ChaincodeStubInterface, erpsystem string, taxcode string, procedure string, countrycode string, client string) (TAXCODE, string) {
	ccRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_TAXCODE, PrimaryKeys: []string{erpsystem, taxcode, procedure, countrycode, client}, Data: ""}.Get()
	var taxCode TAXCODE
	err := json.Unmarshal([]byte(ccRecord), &taxCode)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input tax code:", err, ccRecord)
		return taxCode, "ERROR in parsing input taxcode"
	}
	return taxCode, ""
}
