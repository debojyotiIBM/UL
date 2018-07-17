/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package po

import (
	"encoding/json"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	//	"github.com/jinzhu/copier"
	logging "github.com/op/go-logging"
)

type PO struct {
	ERPSystem      string      `json:"erpsystem"`
	Client         string      `json:"client"`
	PONumber       string      `json:"ponumber"`
	CompanyCode    string      `json:"companycode"`
	VendorId       string      `json:"vendorid"`
	Currency       string      `json:"currency"`
	PoDate         util.BCDate `json:"podate"`
	InvoicingParty string      `json:"invoicingparty"`
	BuyerID        string      `json:"buyerid"`
	BuyerEmailId   string      `json:"buyeremailid"`

	DeletionFlag      string      `json:"deletionflag"`
	POStatus          string      `json:"postatus"`
	CreationDate      util.BCDate `json:"creationdate"`
	PaymentTerms      string      `json:"paymentterms"`
	SuplyingVendor    string      `json:"suplyingvendor"`
	NameCreationUser  string      `json:"namecreationuser"`
	RequestOrName     string      `json:"requestorname"`
	CreationUserEmail string      `json:"creationuseremail"`
}

type POLineItem struct {
	ERPSystem        string  `json:"erpsystem"`
	Client           string  `json:"client"`
	PONumber         string  `json:"ponumber"`
	LineItemNumber   int64   `json:"lineitemnumber"`
	PoStatus         string  `json:"postatus"`
	Description      string  `json:"description"`
	UOM              string  `json:"uom"`
	Quantity         float64 `json:"quantity"`
	Per              float64 `json:"per"`
	UnitPrice        float64 `json:"unitprice"`
	NetOrderValue    float64 `json:"netordervalue"`
	MaterialNumber   string  `json:"materialnumber"`
	ResidualQuantity float64 `json:"residualquantity"`
	SecondaryUOM     string  `json:"secondaryuom"`
	BuyerID          string  `json:"buyerid"`
	PaymentTerms     string  `json:"paymentterms"`

	DeletionFlag       string  `json:"deletionflag"`
	CompanyCode        string  `json:"companycode"`
	OrderPriceUnit     string  `json:"orderpriceunit"`
	GrossOrderValue    float64 `json:"grossordervalue"`
	DeliveryCompleted  string  `json:"deliverycompleted"`
	IsFinalInvoice     string  `json:"isfinalinvoice"`
	InvoiceReceiptFlag string  `json:"invoicereceptionflag"`
	GoodsReceiptFlag   string  `json:"goodsreceiptflag"`
	MatchAgainstGrn    string  `json:"matchagainstgrn"`
	ServiceFlag        string  `json:"serviceflag"`
	ERS                string  `json:"ers"`
	ResidualAmount     float64 `json:"residualamount"`
}

var myLogger = logging.MustGetLogger("Procure-To-Pay : PO")

func FullLoadPoRecords(stub shim.ChaincodeStubInterface, PORecArr string) pb.Response {
	var pos []PO
	err := json.Unmarshal([]byte(PORecArr), &pos)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input PO array:", err)
	}
	for _, po := range pos {
		db.TableStruct{Stub: stub, TableName: util.TAB_PO, PrimaryKeys: []string{po.ERPSystem, po.PONumber, po.Client}, Data: string(util.MarshalToBytes(po))}.Add()
	}
	return shim.Success(nil)
}

