/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package invoice

import (
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	db "github.com/ibm/db"
	util "github.com/ibm/p2p"
	g "github.com/ibm/p2p/grn"
	p "github.com/ibm/p2p/po"
)

type InvoiceSubmitRequest struct {
	BciId  string `json:"bciId"`
	UserID string `json:"UserID"`
	//	InvoiceNumber string `json:"invoiceNumber"`
	Status       string `json:"status"`
	ReasonCode   string `json:"reasonCode"`
	Comments     string `json:"comments"`
	BuyerId      string `json:"buyerId"`
	BuyerEmailId string `json:"buyerEmailId"`
	ScanID       string `json:"Scan_ID"`
	Logging      bool   `json:"Logging"`
}

type Context struct {
	Invoice       Invoice
	VendorID      string
	StatusHistory []InvoiceStatus
	StoreInvoice  bool
	userIDGlobal  string

	UpdatedPOs      map[string]p.PO
	UpdatedPOLines  map[string]p.POLineItem
	UpdatedGRNs     map[string]g.GRN
	DupInvoiceCache map[string]map[string]DuplicateInvoiceCache
}

func InitCache(contextObjPtr *Context) {
	contextObjPtr.UpdatedPOs = make(map[string]p.PO)
	contextObjPtr.UpdatedPOLines = make(map[string]p.POLineItem)
	contextObjPtr.UpdatedGRNs = make(map[string]g.GRN)
	contextObjPtr.DupInvoiceCache = make(map[string]map[string]DuplicateInvoiceCache)
}

func SubmitInvoice(stub shim.ChaincodeStubInterface, request string) pb.Response {
	var contextObj Context
	InitCache(&contextObj)

	var invoice Invoice
	var invStat InvoiceStatus
	var errStr string
	var invStatArr []InvoiceStatus
	var isr InvoiceSubmitRequest

	err := json.Unmarshal([]byte(request), &isr)
	if isr.Logging {
		db.STATE_DB_OPERATIONS_LOGGING = true
	} else {
		db.STATE_DB_OPERATIONS_LOGGING = false
	}

	if err != nil {
		myLogger.Debugf("ERROR in parsing input submission request :", err)
	}

	invoice, errStr = GetInvoice(stub, []string{isr.ScanID, isr.BciId})
	if errStr != "" {
		return ReturnErrorResponse(stub, &contextObj, errStr)
	}

	contextObj.userIDGlobal = isr.UserID

	contextObj.Invoice = invoice

	invStatArr, errStr = GetInvoiceStatus(stub, []string{isr.ScanID, isr.BciId})
	if errStr != "" {
		return ReturnErrorResponse(stub, &contextObj, "Invoice not added to the ledger. Please add . . . "+errStr)
	}
	contextObj.StatusHistory = invStatArr

	invStat = invStatArr[len(invStatArr)-1]
	if invStat.Status != INV_STATUS_INIT {
		invStat.Status = isr.Status
		invStat.ReasonCode = isr.ReasonCode
		invStat.Comments = isr.Comments
		invStat.BuyerId = isr.BuyerId
		invStat.BuyerEmailId = isr.BuyerEmailId
	}

	invStatusCode := invStat.InternalStatus
	invStat1 := CreateInvoiceStatus(stub, &contextObj, isr.BciId, isr.ScanID, isr.Status, isr.ReasonCode, isr.Comments, invStat.InternalStatus, EMPTY_ADDITIONAL_INFO)
	//STATUS_HISTORY = append(STATUS_HISTORY, invStat1)
	contextObj.StatusHistory = append(contextObj.StatusHistory, invStat1)
	myLogger.Debugf("UserID changes after action", contextObj.StatusHistory)
	switch invStatusCode {
	case "ST0003", "ST0005":
		// Resume at STAGE 00 VendorName match
		invStat.InternalStatus = "ST0501"

		//	case "ST0307", "ST0308", "ST0310", "ST0611":
	case "ST0307", "ST0310":
		invStat.InternalStatus = "ST0201"

	case "ST0301":
		return StateChangeBasedOnAction(stub, &contextObj, isr, invStat)

	case "ST0406":
		invStat.InternalStatus = "ST0000"

	case "ST0909":
		return StateChangeBasedOnAction(stub, &contextObj, isr, invStat)

	default:
		myLogger.Debugf("Initial processing of the invoice")
	}
	return TakeActionBasedOnStatus(stub, &contextObj, invStat)
}

