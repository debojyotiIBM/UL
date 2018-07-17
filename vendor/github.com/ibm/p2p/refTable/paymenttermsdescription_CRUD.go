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

type PaymentTermsDescription struct {
	ERPSystem       string `json:"erpsystem"`
	Client          string `json:"client"`
	PaymentTermCode string `json:"paymenttermcode"`
	DayLimit        int64  `json:"daylimit"`
	LanguageKey     string `json:"languagekey"`
	Description     string `json:"description"`
}

func AddPaymentTermsDescRecords(stub shim.ChaincodeStubInterface, payRecArr string) pb.Response {
	var paymentTermsDescriptions []PaymentTermsDescription
	err := json.Unmarshal([]byte(payRecArr), &paymentTermsDescriptions)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input paymentTermsDescriptions array:", err)
	}
	for _, paymentTermsDescription := range paymentTermsDescriptions {
		db.TableStruct{Stub: stub, TableName: util.TAB_PAYMENTTERMSDESCRIPTION, PrimaryKeys: []string{paymentTermsDescription.ERPSystem, paymentTermsDescription.Client, paymentTermsDescription.PaymentTermCode, strconv.Itoa(int(paymentTermsDescription.DayLimit))}, Data: string(util.MarshalToBytes(paymentTermsDescription))}.Add()
	}
	return shim.Success(nil)
}

/*
  Get company data from blockchain
*/
func GetPaymentTermsDescriptions(stub shim.ChaincodeStubInterface, erpsystem string, client string, paymenttermcode string, daylimit string) (PaymentTermsDescription, string) {
	ccRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PAYMENTTERMSDESCRIPTION, PrimaryKeys: []string{erpsystem, client, paymenttermcode, daylimit}, Data: ""}.Get()
	var paymentTermsDescriptions PaymentTermsDescription
	err := json.Unmarshal([]byte(ccRecord), &paymentTermsDescriptions)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input tax code:", err, ccRecord)
		return paymentTermsDescriptions, "ERROR in parsing input paymentTermsDescriptions"
	}
	return paymentTermsDescriptions, ""
}

// GetALL method to get all data
func GetAllPaymentTermsDescription(stub shim.ChaincodeStubInterface) []PaymentTermsDescription {
	var allPaymentTermsDescriptions []PaymentTermsDescription
	PaymentTermsDescriptionsRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PAYMENTTERMSDESCRIPTION, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, grnRow := range PaymentTermsDescriptionsRec {
		var currentPaymentTermsDescription PaymentTermsDescription
		json.Unmarshal([]byte(grnRow), &currentPaymentTermsDescription)
		allPaymentTermsDescriptions = append(allPaymentTermsDescriptions, currentPaymentTermsDescription)
	}

	return allPaymentTermsDescriptions
}