func AddPoLineItemRecords(stub shim.ChaincodeStubInterface, PORecArr string) pb.Response {
	var poLines []POLineItem
	var posByBuyer map[string]string
	posByBuyer = make(map[string]string)
	var polinesByBuyer map[string]string
	polinesByBuyer = make(map[string]string)
	err := json.Unmarshal([]byte(PORecArr), &poLines)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input PO array:", err)
	}
	for _, poLine := range poLines {

		if !isPOLineItemPresent(stub, []string{poLine.ERPSystem, poLine.PONumber, util.GetStringFromInt(poLine.LineItemNumber), poLine.Client}) {
			myLogger.Debugf("AddPOLineItemRecords : POLineitem not present")
			poLine.ResidualQuantity = poLine.Quantity
		} // todo :
		// TODO : This we are doing to populate buyerID in PO. This is Hot Fix. Need to discuss with Gopi on the same. - Starts
		postruct, poerr := GetPO(stub, []string{poLine.ERPSystem, poLine.PONumber, poLine.Client})

		if poerr != "" {
			myLogger.Debugf("NO Po present in the DB :", poerr)
		}
/*
		Ashish Confirmed that the buyerID will be populated in the PO Header, so commenting the below hot fix.
		myLogger.Debugf("Buyer ID============", poLine.BuyerID)
		postruct.BuyerID = poLine.BuyerID
		AddPO(stub, postruct.PONumber, postruct.ERPSystem, postruct.Client, string(util.MarshalToBytes(postruct)))
		//	db.TableStruct{Stub: stub, TableName: util.TAB_PO_BY_BUYER, PrimaryKeys: []string{buyerId}, Data: string(util.MarshalToBytes(postruct))}.Add()
*/
		// collect pos by buyer

		if postruct.BuyerID != "" {
			if posByBuyer[postruct.BuyerID] != "" {
				posByBuyer[postruct.BuyerID] = posByBuyer[postruct.BuyerID] + "|" + postruct.ERPSystem + "~" + postruct.PONumber + "~" + postruct.Client
			} else {
				posByBuyer[postruct.BuyerID] = postruct.ERPSystem + "~" + postruct.PONumber + "~" + postruct.Client
			}
		}

		myLogger.Debugf("Buyer map=========", posByBuyer)

		// collect po Line Items by buyer
		if poLine.BuyerID != "" {
			if polinesByBuyer[poLine.BuyerID] != "" {
				polinesByBuyer[poLine.BuyerID] = polinesByBuyer[poLine.BuyerID] + "|" + poLine.ERPSystem + "~" + poLine.PONumber + "~" + util.GetStringFromInt(poLine.LineItemNumber) + "~" + poLine.Client
			} else {
				polinesByBuyer[poLine.BuyerID] = poLine.ERPSystem + "~" + poLine.PONumber + "~" + util.GetStringFromInt(poLine.LineItemNumber) + "~" + poLine.Client
			}
		}

		// TODO : This we are doing to populate buyerID in PO. This is Hot Fix. Need to discuss with Gopi on the same. - Ends

		AddPOLineItems(stub, poLine.PONumber, poLine.ERPSystem, poLine.Client, poLine.LineItemNumber, string(util.MarshalToBytes(poLine)))
	}

	for buyerId, val := range posByBuyer {
		myLogger.Debugf("List of po's for particular buyer", val)
		util.UpdateReferenceData(stub, util.TAB_PO_BY_BUYER, []string{buyerId}, val)
	}
	for buyerId, val := range polinesByBuyer {
		util.UpdateReferenceData(stub, util.TAB_PO_LINEITEMS_BY_BUYER, []string{buyerId}, val)
	}
	return shim.Success(nil)
}

func POLineMatch(line1 POLineItem, line2 POLineItem) bool {
	return (line1.LineItemNumber == line2.LineItemNumber)
}

func IsBitSet(flag string, pos int) bool {

	if len(flag) < pos {
		return false
	}

	if string(flag[pos-1]) == "1" {
		return true
	} else {
		return false
	}
}

func AddPoRecords(stub shim.ChaincodeStubInterface, PORecArr string) pb.Response {
	var pos []PO
	/*var posByBuyer map[string]string
	posByBuyer = make(map[string]string)*/

	err := json.Unmarshal([]byte(PORecArr), &pos)

	if err != nil {
		myLogger.Debugf("ERROR in parsing input PO array:", err)
	}

	for _, po := range pos {
		AddPO(stub, po.PONumber, po.ERPSystem, po.Client, string(util.MarshalToBytes(po)))
	}
	return shim.Success(nil)

}

func AddPO(stub shim.ChaincodeStubInterface, poNumber string, erpsystem string, client string, poStr string) {
	db.TableStruct{Stub: stub, TableName: util.TAB_PO, PrimaryKeys: []string{erpsystem, poNumber, client}, Data: string(poStr)}.Add()

}

func AddPOLineItems(stub shim.ChaincodeStubInterface, poNumber string, erpsystem string, client string, lineItemNum int64, poStr string) {
	litemNum := util.GetStringFromInt(lineItemNum)
	myLogger.Debugf("line item number===============", litemNum)
	db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS, PrimaryKeys: []string{erpsystem, poNumber, litemNum, client}, Data: string(poStr)}.Add()
	//db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS_BY_BUYER, PrimaryKeys: []string{buyerId}, Data: string(poStr)}.Add()

}

