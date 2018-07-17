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

type CurrencyDescription struct {
	ERPSystem       string `json:"erpsystem"`
	Client          string `json:"client"`
	LanguageKey     string `json:"languagekey"`
	CurrencyCode    string `json:"currencycode"`
	Description     string `json:"description"`
	AlternativeCode string `json:"alternativecode`
}

func AddCurrencyDescRecords(stub shim.ChaincodeStubInterface, currencyDescriptionRecArr string) pb.Response {
	var currencyDescriptions []CurrencyDescription
	err := json.Unmarshal([]byte(currencyDescriptionRecArr), &currencyDescriptions)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input tax code array:", err)
	}
	for _, currencyDescription := range currencyDescriptions {
		db.TableStruct{Stub: stub, TableName: util.TAB_CURRENCYDESCRIPTION, PrimaryKeys: []string{currencyDescription.ERPSystem, currencyDescription.Client, currencyDescription.LanguageKey, currencyDescription.CurrencyCode, currencyDescription.AlternativeCode}, Data: string(util.MarshalToBytes(currencyDescription))}.Add()
	}
	return shim.Success(nil)
}

/*
  Get company data from blockchain
*/
func GetCurrencyDescription(stub shim.ChaincodeStubInterface, erpsystem string, client string, languagekey string, currencycode string, alternativecode string) (CurrencyDescription, string) {
	ccRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_CURRENCYDESCRIPTION, PrimaryKeys: []string{erpsystem, client, languagekey, currencycode, alternativecode}, Data: ""}.Get()
	var currencyDescription CurrencyDescription
	err := json.Unmarshal([]byte(ccRecord), &currencyDescription)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input currencyDescription:", err, ccRecord)
		return currencyDescription, "ERROR in parsing input currencyDescription"
	}
	return currencyDescription, ""
}

// GetALL method to get all data
func GetAllCurrencyDescription(stub shim.ChaincodeStubInterface) []CurrencyDescription {
	var allCurrencyDescriptions []CurrencyDescription
	CurrencyDescriptionsRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_CURRENCYDESCRIPTION, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, grnRow := range CurrencyDescriptionsRec {
		var currentCurrencyDescription CurrencyDescription
		json.Unmarshal([]byte(grnRow), &currentCurrencyDescription)
		allCurrencyDescriptions = append(allCurrencyDescriptions, currentCurrencyDescription)
	}

	return allCurrencyDescriptions
}
