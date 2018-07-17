/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pavana C K
*/

package invoice

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	//"github.com/ibm/p2p/grn"
	util "github.com/ibm/p2p"
	"github.com/ibm/p2p/po"
	"strconv"
	"strings"
)

/*
TBD
*/

func Budgetvalidation(stub shim.ChaincodeStubInterface, contextObjPtr *Context) (int, string, InvoiceStatus) {
	var invoice *Invoice = &(contextObjPtr.Invoice)

	var errorMessage string
	var invoiceStatus InvoiceStatus
	var budgetStatus bool
	var poLinenumber int64
	var poNumber string
	var invoiceQuanity float64
	var ResidualQuantity float64
	var ResidualAmount float64
	var poUnitPrice float64
	var totalAmount float64

	myLogger.Debugf("*************BudgetValidation*************", invoice)

	for _, InvoiceLineItem := range invoice.DcDocumentData.DcLines {

		myLogger.Debugf("Invoice Line Items ================", InvoiceLineItem)

		poLinenumber = strconv.Itoa(int(InvoiceLineItem.PoLine)) 
		poNumber = InvoiceLineItem.PoNumber
		invoiceQuanity := InvoiceLineItem.Quantity

		// if(poLinenumber == nil || strings.TrimSpace(poLinenumber) == ""){
		// 	invoiceStatus, errorMessage = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_INVALID_PO, "", "ST0000", EMPTY_ADDITIONAL_INFO)
		// 	return 0, "ERROR Parsing PO in Budgetvalidation", invoiceStatus
		// }else{
				// []string{erpsystem, poNumber, litemNum, client}
				//invoice_PO, fetchError := po.GetPOLineItem(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, invoice.DcDocumentData.DcHeader.Client, InvoiceLineItem.PoNumber, strconv.Itoa(int(poLinenumber))})
				invoice_PO, fetchError := po.GetPOLineItem(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, invoice.DcDocumentData.DcHeader.Client, InvoiceLineItem.PoNumber, poLinenumber})

				if fetchError != "" {
					//invoiceStatus, errorMessage = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_INVALID_PO, "", "st-budgetValidation-10", EMPTY_ADDITIONAL_INFO)
					invoiceStatus, errorMessage = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_INVALID_PO, "", "ST0000", EMPTY_ADDITIONAL_INFO)
					return 0, "ERROR Parsing PO in Budgetvalidation", invoiceStatus
				} else {
					myLogger.Debugf("****************Check Residual Amount and Quantity******************")
					poUnitPrice := invoice_PO.UnitPrice
					ResidualQuantity := invoice_PO.ResidualQuantity
					ResidualAmount := invoice_PO.ResidualAmount //TODO: Sharath : ResidualAmount does not exists in PO_Line. Plz chk.
					totalAmount := poUnitPrice * invoiceQuanity

					if (totalAmount <= ResidualAmount && invoiceQuanity <= ResidualQuantity ) {
						myLogger.Debugf("****************validation pass for PO Line number******************", poLinenumber)
						budgetStatus = true
					} else {
						myLogger.Debugf("****************validation fail for PO Line number******************", poLinenumber)
						budgetStatus = false
						break
					}
				}
		//}
		
	
	}

	if budgetStatus {
		myLogger.Debugf("Budget Validation is success. Update the Residual Quantity and Amount in PO table")
		for _, InvoiceLineItem := range invoice.DcDocumentData.DcLines {

			myLogger.Debugf("Updating Residual Payment ================", InvoiceLineItem)

			poLinenumber = InvoiceLineItem.PoLine
			poNumber = InvoiceLineItem.PoNumber
			invoiceQuanity := InvoiceLineItem.Quantity

			invoice_PO, fetchError := po.GetPOLineItem(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, InvoiceLineItem.PoNumber, invoice.DcDocumentData.DcHeader.Client, strconv.Itoa(int(InvoiceLineItem.PoLine))})

			if fetchError != "" {
				//	invoiceStatus, errorMessage = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_BUDGET_VALIDATION, "", "st-budgetValidation-10", EMPTY_ADDITIONAL_INFO)
				invoiceStatus, errorMessage = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_INVALID_PO, "", "ST0000", EMPTY_ADDITIONAL_INFO)
				return 0, "ERROR Parsing PO in Budgetvalidation", invoiceStatus
			} else {
				myLogger.Debugf("****************Validate Residual Payment******************")
				poUnitPrice = invoice_PO.UnitPrice
				ResidualQuantity = invoice_PO.ResidualQuantity
				ResidualAmount = invoice_PO.ResidualAmount //TODO: Sharath : ResidualAmount does not exists in PO_Line. Plz chk.
				
				totalAmount = poUnitPrice * invoiceQuanity
				myLogger.Debugf("ResidualAmount = ", ResidualAmount)
				myLogger.Debugf("totalAmount = ", totalAmount)
				invoice_PO.ResidualQuantity = ResidualQuantity - invoiceQuanity
				invoice_PO.ResidualAmount = ResidualAmount - totalAmount //TODO: Sharath : ResidualAmount does not exists in PO_Line. Plz chk.
				
				po.AddPOLineItems(stub, poNumber, invoice.DcDocumentData.DcHeader.ErpSystem, invoice.DcDocumentData.DcHeader.Client, poLinenumber, string(util.MarshalToBytes(invoice_PO)))
			}

				// Get RefDocNumber value from GRN
				// Call GRN method
				//  if RefDocNumber === "VGRN" // then move to VGRN smart contact
				 	invoiceStatus = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", "st-VGRN-01", EMPTY_ADDITIONAL_INFO)
				// else
				// 	invoiceStatus = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", "st-GRN-01", EMPTY_ADDITIONAL_INFO)
				// 	return 1, errorMessage, invoiceStatus
				// }
		}
	} else {
		myLogger.Debugf("Exceed the residual payment and quantity")
		invoiceStatus, errorMessage = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_BUDGET_VALIDATION_RCODE, "", "st-budgetValidation-2", EMPTY_ADDITIONAL_INFO)
		return 2, errorMessage, invoiceStatus
	}
	return 2, errorMessage, invoiceStatus // TODO : Sharath. I have added this return. It was missing. Plz check the values.
}

func BudgetValidation_IBM_AP_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {

	var errStr string
	if invStat.Status == FWD_TO_BUYER {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-11", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == REJECT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-12", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == SELECT_DESELECT_PO { // Select / Deselect PO line to be updated
		// Identify and move the next step - whether VGRN or GRN
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-grn-01", EMPTY_ADDITIONAL_INFO)
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
			// invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, invStat.Status, invStat.ReasonCode, invStat.Comments, invStat.InternalStatus, EMPTY_ADDITIONAL_INFO)
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr,invStat.BciId, invStat.ScanID, invStat.Status, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-19", EMPTY_ADDITIONAL_INFO) //TODO
		}
	} else if invStat.Status == ALT_PO {
			myLogger.Debugf("entered alternate PO loop=====================")
			_, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, invoice.DcDocumentData.DcHeader.PoNumber, invoice.DcDocumentData.DcHeader.Client})
			if fetchErr != "" {
				invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_INVALID_PO, "", "st-ValidPO-1", EMPTY_ADDITIONAL_INFO)
				return 2, errStr, invStat
			}
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_WAITING_INVOICE_FIX, invStat.ReasonCode, invStat.Comments, "ST0000", EMPTY_ADDITIONAL_INFO)
			return 1, errStr, invStat
	} else if invStat.Status == ADDITIONAL_PO {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-10", EMPTY_ADDITIONAL_INFO)
	} 
	return 2, errStr, invStat
}

