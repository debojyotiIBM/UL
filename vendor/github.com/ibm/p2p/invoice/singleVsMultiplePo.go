/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pavana C K
*/

package invoice

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/ibm/p2p/po"
)

/*
This metohd is to check whether Invoice contains single po or multiple po
If invoice contains single line item, then proceed to next stage
If invoice contains multiple line items,check whether all the po's are for same company code
If it is for same company code proceed for next stage,otherwise send to IBM Ap
*/

func CheckSingleOrMultiplePo(stub shim.ChaincodeStubInterface, contextObjPtr *Context) (int, string, InvoiceStatus) {
	var invoice *Invoice = &(contextObjPtr.Invoice)
	var errStr string
	var invStat InvoiceStatus

	myLogger.Debugf("inoice ================", invoice)

	//If Invoice contains single line item
	if len(invoice.DcDocumentData.DcLines) == 1 {
		//Go for next stage
		myLogger.Debugf("Single Line", errStr)
		invStat = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", "st-vendNameAuth-1", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	} else { //If invoice contains multiple line items

		myLogger.Debugf("multi Invoice line", invoice)
		var companyCode string
		var vendorId string
		var lineCounter int = 0
		for invIdx, InvoiceLineItem := range invoice.DcDocumentData.DcLines {
			myLogger.Debugf("inoice line item ================", InvoiceLineItem)
			if invIdx == 0 {
				lineCounter++
				myLogger.Debugf("Entered if==================")
				inv_po, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, InvoiceLineItem.PoNumber, invoice.DcDocumentData.DcHeader.Client})
				if fetchErr != "" {
					invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, "PO NUMBER INVALID OR DOES NOT EXIST", "", "st-singlevsmultiple-2", EMPTY_ADDITIONAL_INFO)
					return 0, "ERROR parsing PO in singlevsMultiplePo", invStat

				}
				myLogger.Debugf("inv_po,fetchErr=========", inv_po, fetchErr)
				companyCode = inv_po.CompanyCode
				vendorId = inv_po.VendorId
				myLogger.Debugf("company Code===========", companyCode, fetchErr)
			} else {
				myLogger.Debugf("Entered else======================")
				inv_po, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, InvoiceLineItem.PoNumber, invoice.DcDocumentData.DcHeader.Client})
				if fetchErr != "" {
					invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, "PO NUMBER INVALID OR DOES NOT EXIST", "", "st-singlevsmultiple-2", EMPTY_ADDITIONAL_INFO)
					return 0, "ERROR parsing PO in singlevsMultiplePo", invStat

				}
				//Check whether all po are for same company code
				if companyCode == inv_po.CompanyCode {
					myLogger.Debugf("companyCode,vendorId=========", inv_po.VendorId, vendorId)
					if vendorId == inv_po.VendorId {
						lineCounter++
						//Go for next stage
						if lineCounter == len(invoice.DcDocumentData.DcLines) {

							invStat = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", "st-vendNameAuth-1", EMPTY_ADDITIONAL_INFO)
						} else {
							continue
						}
					} else {
						invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, "MULTIPLE PO NUMBERS NOT PERTAINING TO SAME VENDOR ID", "", "st-singlevsmultiple-2", EMPTY_ADDITIONAL_INFO)
						return 2, errStr, invStat
					}
					return 1, errStr, invStat
				} else {
					invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, "MULTIPLE PO NUMBERS NOT PERTAINING TO SAME COMPANY CODE ON INVOICE NEED AP ACTION", "", "st-singlevsmultiple-2", EMPTY_ADDITIONAL_INFO)
					return 2, errStr, invStat
					//send to IBM AP
				}

			}
		}

	}
	return 1, errStr, invStat
}

func SingleVsMuliple_Po_IBM_AP_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {
	var errStr string
	var invoice *Invoice = &(contextObjPtr.Invoice)
	if invStat.Status == FWD_TO_BUYER {
		invStat = ForwardToBuyer(stub, invoice, invStat)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-singlevsmultiple-3", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == REJECT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-singlevsmultiple-4", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == UPDATE_CC {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-vendNameAuth-1", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	} else if invStat.Status == PROCESS_OUTSIDE_BC {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, COMPLETED_OUTSIDE_BC, invStat.ReasonCode, invStat.Comments, "st-singlevsmultiple-7", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == SUBMIT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-vendNameAuth-1", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	}
	return 2, errStr, invStat
}

func SingleVsMuliple_Po_Buyer_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {
	var errStr string
	var invoice *Invoice = &(contextObjPtr.Invoice)
	if invStat.Status == FWD_TO_OTHER_BUYER {
		errStr, invStat = ForwardToOtherBuyer(stub, invoice, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-singlevsmultiple-3", EMPTY_ADDITIONAL_INFO)
		}
	} else if invStat.Status == REJECT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-singlevsmultiple-9", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == RETURN_TO_AP {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, "st-singlevsmultiple-10", EMPTY_ADDITIONAL_INFO)
	} else if invStat.ReasonCode == BUYER_DELEGATION {
		errStr, invStat = BuyerDelegation(stub, contextObjPtr, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, invStat.Status, invStat.ReasonCode, invStat.Comments, invStat.InternalStatus, EMPTY_ADDITIONAL_INFO)
		}
	} else if invStat.Status == ALT_PO {
		po, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, invoice.DcDocumentData.DcHeader.PoNumber, invoice.DcDocumentData.DcHeader.Client})
		if fetchErr != "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, "PO NUMBER INVALID OR DOES NOT EXIST", "", "st-singlevsmultiple-2", EMPTY_ADDITIONAL_INFO)
			return 2, errStr, invStat
		}
		myLogger.Debugf("entered alternate PO loop=====================", po)
		//	PoBudgetRevert(stub, invoice, po)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_WAITING_INVOICE_FIX, invStat.ReasonCode, invStat.Comments, "ST0000", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	} else if invStat.Status == UPDATE_CC {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-vendNameAuth-1", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	} else if invStat.Status == ADDITIONAL_PO {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, "st-singlevsmultiple-10", EMPTY_ADDITIONAL_INFO)
	}
	return 2, errStr, invStat
}
