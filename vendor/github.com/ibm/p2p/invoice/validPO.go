/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package invoice

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/ibm/p2p/po"
	npnp "github.com/ibm/p2p/refTable"
)

func PODBIncludeAllPOStatus(stub shim.ChaincodeStubInterface, contextObjPtr *Context, dcLine DCLine) (int, string, InvoiceStatus, string) {

	myLogger.Debugf("Inside PODBIncludeAllPOStatus Starts ")
	var invoice *Invoice = &(contextObjPtr.Invoice)
	var invStat InvoiceStatus
	var errStr string
	// Dynamic Table : PO DB include all the PO with PO status
	var poWithPoStatus string = "N"
	if poWithPoStatus == "N" {
		// 2.2.24

		_, err := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, dcLine.PoNumber, invoice.DcDocumentData.DcHeader.Client})
		if err != "" { // PO not present in PO database - GOTO IBM AP UI
			myLogger.Debugf("PO is not present in the DB ")
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, "PO NUMBER INVALID OR DOES NOT EXIST", invStat.Comments, "st-ValidPO-2", EMPTY_ADDITIONAL_INFO)
			// Dynamic Table for RoutingToUI :
			//			var routingToUI string = "UL"
			//			if routingToUI == "UL" {
			//				//IBM AP UI
			//			} else {
			// Other UI
			//			}
		} else {
			myLogger.Debugf("PO is present in the DB. Go to Next Step .........................................")
			// PO present in PO database :: go to Next Step
			//			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, "CONTINUE", "", invStat.Comments, "st-companyCode-1", EMPTY_ADDITIONAL_INFO)
			myLogger.Debugf("invStat ........................... ", invStat)
			return 1, errStr, invStat, ""
		}
	} else { // poWithPoStatus = "Y"
		/*	invoicePO, err := po.GetPO(stub, []string{invoice.ErpSystem(), dcLine.PoNumber()})
			if err == "" { // PO present in PO database
				var notActiveFlag = false
				for _, poLineItem := range *invoicePO.LineItems() {
					if "ACTIVE" != strings.ToUpper(poLineItem.PoStatus()) {
						notActiveFlag = true
						break
					}
				}
				if notActiveFlag {
					// TODO :: Go to Buyer
				} else {
					// Next Step : Do nothing..
				}
			} else {
				// PO Not present
				// Dynamic Table : Auto Rejection Permitted
				var autoRejectionPermitted string = "Y"
				if autoRejectionPermitted == "Y" {
					vendor, err := vmd.GetVendor(stub, invoice.ErpSystem(), invoice.VendorID())
					if err == "" {
						// Check NPNP List
						var presentInNPNPList string = "Y"
						if presentInNPNPList == "Y" {
							// Go to NON PO Process
						} else {
							// Reject. Businees shld give clarity
						}
					} else {
						// Vendor not present
						// Dynamic table enabling to route invoice to different UI level
					}
				}

			} */
	}
	myLogger.Debugf("Inside PODBIncludeAllPOStatus Ends ....... ")
	return 2, errStr, invStat, ""
}

func ValidPO(stub shim.ChaincodeStubInterface, contextObjPtr *Context) (int, string, InvoiceStatus, string) {
	var invoice *Invoice = &(contextObjPtr.Invoice)
	var invStat InvoiceStatus
	var errStr string
	var lineCounter int = 0
	//	var paymentMethod string
	myLogger.Debugf("********************************* VALID PO *********************************************************", invoice)
	for _, InvoiceLineItem := range invoice.DcDocumentData.DcLines {
		// Call Dynamic Table Mandated Valid PO @ Platform Submission
		myLogger.Debugf("Inside For Invoice BCIID = ", invoice.BCIID)
		var mandatedValidPOAtPlatform_DT string = "N"
		if mandatedValidPOAtPlatform_DT == "N" {
			// Call dynamic Table InvoiceSubmittedWithoutPO
			var InvoiceSubmittedWithoutPO_DT string = "Y"
			if InvoiceSubmittedWithoutPO_DT == "Y" {
				// Check PO provided on Invoice :  If Present GO to Link 5 :  ===> 2.3.27
				if InvoiceLineItem.PoNumber != "" && len(InvoiceLineItem.PoNumber) > 0 {
					myLogger.Debugf("PONumber present on Invoice  PoNumber= ", InvoiceLineItem.PoNumber)
					returnCode, errStr, invStat, _ := PODBIncludeAllPOStatus(stub, contextObjPtr, InvoiceLineItem)
					if returnCode == 1 {
						lineCounter++
					} else {
						return returnCode, errStr, invStat, ""
					}
				} else {
					myLogger.Debugf("PONumber not present on Invoice.....  PoNumber= ", InvoiceLineItem.PoNumber)
					// Dynamic Table : Auto Rejection Permitted at Company code level
					var autoRejectionPermitted_DT string = "N"
					if autoRejectionPermitted_DT == "N" { // If YES
						// GO to IBM AP
						invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, "NON COMPLIANCE TO NO PO NO PAY", invStat.Comments, "st-ValidPO-1", EMPTY_ADDITIONAL_INFO)
						return 2, errStr, invStat, ""
					} else {
						/*vendor, err := vmd.GetVendor(stub, invoice.ErpSystem(), invoice.VendorID())
						// Vendor Present
						if err == "" {
							// TODO : NPNP List

						} else {
							// Vendor does not Exists. Go to IBM AP UI

						}*/
					}
				}
			} else {
				// if InvoiceSubmittedWithoutPO_DT = "N"
				//PODBIncludeAllPOStatus(stub, invoice, InvoiceLineItem)
			}
		} else {
			//If YES --> GO to Next Sep
		}

	} // End of For

	if lineCounter == len(invoice.DcDocumentData.DcLines) {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, "CONTINUE", "", invStat.Comments, "st-companyCode-1", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat, ""
	}
	return 2, errStr, invStat, ""
}