func GetPosByBuyerID(stub shim.ChaincodeStubInterface, buyerID string) ([]PO, error) {
	record, fetchErr := db.TableStruct{Stub: stub, TableName: util.TAB_PO_BY_BUYER, PrimaryKeys: []string{buyerID}, Data: ""}.Get()
	myLogger.Debugf("po header for partocular buyer=============", record)
	if fetchErr != nil {
		myLogger.Debugf("ERROR in parsing po :", fetchErr)
	}
	var pos []PO
	if record != "" {
		poNumbers := strings.Split(record, "|")
		for _, poNumber := range poNumbers {
			pokeys := strings.Split(poNumber, "~")
			po, _ := GetPO(stub, pokeys) // PO number would be stored as  erpsystem~client~PoNumber
			pos = append(pos, po)
		}
	}
	myLogger.Debugf("PO for buyer=====================", pos)
	return pos, nil
}

// Add po line items to buyer
func GetPoLineitemsByBuyerID(stub shim.ChaincodeStubInterface, buyerID string) ([]POLineItem, error) {
	record, fetchErr := db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS_BY_BUYER, PrimaryKeys: []string{buyerID}, Data: ""}.Get()
	if fetchErr != nil {
		myLogger.Debugf("ERROR in parsing po :", fetchErr)
	}
	var pos []POLineItem
	if record != "" {
		poNumbers := strings.Split(record, "|")
		for _, poNumber := range poNumbers {
			pokeys := strings.Split(poNumber, "~")
			po, _ := GetPOLineItem(stub, pokeys) // PO number would be stored as  erpsystem~client~PoNumber
			pos = append(pos, po)
		}
	}
	return pos, nil
}

func GetPO(stub shim.ChaincodeStubInterface, keys []string) (PO, string) {
	poRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO, PrimaryKeys: keys, Data: ""}.Get()
	var po PO
	err := json.Unmarshal([]byte(poRecord), &po)
	if err != nil {
		myLogger.Debugf("ERROR parsing PO  :", err)
		return po, "ERROR parsing PO"
	}
	return po, ""
}

func GetPOLineItem(stub shim.ChaincodeStubInterface, keys []string) (POLineItem, string) {
	myLogger.Debugf("Get po line items key=============", keys)
	poRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS, PrimaryKeys: keys, Data: ""}.Get()
	var po POLineItem
	err := json.Unmarshal([]byte(poRecord), &po)
	if err != nil {
		myLogger.Debugf("ERROR parsing PO  :", err)
		return po, "ERROR parsing PO"
	}
	return po, ""
}

func isPOLineItemPresent(stub shim.ChaincodeStubInterface, keys []string) bool {
	myLogger.Debugf("Get po line items key=============", keys)
	poRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS, PrimaryKeys: keys, Data: ""}.Get()
	if poRecord == "" {
		return false
	}
	return true
}

func GetPOByVendor(stub shim.ChaincodeStubInterface, vendorId string) []PO {
	var vendorsPO []PO
	posRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, poRow := range posRec {
		var currentPo PO
		json.Unmarshal([]byte(poRow), &currentPo)
		if currentPo.VendorId == vendorId {
			vendorsPO = append(vendorsPO, currentPo)
		}
	}
	return vendorsPO
}

func GetAllPOs(stub shim.ChaincodeStubInterface) []PO {
	var allPOs []PO
	posRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, poRow := range posRec {
		var currentPo PO
		json.Unmarshal([]byte(poRow), &currentPo)
		allPOs = append(allPOs, currentPo)
	}

	return allPOs
}

func GetAllPOLineItemsByPO(stub shim.ChaincodeStubInterface, keys []string) []POLineItem {

	var allPOLineItems []POLineItem
	poLineItems, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS, PrimaryKeys: keys, Data: ""}.GetAll()
	for _, poLineItem := range poLineItems {
		var currentPoLineItem POLineItem
		json.Unmarshal([]byte(poLineItem), &currentPoLineItem)
		allPOLineItems = append(allPOLineItems, currentPoLineItem)
	}
	return allPOLineItems
}

func GetPOLineItemPartially(stub shim.ChaincodeStubInterface, keys []string) (POLineItem, string) {
	poRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS, PrimaryKeys: keys, Data: ""}.GetAll()
	var po POLineItem
	var isRecordExists bool
	for _, poRow := range poRecord {
		json.Unmarshal([]byte(poRow), &po)
		isRecordExists = true
		break
	}
	if !isRecordExists {
		return po, "Po Line doesn't Exist"
	}
	return po, ""
}