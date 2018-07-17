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
	"github.com/op/go-logging"
)

type Currency struct {
	ERPSystem       string `json:"erpsystem"`
	Client          string `json:"client"`
	CurrencyCode    string `json:"currencycode"`
	AlternativeCode string `json:"alternativecode`
	LegacyCode      string `json:"legacycode"`
}

func AddCurrencyRecords(stub shim.ChaincodeStubInterface, currencyRecArr string) pb.Response {
	var myLogger = logging.MustGetLogger("Procure-To-Pay")
	var currencys []Currency
	err := json.Unmarshal([]byte(currencyRecArr), &currencys)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input tax code array:", err)
	}
	for _, currency := range currencys {
		db.TableStruct{Stub: stub, TableName: util.TAB_CURRENCY, PrimaryKeys: []string{currency.ERPSystem, currency.Client, currency.CurrencyCode, currency.AlternativeCode, currency.LegacyCode}, Data: string(util.MarshalToBytes(currency))}.Add()
	}
	return shim.Success(nil)
}

/*
  Get company data from blockchain
*/
func GetCurrency(stub shim.ChaincodeStubInterface, erpsystem string, client string, currencycode string, alternativecode string, legacycode string) (Currency, string) {
	var myLogger = logging.MustGetLogger("Procure-To-Pay")
	ccRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_CURRENCY, PrimaryKeys: []string{erpsystem, client, currencycode, alternativecode, legacycode}, Data: ""}.Get()
	var currency Currency
	err := json.Unmarshal([]byte(ccRecord), &currency)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input currency:", err, ccRecord)
		return currency, "ERROR in parsing input currency"
	}
	return currency, ""
}

// GetALL method to get all data
func GetAllCurrency(stub shim.ChaincodeStubInterface) []Currency {
	var allCurrencys []Currency
	CurrencysRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_CURRENCY, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, grnRow := range CurrencysRec {
		var currentCurrency Currency
		json.Unmarshal([]byte(grnRow), &currentCurrency)
		allCurrencys = append(allCurrencys, currentCurrency)
	}

	return allCurrencys
}
