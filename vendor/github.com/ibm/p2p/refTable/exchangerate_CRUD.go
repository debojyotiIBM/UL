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

type ExchangeRate struct {
	ERPSystem          string      `json:"erpsystem"`
	Client             string      `json:"client"`
	ExchangeRateType   string      `json:"exchangeratetype"`
	FromCurr           string      `json:"fromcurr"`
	ToCurr             string      `json:"tocurr"`
	EffectiveDate      util.BCDate `json:"effectivedate"`
	ExchangeRate       float64     `json:"exchangerate"`
	RatioFromCurrUnits float64     `json:"ratiofromcurrunits"`
	RatioToCurrUnits   float64     `json:"ratiotocurrunits"`
}

// Exchange Rate ADD Function

func AddExchangeRecords(stub shim.ChaincodeStubInterface, exchangeRecArr string) pb.Response {
	var exchangeRates []ExchangeRate
	err := json.Unmarshal([]byte(exchangeRecArr), &exchangeRates)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input exchangerate array:", err)
	}
	for _, exchangeRate := range exchangeRates {
		db.TableStruct{Stub: stub, TableName: util.TAB_EXCHANGERATE, PrimaryKeys: []string{exchangeRate.ERPSystem, exchangeRate.Client, exchangeRate.ExchangeRateType, exchangeRate.FromCurr, exchangeRate.ToCurr, (exchangeRate.EffectiveDate).String()}, Data: string(util.MarshalToBytes(exchangeRate))}.Add()
	}
	return shim.Success(nil)
}

/*
  Get company data from blockchain
*/
func GetExchangeRate(stub shim.ChaincodeStubInterface, erpsystem string, client string, exchangeratetype string, fromcurr string, tocurr string, effectivedate string) (ExchangeRate, string) {
	ccRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_EXCHANGERATE, PrimaryKeys: []string{erpsystem, client, exchangeratetype, fromcurr, tocurr, effectivedate}, Data: ""}.Get()
	var exchangeRate ExchangeRate
	err := json.Unmarshal([]byte(ccRecord), &exchangeRate)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input exchange rate:", err, ccRecord)
		return exchangeRate, "ERROR in parsing input exchage rate"
	}
	return exchangeRate, ""
}

// GetALL method to get all data
func GetAllExchangeRate(stub shim.ChaincodeStubInterface) []ExchangeRate {
	var allExchangeRates []ExchangeRate
	ExchangeRatesRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_EXCHANGERATE, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, grnRow := range ExchangeRatesRec {
		var currentExchangeRate ExchangeRate
		json.Unmarshal([]byte(grnRow), &currentExchangeRate)
		allExchangeRates = append(allExchangeRates, currentExchangeRate)
	}

	return allExchangeRates
}
