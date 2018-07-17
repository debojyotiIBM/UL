/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pavana C K
*/

package invoice

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/ibm/p2p/po"
	util "github.com/ibm/p2p"
)

//Pseudo Code

/* fetch invoiceCurrency from invoice
set invoiceCurrency to fetched invoiceCurrency from invoice header
if invoiceCurrency is empty
fail to IBM AP

set linecounter to 0
for each line in invoice
fetch ponumber
set poNumber to fetched ponumber from invoice line
fetch poCurrency
if poCurrency is equal to invoiceCurrency
	increment counter
	if linecounter is equal to invoiceline length
		go for next stage
	else
		skip the line
else
go for next stage // need to change later once dynamic table details are provided(2 dynamic table check and Should fail to IBM AP) */

func ValidateCurrency(stub shim.ChaincodeStubInterface, contextObjPtr *Context) (int, string, InvoiceStatus) {
	var invoice *Invoice = &(contextObjPtr.Invoice)
	var errStr string
	var invStat InvoiceStatus
	var poCurrency string
	var invoiceCurrency string = EMPTY
	invoiceCurrency = invoice.DcDocumentData.DcHeader.CurrencyCode
	if invoiceCurrency != EMPTY {
		var lineCounter int = 0

		for i, InvoiceLineItem := range invoice.DcDocumentData.DcLines {
			inv_po, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, InvoiceLineItem.PoNumber, invoice.DcDocumentData.DcHeader.Client})
			if fetchErr != "" {
				invStat = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", "st-lineAggregation-1", EMPTY_ADDITIONAL_INFO)
				/* invStat, errStr = SetInvoiceStatus(stub, contextObjPtr,invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_INVALID_PO, "", "st-currencyValidation-2", EMPTY_ADDITIONAL_INFO)
				return 2, "ERROR parsing PO in singlevsMultiplePo", invStat */

			}
			poCurrency = inv_po.Currency
			if invoiceCurrency == poCurrency {
				lineCounter++
				if lineCounter == len(invoice.DcDocumentData.DcLines) {
					//Go to next stage
					invStat = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", "st-lineAggregation-1", EMPTY_ADDITIONAL_INFO)
				} else {
					continue
				}
			} else {
				myLogger.Debugf("Entered else part")
				invoice.DcDocumentData.DcLines[i].D_currencyMissMatch = true
				contextObjPtr.StoreInvoice = true
				invStat = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", "st-lineAggregation-1", EMPTY_ADDITIONAL_INFO)
			}
		}
	} else {
		/* invStat, errStr = SetInvoiceStatus(stub,contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_COMPANY_CODE_MISSMATCH, "", "st-currencyValidation-3", EMPTY_ADDITIONAL_INFO)
		return 2, errStr, invStat */
		//fail to IBM AP
		invStat = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", "st-lineAggregation-1", EMPTY_ADDITIONAL_INFO)
	}
	if contextObjPtr.StoreInvoice {
		myLogger.Debugf("Entered updating part")
		invoice.DcDocumentData.DcHeader.D_currencyMissMatch = true
		AddInvoice(stub, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, string(util.MarshalToBytes(invoice)), false)		
	}
	return 1, errStr, invStat
}

func CurrencyValidation_IBM_AP_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invoice Invoice, invStat InvoiceStatus) (int, string, InvoiceStatus) {
	var errStr string
	if invStat.Status == FWD_TO_BUYER {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-currencyValidation-4", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == REJECT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-currencyValidation-5", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == PROCESS_OUTSIDE_BC {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, COMPLETED_OUTSIDE_BC, invStat.ReasonCode, invStat.Comments, "st-currencyValidation-6", EMPTY_ADDITIONAL_INFO)
	}
	return 2, errStr, invStat
}

func CurrencyValidation_Buyer_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {
	var invoice *Invoice = &(contextObjPtr.Invoice)
	var errStr string
	if invStat.Status == FWD_TO_OTHER_BUYER {
		errStr, invStat = ForwardToOtherBuyer(stub, invoice, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-currencyValidation-4", EMPTY_ADDITIONAL_INFO)
		}
	} else if invStat.Status == REJECT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-currencyValidation-7", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == RETURN_TO_AP {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, "st-currencyValidation-2", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == APPROVE {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-lineAggregation-1", EMPTY_ADDITIONAL_INFO)
	} else if invStat.ReasonCode == BUYER_DELEGATION {
		errStr, invStat = BuyerDelegation(stub, contextObjPtr, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, invStat.Status, invStat.ReasonCode, invStat.Comments, invStat.InternalStatus, EMPTY_ADDITIONAL_INFO)
		}
	}
	return 2, errStr, invStat
}
