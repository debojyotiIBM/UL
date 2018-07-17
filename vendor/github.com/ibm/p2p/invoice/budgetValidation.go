/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Venkatachalam Vairavan
*/

package invoice

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	//"github.com/ibm/p2p/grn"
	util "github.com/ibm/p2p"
	"github.com/ibm/p2p/po"
)

func Budgetvalidation(stub shim.ChaincodeStubInterface, contextObjPtr *Context) (int, string, InvoiceStatus) {
	var invoice *Invoice = &(contextObjPtr.Invoice)

	var errStr string
	var lineCounter int = 0
	//var pos []po.POLineItem
	var invStat InvoiceStatus

	myLogger.Debugf("*************BudgetValidation*************", invoice)

	for idx, InvoiceLineItem := range invoice.DcDocumentData.DcLines {

		myLogger.Debugf("Invoice Line Items ================", invoice.DcDocumentData.DcHeader.ErpSystem, InvoiceLineItem.PoNumber, util.GetStringFromInt(InvoiceLineItem.PoLine))
		invoice_PO, fetchError := po.GetPOLineItemPartially(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, InvoiceLineItem.PoNumber, util.GetStringFromInt(InvoiceLineItem.PoLine)})

		if fetchError != "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, "PO NUMBER INVALID OR DOES NOT EXIST", "", "st-budgetValidation-6", EMPTY_ADDITIONAL_INFO)
			return 2, "", invStat
		}
		myLogger.Debugf("****************Check Residual Amount and Quantity******************", invoice_PO)
		ResidualQuantity := invoice_PO.ResidualQuantity
		//totalLineValue := invoice_PO.UnitPrice * InvoiceLineItem.Quantity
		myLogger.Debugf("****************validation ******************", InvoiceLineItem.Quantity, ResidualQuantity)
		if InvoiceLineItem.Quantity <= ResidualQuantity {
			myLogger.Debugf("****************validation pass for PO Line number******************", InvoiceLineItem.PoLine)
			lineCounter++
			invoice.DcDocumentData.DcLines[idx].D_outOfBudget = false
			invoice_PO.ResidualQuantity = ResidualQuantity - InvoiceLineItem.Quantity
			//	pos = append(pos, invoice_PO)
			po.AddPOLineItems(stub, invoice_PO.PONumber, invoice_PO.ERPSystem, invoice_PO.Client, invoice_PO.LineItemNumber, string(util.MarshalToBytes(invoice_PO)))

		} else {
			myLogger.Debugf("****************validation fail for PO Line number******************", InvoiceLineItem.PoLine)
			invoice.DcDocumentData.DcLines[idx].D_outOfBudget = true

		}

	}

	AddInvoice(stub, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, string(util.MarshalToBytes(invoice)), false)
	if lineCounter == len(invoice.DcDocumentData.DcLines) {
		/* for _, poLine := range pos {

		} */
		invStat = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", "st-VGRN-start", EMPTY_ADDITIONAL_INFO)
	} else {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_BUDGET_VALIDATION_RCODE, invStat.Comments, "st-budgetValidation-7", EMPTY_ADDITIONAL_INFO)
		return 2, errStr, invStat
	}

	return 1, errStr, invStat
}

func BudgetValidation_IBM_AP_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {
	var invoice *Invoice = &(contextObjPtr.Invoice)
	var errStr string
	if invStat.Status == FWD_TO_BUYER {
		invStat = ForwardToBuyer(stub, invoice, invStat)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-11", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == REJECT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-12", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == SELECT_DESELECT_PO { // Select / Deselect PO line to be updated
		// Identify and move to next step - whether VGRN or GRN
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-1", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	} else if invStat.Status == "RECONSTRUCTED" {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-VGRN-start", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	}
	return 2, errStr, invStat
}

func BudgetValidation_BUYER_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {
	var invoice *Invoice = &(contextObjPtr.Invoice)
	var errStr string
	if invStat.Status == FWD_TO_OTHER_BUYER {
		errStr, invStat = ForwardToOtherBuyer(stub, invoice, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-11", EMPTY_ADDITIONAL_INFO)
		}
	} else if invStat.Status == REJECT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-6", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == RETURN_TO_AP {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-7", EMPTY_ADDITIONAL_INFO)
	} else if invStat.ReasonCode == BUYER_DELEGATION {
		errStr, invStat = BuyerDelegation(stub, contextObjPtr, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, invStat.Status, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-11", EMPTY_ADDITIONAL_INFO) //TODO
		}
	} else if invStat.Status == ALT_PO {
		_, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, invoice.DcDocumentData.DcHeader.PoNumber, invoice.DcDocumentData.DcHeader.Client})
		if fetchErr != "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_INVALID_PO, "", "st-budgetValidation-7", EMPTY_ADDITIONAL_INFO)
			return 2, errStr, invStat
		}
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_WAITING_INVOICE_FIX, invStat.ReasonCode, invStat.Comments, "ST0000", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	} else if invStat.Status == "BUDGET INCREASED" {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, "AWAITING DB REFRESH", invStat.ReasonCode, invStat.Comments, "st-budgetValidation-7", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == "AWAITING DB REFRESH" {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, invStat.Status, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-1", EMPTY_ADDITIONAL_INFO)
		//invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, invStat.Status, invStat.ReasonCode, invStat.Comments, "st-LineSelection-start", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	}
	return 2, errStr, invStat
}