func ValidateInNPNPList(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {
	myLogger.Debugf("*************************  ValidateInNPNPList **************************************")
	var invoice *Invoice = &(contextObjPtr.Invoice)
	npnp, err := npnp.GetNpnp(stub, invoice.DcDocumentData.DcHeader.ErpSystem, invoice.DcDocumentData.DcHeader.CompanyCode, invoice.DcDocumentData.DcHeader.VendorID, invoice.DcDocumentData.DcHeader.Client)
	myLogger.Debugf("npnp", npnp, "err", err)
	if err != nil {
		myLogger.Debugf("Its not in NPNP List") // The invoice stays in NPNP DB refresh state.
		return 2, "", invStat
	} else {
		if(npnp.IsInNoPOVendorList) {
			invStat.Status = SUBMIT_AS_NON_PO
			return ValidPO_IBMAP_Action(stub, contextObjPtr, invStat)
		}
		return 2, "", invStat
	}
}

func ValidPO_IBMAP_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {
	var invoice *Invoice = &(contextObjPtr.Invoice)
	myLogger.Debugf("IBM AP ACTION - ", invStat.Status)
	var errStr string
	if invStat.Status == FWD_TO_BUYER {
		myLogger.Debugf("entered f/w to buyer")
		invStat = ForwardToBuyer(stub, invoice, invStat)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-ValidPO-3", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == INV_STATUS_REJECTED {
		//		invStat, errStr = SetInvoiceStatus(stub, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-ValidPO-5", EMPTY_ADDITIONAL_INFO)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-ValidPO-5", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == SUBMIT_AS_NON_PO {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, "COMPLETED", SUBMIT_AS_NON_PO, invStat.Comments, "st-ValidPO-6", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == PROCESS_OUTSIDE_BC {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, COMPLETED_OUTSIDE_BC, PROCESS_OUTSIDE_BC, invStat.Comments, "st-ValidPO-7", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == WAITING_ADDITION_TO_NPNP_EXCEPTION_LIST {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, NPNP_DB_REFRESH, AWAITING_PROCUREMENT_CONFIRMATION, invStat.Comments, "st-ValidPO-15", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == NPNP_DB_REFRESH {
		myLogger.Debugf("*************************  NPNP_DB_REFRESH **************************************")
		_, errStr, invStat = ValidateInNPNPList(stub, contextObjPtr, invStat)
	} else if invStat.Status == "RECONSTRUCTED" {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-companyCode-1", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	}
	return 2, errStr, invStat
}

func ValidPO_BUYER_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {
	var invoice *Invoice = &(contextObjPtr.Invoice)
	myLogger.Debugf("BUYER ACTION - ", invStat.Status)
	var errStr string
	if invStat.Status == FWD_TO_OTHER_BUYER {
		errStr, invStat = ForwardToOtherBuyer(stub, invoice, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-ValidPO-3", EMPTY_ADDITIONAL_INFO)
		}
	} else if invStat.Status == BUYER_DELEGATION {
		errStr, invStat = BuyerDelegation(stub, contextObjPtr, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, invStat.Status, invStat.ReasonCode, invStat.Comments, "st-ValidPO-9", EMPTY_ADDITIONAL_INFO) //TODO
		}
	} else if invStat.Status == RETURN_TO_AP {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, "st-ValidPO-10", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == INV_STATUS_REJECTED {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-ValidPO-11", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == UNBLOCK_PO {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_WAITING_PO_REFRESH, invStat.ReasonCode, invStat.Comments, "ST0000", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == UNMARK_FOR_DELETION {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, UNMARK_FOR_DELETION, invStat.ReasonCode, invStat.Comments, "st-ValidPO-13", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == ALT_PO {
		myLogger.Debugf("entered alternate PO loop=====================")
		_, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, invoice.DcDocumentData.DcHeader.PoNumber, invoice.DcDocumentData.DcHeader.Client})
		if fetchErr != "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_INVALID_PO, "", "st-ValidPO-1", EMPTY_ADDITIONAL_INFO)
			return 2, errStr, invStat
		}
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_WAITING_INVOICE_FIX, invStat.ReasonCode, invStat.Comments, "ST0000", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	} else if invStat.Status == INV_STATUS_REJECTED {
		//		invStat, errStr = SetInvoiceStatus(stub, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-ValidPO-5", EMPTY_ADDITIONAL_INFO)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-ValidPO-14", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == ADDITIONAL_PO {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, "st-ValidPO-10", EMPTY_ADDITIONAL_INFO)
	}
	return 2, errStr, invStat
}
