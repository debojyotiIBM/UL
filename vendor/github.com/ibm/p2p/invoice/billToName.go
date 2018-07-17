/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pavana C K
*/

package invoice

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	companyCode "github.com/ibm/p2p/companyCode"
	"github.com/ibm/p2p/po"
)

/*
This metohd is to select appropriate company code for the invoice
If invoice contains single line item, then select the company code for that particular po and update company code in the invoice, proceed for next stage
If invoice contains multiple line items,check whether all the po's are for same company code. If yes , Select the company code for
that particular PO, update company code in the invoice and proceed for next stage,otherwise send to IBM AP
*/

func BilltoNameMatch(stub shim.ChaincodeStubInterface, contextObjPtr *Context) (int, string, InvoiceStatus) {
	var invoice *Invoice = &(contextObjPtr.Invoice)
	var errStr string
	var invStat InvoiceStatus
	myLogger.Debugf("inoice ================", invoice)

	//If Invoice contains single line item
	if len(invoice.DcDocumentData.DcLines) == 1 {
		//Get corresponding po for that line
		myLogger.Debugf("Single Line, So select company code for invoice", errStr)
		inv_po, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, invoice.DcDocumentData.DcLines[0].PoNumber, invoice.DcDocumentData.DcHeader.Client})
		myLogger.Debugf("inv_po,fetchErr=========", inv_po, fetchErr)

		if fetchErr != "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, "PO NUMBER INVALID OR DOES NOT EXIST", "", "st-billToName-2", EMPTY_ADDITIONAL_INFO)
			return 0, "ERROR parsing PO in billToName", invStat
			// return 0, fetchErr, invStat
		}

		//Override inovoice company code with po's company code
		//	invoice.SetCompanyCode(inv_po.CompanyCode)
		myLogger.Debugf("derived_CompanyCode===========", inv_po.CompanyCode)
		cc, ccFetchErr := companyCode.GetCompanyCode(stub, invoice.DcDocumentData.DcHeader.ErpSystem, inv_po.CompanyCode)
		if ccFetchErr != "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, "INVALID COMPANY CODE", invStat.Comments, "st-billToName-2", EMPTY_ADDITIONAL_INFO)
			return 2, errStr, invStat
		}
		if cc.Name == invoice.DcDocumentData.DcSwissHeader.BuyerName {
			invStat = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", "st-bciValidation-1", EMPTY_ADDITIONAL_INFO)
			return 1, errStr, invStat
		} else {
			//reject
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, BILL_TO_NAME_MISSMATCH, invStat.Comments, "st-billToName-2", EMPTY_ADDITIONAL_INFO)
			return 2, errStr, invStat
		}

	} else { //If invoice contains multiple line items

		myLogger.Debugf("Comapny code selection multi Invoice line", invoice)
		var poCompanyCode string
		var lineCounter int = 0
		for invIdx, InvoiceLineItem := range invoice.DcDocumentData.DcLines {

			myLogger.Debugf("inoice line item ================", InvoiceLineItem)
			if invIdx == 0 {
				lineCounter++
				myLogger.Debugf("Entered if==================", lineCounter)
				//Get corresponding po for that line
				inv_po, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, InvoiceLineItem.PoNumber, invoice.DcDocumentData.DcHeader.Client})

				if fetchErr != "" {
					invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, "PO NUMBER INVALID OR DOES NOT EXIST", "", "st-billToName-2", EMPTY_ADDITIONAL_INFO)
					return 0, "ERROR parsing PO in billToName", invStat
					// return 0, fetchErr, invStat
				}
				myLogger.Debugf("inv_po,fetchErr=========", inv_po, fetchErr)
				//Get comapnycode to compare with rest of the line's
				poCompanyCode = inv_po.CompanyCode

			} else {
				myLogger.Debugf("Entered else ======================")
				//Get corresponding po for that line
				inv_po, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, InvoiceLineItem.PoNumber, invoice.DcDocumentData.DcHeader.Client})

				if fetchErr != "" {
					invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, "PO NUMBER INVALID OR DOES NOT EXIST", "", "st-billToName-2", EMPTY_ADDITIONAL_INFO)
					return 2, errStr, invStat

				}
				myLogger.Debugf("inv_po,fetchErr=========", inv_po, fetchErr, inv_po.CompanyCode)
				//Check whether all po are for same company code
				if poCompanyCode == inv_po.CompanyCode {
					lineCounter++
					//To check whether company code comaprison is done with all the lines
					myLogger.Debugf("Linecounter and dclength equal?=============", lineCounter, len(invoice.DcDocumentData.DcLines))
					if lineCounter == len(invoice.DcDocumentData.DcLines) {
						//Override inovoice company code with po's company code
						cc, ccFetchErr := companyCode.GetCompanyCode(stub, invoice.DcDocumentData.DcHeader.ErpSystem, inv_po.CompanyCode)
						if ccFetchErr != "" {
							invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, "COMPANY CODE DOES NOT EXIST", invStat.Comments, "st-billToName-2", EMPTY_ADDITIONAL_INFO)
							return 2, errStr, invStat
						}
						if cc.Name == invoice.DcDocumentData.DcSwissHeader.BuyerName {
							myLogger.Debugf("Buyer Name============", invoice.DcDocumentData.DcSwissHeader.BuyerName)
							invStat = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", "st-bciValidation-1", EMPTY_ADDITIONAL_INFO)
						} else {
							invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, BILL_TO_NAME_MISSMATCH, invStat.Comments, "st-billToName-2", EMPTY_ADDITIONAL_INFO)
							return 2, errStr, invStat
							//reject
						}

					} else {
						continue
					}
					return 1, errStr, invStat
				} else {
					myLogger.Debugf("entered else===========")
					invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_COMPANY_CODE_MISSMATCH, "", "st-billToName-2", EMPTY_ADDITIONAL_INFO)
					return 2, errStr, invStat
					//send to IBM AP
				}

			}
		}
	}

	return 1, errStr, invStat
}

func BilltoNameMatch_IBM_AP_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {
	var errStr string
	var invoice *Invoice = &(contextObjPtr.Invoice)
	if invStat.Status == FWD_TO_BUYER {
		invStat = ForwardToBuyer(stub, invoice, invStat)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-billToName-10", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == REJECT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-billToName-2", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == PROCESS_OUTSIDE_BC {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, COMPLETED_OUTSIDE_BC, invStat.ReasonCode, invStat.Comments, "st-billToName-7", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == "RECONSTRUCTED" {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-bciValidation-1", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	} else if invStat.Status == SUBMIT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-bciValidation-1", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	}
	return 2, errStr, invStat
}

func BilltoName_Buyer_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {
	var errStr string
	var invoice *Invoice = &(contextObjPtr.Invoice)
	if invStat.Status == FWD_TO_OTHER_BUYER {
		errStr, invStat = ForwardToOtherBuyer(stub, invoice, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-billToName-10", EMPTY_ADDITIONAL_INFO)
		}
	} else if invStat.Status == REJECT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-billToName-5", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == RETURN_TO_AP {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, "st-billToName-6", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == APPROVE {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-bciValidation-1", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	} else if invStat.ReasonCode == BUYER_DELEGATION {
		errStr, invStat = BuyerDelegation(stub, contextObjPtr, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, invStat.Status, invStat.ReasonCode, invStat.Comments, invStat.InternalStatus, EMPTY_ADDITIONAL_INFO)
		}
	} else if invStat.Status == ADDITIONAL_PO {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, "st-billToName-6", EMPTY_ADDITIONAL_INFO)
	}
	return 2, errStr, invStat
}
