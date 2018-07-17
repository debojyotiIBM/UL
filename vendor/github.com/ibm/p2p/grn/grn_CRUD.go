/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package grn

import (
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	//"github.com/jinzhu/copier"
	logging "github.com/op/go-logging"
)

type GRN struct {
	ErpSystem    string `json:"erpsystem"`
	Client       string `json:"client"`
	GrnNumber    string `json:"grnnumber"`
	PONumber     string `json:"ponumber"`
	POItemNumber int64  `json:poitemnumber`
	Sequence     int64  `json:"sequence"`
	TransType    string `json:"transtype"`
	MatDocNumber string `json:"matdocnumber"`
	MatDocYear   int64  `json:"matdocyear"`
	GRNLineItem  int64  `json:grnlineitem`

	Quantity     float64 `json:"quantity"`
	RefDocNumber string  `json:"refdocnumber"`
	Currency     string  `json:"currency"`
	GRNValue     float64 `json:grnvalue`

	DocPostDate       util.BCDate `json:"docpostdate"`
	DeliveryCompleted string      `json:"deliverycompleted"`
	LocalCurrency     string      `json:"localcurrency"`
	CreationUser      string      `json:"creationuser"`

	ResidualQuantity float64 `json:"residualquantity"`
}

type InvLine struct {
	InvNumber     string `json:"invnumber"`
	InvLineNumber int64  `json:"invlinenumber"`
}

var myLogger = logging.MustGetLogger("Procure-To-Pay : GRN")

/*
	Add GRN records to blockchain
*/

/*func GRNLineMatch(line1 GRNLineItem, line2 GRNLineItem) bool {
	return (line1.PoNumber == line2.PoNumber &&
		line1.PoLineItemNumber == line2.PoLineItemNumber)
}*/
func AddGrnRecords(stub shim.ChaincodeStubInterface, grnRecArr string) pb.Response {
	/*var grnsByPo map[string]string
	grnsByPo = make(map[string]string)*/

	var grns []GRN
	err := json.Unmarshal([]byte(grnRecArr), &grns)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input GRN array:", err, grns)
	}
	for _, grn := range grns {

		//grnUpdated := GroupGrnLineItems(grn)
		// var lineItems []GRNLineItem
		// for _, line := range *grnUpdated.LineItems() {
		// 	lineItems = append(lineItems, line)
		// }
		// grnUpdated.SetLineItems(lineItems)
		poitemNum := util.GetStringFromInt(grn.POItemNumber)
		seq := util.GetStringFromInt(grn.Sequence)
		matdocyear := util.GetStringFromInt(grn.MatDocYear)
		grnlineitem := util.GetStringFromInt(grn.GRNLineItem)
		AddGRN(stub, []string{grn.ErpSystem, grn.PONumber, poitemNum, grn.GrnNumber, grnlineitem, seq, grn.TransType, matdocyear, grn.Client}, string(util.MarshalToBytes(grn)))

		// collect GRNS by PO
		/*	if len(*grn.LineItems) > 0 {
			for _, line := range *grn.LineItems {
				if grnsByPo[line.PoNumber] != "" {
					grnsByPo[line.PoNumber] = grnsByPo[line.PoNumber] + "|" + grn.ErpSystem + "~" + grn.GrnNumber
				} else {
					grnsByPo[line.PoNumber] = grn.ErpSystem+ "~" + grn.GrnNumber
				}
			}
		}*/
	}

	/*for poNumber, val := range grnsByPo {
		util.UpdateReferenceData(stub, util.TAB_GRN_BY_PO, []string{poNumber}, val)
	}*/
	return shim.Success(nil)
}

func AddGRN(stub shim.ChaincodeStubInterface, keys []string, grnStr string) {
	db.TableStruct{Stub: stub, TableName: util.TAB_GRN, PrimaryKeys: keys, Data: grnStr}.Add()
}