func ReturnErrorResponse(stub shim.ChaincodeStubInterface, contextObjPtr *Context, errStr string) pb.Response {
	if contextObjPtr.StoreInvoice {
		AddInvoice(stub, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, string(util.MarshalToBytes(contextObjPtr.Invoice)), false)
	}
	return shim.Error(errStr)
}

func ReturnPendingResponse(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) pb.Response {
	saveCache(stub, contextObjPtr)
	if contextObjPtr.StoreInvoice {
		AddInvoice(stub, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, string(util.MarshalToBytes(contextObjPtr.Invoice)), false)
	}
	return shim.Success(util.MarshalToBytes(contextObjPtr.StatusHistory))
}

func TakeActionBasedOnStatus(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) pb.Response {
	invStatusCode := invStat.InternalStatus
	myLogger.Debugf("Internal Status : ", invStatusCode)

	// Code added by Lohit - Following code was added to test  submit-invoice without any stage functions.
	// Remove it once you add first stage function
	// --1 From Here
	statusCode := 0
	errStr := ""
	vendorId := ""

	// if invStatusCode == "ST0000" {
	// 	invStatusCode = "st-DuplicateInvoiceCheck-start"
	// }

	// --1 Till Here

	switch invStatusCode {

	case "ST0000":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VALID PO                                                                        ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat, vendorId := MatchRemitToID(stub, contextObjPtr.Invoice)
		statusCode, errStr, invStat, _ := ValidPO(stub, contextObjPtr)
		//		statusCode, errStr, invStat := DeterminePOUnitPrice(stub, contextObjPtr)
		myLogger.Debugf("statusCode**************************************************************************************", statusCode)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-ValidPO-1", "st-ValidPO-2", "st-ValidPO-10", "st-ValidPO-15":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VALID PO  :   ValidPO_IBMAP_Action                                                                    ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat, vendorId := MatchRemitToID(stub, contextObjPtr.Invoice)
		statusCode, errStr, invStat := ValidPO_IBMAP_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-ValidPO-3":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VALID PO   : ValidPO_BUYER_Action                                                                  ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat, vendorId := MatchRemitToID(stub, contextObjPtr.Invoice)
		statusCode, errStr, invStat := ValidPO_BUYER_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-companyCode-1":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Company Code Selection                                                                           ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := SelectCompanyCode(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-companyCode-2", "st-companyCode-13", "st-companyCode-9":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Company Code Selection IBM Action                                                                ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := Select_Company_Code_IBM_AP_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-companyCode-3":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Company Code Selection Buyer Action                                                              ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := Select_Company_Code_Buyer_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-billToName-1":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Bill To Name Selection                                                                           ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := BilltoNameMatch(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-billToName-9", "st-billToName-2", "st-billToName-6":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Bill To Name IBM Action                                                                         ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := BilltoNameMatch_IBM_AP_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)
	case "st-billToName-10":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Bill To Name IBM Action Buyer Action                                                                        ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := BilltoName_Buyer_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-singlevsmultiple-1":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Single Vs Multiple PO                                                                          ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := CheckSingleOrMultiplePo(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-singlevsmultiple-14", "st-singlevsmultiple-2", "st-singlevsmultiple-10":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Single Vs Multiple PO IBM AP Action                                                                        ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := SingleVsMuliple_Po_IBM_AP_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)
	case "st-singlevsmultiple-3":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Single Vs Multiple PO Buyer Action                                                                        ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := SingleVsMuliple_Po_Buyer_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-postfacto-1":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Post Facto PO                                                                       ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := PostfactoPo(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-currencyValidation-1":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Currency Validation                                                                       ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := ValidateCurrency(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0406":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : REMIT TO ID MATCH                                                                            ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat, vendorId := MatchRemitToID(stub, contextObjPtr.Invoice)
		contextObjPtr.VendorID = vendorId
		stub.PutState(contextObjPtr.Invoice.BCIID+contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID+"VENDORID", util.MarshalToBytes(contextObjPtr.VendorID))
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

		// statusCode, errStr, invStat := MatchVendorName(stub, contextObjPtr.Invoice)
		// return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0002", "ST0007":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VENDOR NAME MATCH  - RESUMED BY IBM AP                                                       ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := MatchVendorName_AP_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0401":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : DUPLICATE INVOICE CHECK                                                                      ")
		myLogger.Debugf("******************************************************************************************************")
		if contextObjPtr.VendorID == "" {
			vendorIdState, _ := stub.GetState(contextObjPtr.Invoice.BCIID + contextObjPtr.Invoice.DcDocumentData.DcHeader.InvoiceNumber + "VENDORID")
			contextObjPtr.VendorID = string(vendorIdState)
			if contextObjPtr.VendorID == "" {
				return ReturnErrorResponse(stub, contextObjPtr, "ERROR : Remit to ID not found yet. Please run through MatchRemitToID stage")
			}
		}
		//statusCode, errStr, invStat := VerifyDuplicate(stub, contextObjPtr.Invoice, contextObjPtr.VendorID)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0402", "ST0403", "ST0405", "ST0409", "ST0410", "ST0411":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : REMIT TO ID MATCH  - RESUMED BY IBM AP                                                       ")
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("invStat", invStat)
		myLogger.Debugf("invStat.Status", invStat.Status)
		//statusCode, errStr, invStat := MatchRemitToID_AP_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0407":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : REMIT TO ID MATCH  - RESUMED BY BUYER                                                        ")
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("invStat", invStat)
		//statusCode, errStr, invStat := MatchRemitToID_Buyer_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	/*case "ST0501":
	myLogger.Debugf("******************************************************************************************************")
	myLogger.Debugf("  STEP : VENDOR NAME MATCH                                                                            ")
	myLogger.Debugf("******************************************************************************************************")
	statusCode, errStr, invStat := MatchVendorName(stub, contextObjPtr.Invoice)
	return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)*/

	case "ST0004":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VENDOR NAME MATCH - RESUMED BY BUYER                                                         ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := MatchVendorName_Buyer_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	//case "ST0001":
	case "ST0501":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VENDOR ID MATCH                                                                              ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := MatchVendorId(stub, contextObjPtr.Invoice)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0101":

		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : COMPANY CODE MATCH                                                                           ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := MatchCompanyCode(stub, contextObjPtr.Invoice)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST01001":

		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VERIFY TAX                                                                                   ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := VerifyTax(stub, contextObjPtr.Invoice)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

		//	case "ST0308", "ST0201" ,"ST0611","ST0805":
	case "ST0201":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VERIFY DELETED / BLOCKED / CLOSED PO                                                         ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := VerifyDeletedOrBlockedPO(stub, contextObjPtr.Invoice)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0303":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : LINE ITEM SELECTION                                                                          ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := SelectLineItem(stub, contextObjPtr.Invoice)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0302", "ST0304", "ST0310", "ST0311":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VERIFY DELETED / BLOCKED / CLOSED PO  - RESUMED BY IBM AP                                    ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := VerifyDeletedOrBlockedPO_AP_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0305":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VERIFY DELETED / BLOCKED / CLOSED PO  - RESUMED BY BUYER                                     ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := VerifyDeletedOrBlockedPO_Buyer_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0602", "ST0605", "ST0613":

		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : LINE ITEM SELECTION  - RESUMED BY IBM AP                                                     ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := SelectLineItem_AP_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

		// Venkat code start here
	// Venkat code start here

	case "st-budgetValidation-1":

		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Budget Validation                                                                            ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := Budgetvalidation(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-budgetValidation-7", "st-budgetValidation-10", "st-budgetValidation-12":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Budget Validation IBM AP Action                                                                ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := BudgetValidation_IBM_AP_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-budgetValidation-2", "st-budgetValidation-11":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Budget Validation IBM Buyer Action                                                                ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := BudgetValidation_BUYER_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	// Venkat code end here

	//Pavana Changes
	/*	case "ST0607", "ST0612", "ST0609":

		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VERIFY ADDITIONAL LINE ITEMS                                                                 ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := VerifyAdditionaLineItems(stub, contextObjPtr.Invoice)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)*/

	//Pavana Changes

	case "ST0608":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : LINE ITEM SELECTION  - RESUMED BY IBM AP                                                      ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := SelectLineItem_AP_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0601A":

		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VERIFY ADDITIONAL LINE ITEMS                                                                 ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := VerifyAdditionaLineItems(stub, contextObjPtr.Invoice)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0607A":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : LINE ITEM SELECTION  - RESUMED BY BUYER                                                     ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := SelectLineItem_Buyer_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

		//Pavana Changes

	case "ST0607", "ST0612", "ST0609":

		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VERIFY PO BUDGET                                                                ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := PoBudget(stub, contextObjPtr.Invoice)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0602A", "ST0609A":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VERIFY PO BUDGET  - RESUMED BY IBM AP                                    ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := PoBudget_AP_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0604A":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VERIFY PO BUDGET   - RESUMED BY BUYER                                     ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := PoBudget_Buyer_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0702":

		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VERIFY ADDITIONAL LINE ITEMS - RESUMED BY BUYER                                              ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := VerifyAdditionaLineItems_Buyer_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)
	case "ST0704":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VERIFY ADDITIONAL LINE ITEMS  - RESUMED BY IBMAP                                                     ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := VerifyAdditionaLineItems_AP_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

		//Pavana Changes

	case "ST0701", "ST0703", "ST0905", "ST0909":

		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : GRN QUANTITY MATCH                                                                           ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := MatchQuantityWithGRN(stub, contextObjPtr.Invoice)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0903", "ST0907", "ST0908":

		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : GRN QUANTITY MATCH  - RESUMED BY BUYER                                                       ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := MatchQuantityWithGRN_Buyer_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0901":

		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : MATCH PRICE PER UNIT                                                                         ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := MatchPricePerUnit(stub, contextObjPtr.Invoice)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST0802":

		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : MATCH PRICE PER UNIT  - RESUMED BY BUYER                                                     ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := MatchPricePerUnit_Buyer_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST01101":

		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : REPORTABILITY                                                     ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := VerifyReportability(stub, contextObjPtr.Invoice)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST01103":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : REPORTABILITY    - RESUMED BY BUYER                                                  ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := Reportabilty_Buyer_Action(stub, contextObjPtr.Invoice, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "ST01201":

		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : POST FACTO PO                                                     ")
		myLogger.Debugf("******************************************************************************************************")
		//statusCode, errStr, invStat := VerifyPoCreationDate(stub, contextObjPtr.Invoice)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-lineAggregation-1":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Line Aggregation                                                                     ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := LineAggregration(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-bciValidation-1":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : BCI Validation                                                                   ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := BCIValidation(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-remitToIid-1":
		myLogger.Debugf("*********************************************")
		myLogger.Debugf("STEP : Remit To ID                           ")
		myLogger.Debugf("*********************************************")
		statusCode, errStr, invStat := RemitToId(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-remitToId-21", "st-remitToId-31", "st-remitToId-22":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Remit To Id Buyer Action Selection                                                            ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := Select_Remit_TO_ID_Buyer_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-remitToId-04", "st-remitToId-03", "st-remitToId-02", "st-remitToId-01", "st-remitToId-33":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Remit To Id Selection IBM AP Action                                                                ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := Select_Remit_TO_ID_IBM_AP_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-DuplicateInvoiceCheck-start":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Duplicate Invoice Check                                                                          ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := PerformDuplicateInvoiceCheck(stub, contextObjPtr, false)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-DuplicateInvoiceCheck-InvNumChanged":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Duplicate Invoice Check with Invoice Number Changed                                                                         ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := PerformDuplicateInvoiceCheck(stub, contextObjPtr, true)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-DuplicateInvoiceCheck-ap1":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Duplicate Invoice Check IBM AP Action                                                                               ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := DuplicateInvoiceCheck_IBM_AP_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-LineSelection-start":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Line Selection                                                                               ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := PerformLineSelection(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-LineSelection-ap2", "st-LineSelection-ap3", "st-LineSelection-ap4", "st-LineSelection-ap5", "st-LineSelection-ap6", "st-LineSelection-ap7":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Line Selection IBM AP Action                                                                        ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := LineSelection_IBM_AP_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-LineSelection-buy2", "st-LineSelection-buy3", "st-LineSelection-buy4", "st-LineSelection-buy5", "st-LineSelection-buy6", "st-LineSelection-buy7":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Line Selection Buyer Action                                                                        ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := LineSelection_Buyer_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-determinePOUnitPrice-1":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Determine PO Unit Price                                                                        ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := DeterminePOUnitPrice(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-VGRN-start":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VGRN                                                                               ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := PerformVGRNCheck(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)
	case "st-VGRN-ap5", "st-VGRN-ap6":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VGRN IBM AP Action                                                                        ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := VGRN_IBM_AP_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)
	case "st-VGRN-buyer1", "st-VGRN-buyer2":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : VGRN Buyer Action                                                                        ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := VGRN_Buyer_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-GRNSelection-start", "st-GRNSelection-1":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : GRN Selection                                                                               ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := PerformGRNSelection(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)
	case "st-GRNSelection-ap5", "st-GRNSelection-ap6", "st-GRNSelection-ap7":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : GRN Selection IBM AP Action                                                                        ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := GRNSelection_IBM_AP_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)
	case "st-GRNSelection-buyer1", "st-GRNSelection-buyer2":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : GRN Selection Buyer Action                                                                        ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := GRNSelection_Buyer_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-DCGR-1":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : DCGR                                                                         ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := PerformDCGR(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	case "st-commonUnitOfMeasurement-1":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Common Unit Of Measurement                                                                               ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := ValidateCommonUnitOfMeasurement(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)
	case "st-commonUnitOfMeasurement-2":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Common Unit Of Measurement IBM AP Action                                                                             ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := CommonUnitOfMeasurement_IBM_AP_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)
	case "st-commonUnitOfMeasurement-3":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Common Unit Of Measurement Buyer Action                                                                         ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := CommonUnitOfMeasurement_Buyer_Action(stub, contextObjPtr, invStat)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)
	case "st-vendNameAuth-1":
		myLogger.Debugf("******************************************************************************************************")
		myLogger.Debugf("  STEP : Vendor Name Authentication                                                                      ")
		myLogger.Debugf("******************************************************************************************************")
		statusCode, errStr, invStat := VendorNameAuthentication(stub, contextObjPtr)
		return checkStatus(stub, contextObjPtr, statusCode, errStr, invStat)

	default:
		StoreInvoiceStatusHistory(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, "")
		if contextObjPtr.StoreInvoice {
			AddInvoice(stub, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, string(util.MarshalToBytes(contextObjPtr.Invoice)), false)
		}
		return shim.Success(nil)
	}
}

