/*
   Copyright IBM Corp. 2018 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Lohit Krishnan
*/

package invoice

import (
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	"github.com/ibm/p2p/grn"
	"github.com/ibm/p2p/po"
)

func PerformVGRNCheck(stub shim.ChaincodeStubInterface, contextObjPtr *Context) (int, string, InvoiceStatus) {
	const NEXT_STEP = "st-TaxHandling-1"
	const GRNSelectionStep = "st-GRNSelection-start"

	var errStr string
	var invStat InvoiceStatus

	var totalACValue float64

	for _, dcLine := range contextObjPtr.Invoice.DcDocumentData.AdditionalLineItems {
		totalACValue += dcLine.Amount
	}

	var totalInvoiceValue = contextObjPtr.Invoice.DcDocumentData.DcSwissHeader.TotalNet - totalACValue

	result, err := checkInvoiceForVGRNProcessing(stub, &contextObjPtr.Invoice)
	if err != "" {
		switch err {
		case "DUPLICATE POLINES DETECTED":
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, err, "", "st-VGRN-ap1", EMPTY_ADDITIONAL_INFO)
		case "POLINES NOT FOUND FOR INVOICE":
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, err, "", "st-VGRN-ap2", EMPTY_ADDITIONAL_INFO)
		case "MATCHAGAINSTGRN FLAG ISSUE":
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, err, "", "st-VGRN-ap3", EMPTY_ADDITIONAL_INFO)
		default:
			errStr = "UNKNOWN ERROR"
		}
		return 2, errStr, invStat
	}
	myLogger.Debugf("VGRN : After VGRN Check : ", result)

	switch result {
	case "3way":
		//GRN selection step is performed in this case.
		invStat = UpdateInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", GRNSelectionStep, EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	case "both":
		// If invoice has lines with both VGRN and 3-way match, fail it to IBM-AP.
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_COMPLETED, EXIST_2WAY_3WAY_VGRN, "", "st-VGRN-end", EMPTY_ADDITIONAL_INFO)
		return 2, errStr, invStat
	default:
		break
	}

	//Do the VGRN Check
	//Collect all the VGRNs for this PO, POLine
	// Function getPOxActivePOLines_NonActivePOLinesForInvoice is written in LineSelection.go
	mPOxActivePOLines_NonActivePOLines := getPOxActivePOLines_NonActivePOLinesForInvoice(stub, &contextObjPtr.Invoice)

	myLogger.Debugf("VGRN : Number of POs in the map = ", len(mPOxActivePOLines_NonActivePOLines))

	grnsActivePOLines := getGRNSForActivePOLines(stub, &mPOxActivePOLines_NonActivePOLines)

	myLogger.Debugf("VGRN : Number of GRNs for Active PO Lines = ", len(grnsActivePOLines))

	vgrnsActivePOLines := filterVGRNS(&grnsActivePOLines)

	myLogger.Debugf("VGRN : Number of VGRNS  = ", len(vgrnsActivePOLines))
	//3.13.3

	if len(vgrnsActivePOLines) >= 1 {
		//3.13.4
		AggregateGRNValue := GetGRNValue(stub, contextObjPtr, vgrnsActivePOLines)
		myLogger.Debugf("VGRN : AggregateGRN Value = ", AggregateGRNValue)
		myLogger.Debugf("VGRN : TotalInvoiceValue = ", totalInvoiceValue)
		if AggregateGRNValue < totalInvoiceValue {
			//3.13.15 & 3.13.6
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_BUYER, INSUFFICIENT_FUND_VGRN, "", "st-VGRN-buyer1", EMPTY_ADDITIONAL_INFO)
			return 2, errStr, invStat
		} else {
			invStat := UpdateInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", NEXT_STEP, EMPTY_ADDITIONAL_INFO)
			return 1, "", invStat
		}

	}
	invStat = UpdateInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", NEXT_STEP, EMPTY_ADDITIONAL_INFO)
	return 1, "", invStat
}

func filterVGRNS(grns *[]grn.GRN) []grn.GRN {
	var result []grn.GRN
	for _, GRN := range *grns {
		if GRN.MatDocNumber == "VGRN" {
			result = append(result, GRN)
		}
	}
	return result

}