func GetAllGRNs(stub shim.ChaincodeStubInterface) pb.Response {
	var grns []GRN
	recordsMap, _ := db.TableStruct{Stub: stub, TableName: util.TAB_GRN, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, rec := range recordsMap {
		var grn GRN
		err := json.Unmarshal([]byte(rec), &grn)
		if err != nil {
			myLogger.Debugf("ERROR parsing grns  :", err)
			return shim.Error("ERROR parsing Vendors")
		}
		grns = append(grns, grn)
	}
	return shim.Success(util.MarshalToBytes(grns))
}

func GetGRN(stub shim.ChaincodeStubInterface, keys []string) (GRN, string) {
	record, _ := db.TableStruct{Stub: stub, TableName: util.TAB_GRN, PrimaryKeys: keys, Data: ""}.Get()
	var grn GRN
	err := json.Unmarshal([]byte(record), &grn)
	if err != nil {
		myLogger.Debugf("ERROR parsing grn  :", err)
		return grn, "ERROR parsing grn"
	}
	return grn, ""
}

func GetGrnsByPO(stub shim.ChaincodeStubInterface, keys []string) []GRN {
	var grnsByPO []GRN
	grnRecords, grnFetchErr := db.TableStruct{Stub: stub, TableName: util.TAB_GRN, PrimaryKeys: keys, Data: ""}.GetAll()
	if grnFetchErr != nil {
		myLogger.Debugf("ERROR in fetching grn with PO :", grnFetchErr)
	}

	for _, grnRow := range grnRecords {
		var currentGRN GRN
		json.Unmarshal([]byte(grnRow), &currentGRN)
		grnsByPO = append(grnsByPO, currentGRN)
	}

	return grnsByPO
}

func GetGrnsByPOAndPOLine(stub shim.ChaincodeStubInterface, keys []string) []GRN {
	var grnsByPOAndPOLine []GRN
	grnRecords, grnFetchErr := db.TableStruct{Stub: stub, TableName: util.TAB_GRN, PrimaryKeys: keys, Data: ""}.GetAll()
	if grnFetchErr != nil {
		myLogger.Debugf("ERROR in fetching grn with PO and POLine :", grnFetchErr)
	}

	for _, grnRow := range grnRecords {
		var currentGRN GRN
		json.Unmarshal([]byte(grnRow), &currentGRN)
		grnsByPOAndPOLine = append(grnsByPOAndPOLine, currentGRN)
	}

	return grnsByPOAndPOLine
}

func FilterGRNs(stub shim.ChaincodeStubInterface, keys []string) []GRN {
	var selectedGrns []GRN
	recordsMap, _ := db.TableStruct{Stub: stub, TableName: util.TAB_GRN, PrimaryKeys: keys, Data: ""}.GetAll()
	for _, rec := range recordsMap {
		var grn GRN
		err := json.Unmarshal([]byte(rec), &grn)
		if err != nil {
			myLogger.Debugf("ERROR parsing grns  :", err)
			//return shim.Error("ERROR parsing Vendors")
		}
		selectedGrns = append(selectedGrns, grn)
	}
	return selectedGrns
}

/*func GetGrnsByPoAndLineItemNumber(stub shim.ChaincodeStubInterface, poNumber string, lineNumber int64) ([]GRN, string) {
	var selectedGrns []GRN
	grns := GetGrnsByPO(stub, poNumber)
	for _, grn := range grns {
		for _, line := range *grn.LineItems() {
			if line.PoNumber() == poNumber && line.PoLineItemNumber() == lineNumber {
				grn.SetLineItems([]GRNLineItem{line})
				selectedGrns = append(selectedGrns, grn)
			}
		}
	}
	return selectedGrns, ""
}*/

/*func GroupGrnLineItems(grn GRN) GRN {
	var selectedLineItems []GRNLineItem
	var uniquePoAndLineNumbers []string

	for _, line := range *grn.LineItems {
		if !(util.StringArrayContains(uniquePoAndLineNumbers, line.PoNumber+"~"+util.GetStringFromInt(line.PoLineItemNumber)) {
			uniquePoAndLineNumbers = append(uniquePoAndLineNumbers, line.PoNumber+"~"+util.GetStringFromInt(line.PoLineItemNumber)
		}
	}

	for _, uniqueId := range uniquePoAndLineNumbers {
		keys := strings.Split(uniqueId, "~")
		var selectedLineItem GRNLineItem
		selectedLineItem.PoNumber=keys[0]
		selectedLineItem.PoLineItemNumber=util.GetIntFromString(keys[1])

		var addedQty, residualQty float64
		addedQty = 0.0
		residualQty = 0.0
		for _, line := range *grn.LineItems {
			if keys[0] == line.PoNumber && util.GetIntFromString(keys[1]) == line.PoLineItemNumber {
				addedQty = addedQty + line.PoQuantity
				if line.ResidualQuantity != 0.0 {
					residualQty = residualQty + line.ResidualQuantity
				} else {
					residualQty = residualQty + addedQty
				}
			}
		}
		selectedLineItem.ResidualQuantity=(residualQty)
		selectedLineItem.PoQuantity=(addedQty)
		selectedLineItems = append(selectedLineItems, selectedLineItem)
	}
	grn.LineItems(selectedLineItems)
	return grn
}*/