func checkStatus(stub shim.ChaincodeStubInterface, contextObjPtr *Context, statusCode int, errStr string, invStat InvoiceStatus) pb.Response {
	myLogger.Debugf("Inside checkStatus..................................... statusCode", statusCode)
	if statusCode == 0 || errStr != "" {
		return ReturnErrorResponse(stub, contextObjPtr, errStr)
	}
	if statusCode == 1 {
		return TakeActionBasedOnStatus(stub, contextObjPtr, invStat)
	}
	if statusCode == 2 {
		return ReturnPendingResponse(stub, contextObjPtr, invStat)
	}
	return shim.Success(nil)
}

func StateChangeBasedOnAction(stub shim.ChaincodeStubInterface, contextObjPtr *Context, isr InvoiceSubmitRequest, invStat InvoiceStatus) pb.Response {
	status := invStat.Status
	int_status := invStat.InternalStatus

	if isr.Status == BCI_ACTION_EMAIL_TRIGGERED {
		if invStat.InternalStatus == "ST0301" {
			status = INV_STATUS_PENDING_BUYER
			int_status = "ST0305"
		}
	}

	if invStat.InternalStatus == "ST0909" && isr.Status == USR_AP_ACT_REJECTED {
		status = USR_BUYER_ACT_REJECTED
		int_status = "ST0908"
	}

	SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, status, invStat.ReasonCode, invStat.Comments, int_status, invStat.AdditionalInfo)
	return shim.Success(nil)
}

func saveCache(stub shim.ChaincodeStubInterface, contextObjPtr *Context) {
	//Save all the cache in contextObjPtr.
	myLogger.Debugf("Saving GRN and POResiduals")
	StoreGRNResiduals(stub, contextObjPtr)
	StorePOResiduals(stub, contextObjPtr)
}