// This function returns (result, error) where
// result can be either (vgrn, grn or both)
// If poLine.GoodsReceiptFlag == 'X' then it is 3 way, else it is 2 way.
//If there are both 2way and 3way in different po lines for the same invoice, then return "both"
func checkInvoiceForVGRNProcessing(stub shim.ChaincodeStubInterface, invoice *Invoice) (string, string) {
	var result = ""
	for _, dcLine := range invoice.DcDocumentData.DcLines {
		cPONumber := dcLine.PoNumber
		cPOLine := util.GetStringFromInt(dcLine.PoLine)
		cERPSystem := invoice.DcDocumentData.DcHeader.ErpSystem
		poLineItemsRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS, PrimaryKeys: []string{cERPSystem, cPONumber, cPOLine}, Data: ""}.GetAll()
		if len(poLineItemsRec) > 1 {
			// Error : There are two entries with the same <ERPSystem, PO, POLine> combination
			// Reason : "Duplicate POLines detected", Status : "IBM-AP Action Pending"
			return "", DUPLICATE_POLINES
		}
		if len(poLineItemsRec) == 0 {
			// Error : There are no entries with the same <ERPSystem, PO, POLine> combination
			// Reason : "No POLines found for the InvoiceLines", Status : "IBM-AP Action Pending"
			return "", POLINES_NOT_FOUND
		}
		for _, poLineItemRow := range poLineItemsRec {
			var currentPOLineItem po.POLineItem
			json.Unmarshal([]byte(poLineItemRow), &currentPOLineItem)
			if currentPOLineItem.MatchAgainstGrn == "X" {
				if result == "VGRN" {
					result = "both"
					return result, ""
				}
				result = "3way"
			} else if currentPOLineItem.MatchAgainstGrn == "" {
				if result == "3way" {
					result = "both"
					return result, ""
				}
				result = "VGRN"
			} else {
				return "", MATCHAGAINSTGRN_FLAG_ISSUE
			}
		}
	}
	return result, ""
}

func VGRN_IBM_AP_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {

	const NEXT_STEP = "st-TaxHandling-1"

	var errStr string
	if invStat.Status == FWD_TO_BUYER {
		invStat = ForwardToBuyer(stub, &contextObjPtr.Invoice, invStat)
		newInternalStatus := "st-VGRN-buyer2"
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, newInternalStatus, EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == REJECT {
		RevertGRNResidualForCompleteInvoice(stub, contextObjPtr)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-VGRN-end", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == PROCESS_OUTSIDE_BC {
		RevertGRNResidualForCompleteInvoice(stub, contextObjPtr)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, COMPLETED_OUTSIDE_BC, invStat.ReasonCode, invStat.Comments, "st-VGRN-end", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == RECONSTRUCTED {
		BUDGET_VALIDATION_STEP := "st-budgetValidation-1"
		invStat = UpdateInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, BUDGET_VALIDATION_STEP, EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	} else if invStat.Status == "" || invStat.Status == CONTINUE {
		invStat = UpdateInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, NEXT_STEP, EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	}
	return 2, errStr, invStat
}

func VGRN_Buyer_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {

	// UI STATUS TO BE HANDLED
	// 1) AWAITING BUYER ACTION HOLD INVOICE
	var errStr string
	if invStat.Status == FWD_TO_OTHER_BUYER {

		errStr, invStat = ForwardToOtherBuyer(stub, &contextObjPtr.Invoice, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, invStat.InternalStatus, EMPTY_ADDITIONAL_INFO)
		}
	} else if invStat.Status == REJECT {
		RevertGRNResidualForCompleteInvoice(stub, contextObjPtr)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-VGRN-end", EMPTY_ADDITIONAL_INFO)

	} else if invStat.Status == RETURN_TO_AP {
		newInternalStatus := "st-VGRN-ap5"
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, newInternalStatus, EMPTY_ADDITIONAL_INFO)

	} else if invStat.Status == BUYER_DELEGATION {

		errStr, invStat = BuyerDelegation(stub, contextObjPtr, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, invStat.Status, invStat.ReasonCode, invStat.Comments, invStat.InternalStatus, EMPTY_ADDITIONAL_INFO)
		}
	} else if invStat.Status == ALT_PO {
		RevertGRNResidualForCompleteInvoice(stub, contextObjPtr)
		myLogger.Debugf("VGRN : entered alternate PO loop=====================")

		invStat = UpdateInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, ALT_PO, invStat.ReasonCode, invStat.Comments, "ST0000", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat

	} else if invStat.Status == ADDITIONAL_PO {
		RevertGRNResidualForCompleteInvoice(stub, contextObjPtr)
		myLogger.Debugf("VGRN : entered Additional PO loop=====================")

		invStat = UpdateInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, ALT_PO, invStat.ReasonCode, invStat.Comments, "ST0000", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat

	} else if invStat.Status == PO_BUDGET_INCREASED {
		BUDGET_VALIDATION_STEP := "st-budgetValidation-1"
		invStat = UpdateInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_WAITING_PO_REFRESH, invStat.ReasonCode, invStat.Comments, BUDGET_VALIDATION_STEP, EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	}
	return 2, errStr, invStat
}
