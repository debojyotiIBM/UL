/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pavana C K
*/

package invoice

import (
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	grn "github.com/ibm/p2p/grn"
	"github.com/ibm/p2p/po"
)

func ValidateCommonUnitOfMeasurement(stub shim.ChaincodeStubInterface, contextObjPtr *Context) (int, string, InvoiceStatus) {
	var invoice *Invoice = &(contextObjPtr.Invoice)
	var errStr string
	var invStat InvoiceStatus
	var lineCounter int = 0
	var grnValueMap map[float64]string
	for idx, InvoiceLineItem := range invoice.DcDocumentData.DcLines {
		var grns []grn.GRN
		grnValueMap = make(map[float64]string)
		grns = filterGrnsWithoutReferenceAndPastDatedGRNS(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, InvoiceLineItem.PoNumber, util.GetStringFromInt(InvoiceLineItem.PoLine)}, invoice)
		myLogger.Debugf("Filtered GRN's=================", grns)
		for _, grnRec := range grns {
			myLogger.Debugf("GRNS after filter======", idx, grnValueMap[grnRec.GRNValue], invoice.DcDocumentData.DcLines[idx].Amount,grnRec.GRNValue)
			if (grnRec.GRNValue == invoice.DcDocumentData.DcLines[idx].Amount) && (grnValueMap[grnRec.GRNValue] == EMPTY) {
			myLogger.Debugf("GRN condition entered=================")
				grnValueMap[grnRec.GRNValue] = grnRec.GrnNumber
				poLineItem, _ := po.GetPOLineItemPartially(stub, []string{grnRec.ErpSystem, grnRec.PONumber, util.GetStringFromInt(grnRec.POItemNumber)})
				myLogger.Debugf("GRN condition entered=================", poLineItem)
				lineCounter++

				invoice.DcDocumentData.DcLines[idx].Quantity = grnRec.Quantity
				//invoice.DcDocumentData.DcLines[idx].Description = poLineItem.Description
				invoice.DcDocumentData.DcLines[idx].UnitPrice = poLineItem.UnitPrice

			} else {
			myLogger.Debugf("GRN else condition entered=================", grnValueMap[grnRec.GRNValue])
				if grnValueMap[grnRec.GRNValue] != EMPTY {
                        myLogger.Debugf("GRN else and if condition entered=================")
					invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_NEED_MEASUREMENT_VERIFICATION, invStat.Comments, "st-commonUnitOfMeasurement-2", EMPTY_ADDITIONAL_INFO)
					return 2, errStr, invStat
				}
			}
		}

	}
	if lineCounter == len(invoice.DcDocumentData.DcLines) {
		contextObjPtr.StoreInvoice = true

	} else {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_NEED_MEASUREMENT_VERIFICATION, invStat.Comments, "st-commonUnitOfMeasurement-2", EMPTY_ADDITIONAL_INFO)
		return 2, errStr, invStat
	}

	if contextObjPtr.StoreInvoice {
		AddInvoice(stub, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, string(util.MarshalToBytes(invoice)), false)
	}
	invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-1", EMPTY_ADDITIONAL_INFO)
	return 1, errStr, invStat

}
func filterGrnsWithoutReferenceAndPastDatedGRNS(stub shim.ChaincodeStubInterface, keys []string, invoice *Invoice) []grn.GRN {
	invDate := (invoice.DcDocumentData.DcHeader.DocDate).Time()
	marginDate := invDate.AddDate(0, 0, -SIX_DAYS) // Add to constants
	recordMap := grn.FilterGRNs(stub, keys)
	var grns []grn.GRN
	for _, grnRec := range recordMap {
		deliveryDate := (grnRec.DocPostDate).Time()
		myLogger.Debugf("grnRec.RefDocNumber", grnRec.RefDocNumber, deliveryDate.After(marginDate))
		if grnRec.RefDocNumber == EMPTY && deliveryDate.After(marginDate) {
			myLogger.Debugf("Entered grn selection condition")
			grns = append(grns, grnRec)
		}
	}
	return grns
}

func getPOLineItemPartially(stub shim.ChaincodeStubInterface, keys []string) (po.POLineItem, string) {
	poRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS, PrimaryKeys: keys, Data: EMPTY}.GetAll()
	var po po.POLineItem
	for _, poRow := range poRecord {
		err := json.Unmarshal([]byte(poRow), &po)
		if err != nil {
			return po, "ERROR parsing PO"
		}
		break
	}
	return po, EMPTY
}

func CommonUnitOfMeasurement_IBM_AP_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {
	var errStr string
	var invoice *Invoice = &(contextObjPtr.Invoice)
	if invStat.Status == FWD_TO_BUYER {
		invStat = ForwardToBuyer(stub, invoice, invStat)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-commonUnitOfMeasurement-3", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == REJECT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-commonUnitOfMeasurement-4", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == PROCESS_OUTSIDE_BC {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, COMPLETED_OUTSIDE_BC, invStat.ReasonCode, invStat.Comments, "st-commonUnitOfMeasurement-5", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == "RECONSTRUCTED" {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-1", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	}
	return 2, errStr, invStat
}

func CommonUnitOfMeasurement_Buyer_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {
	var errStr string
	var invoice *Invoice = &(contextObjPtr.Invoice)
	if invStat.Status == FWD_TO_OTHER_BUYER {
		errStr, invStat = ForwardToOtherBuyer(stub, invoice, invStat)
		if errStr == EMPTY {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-commonUnitOfMeasurement-3", EMPTY_ADDITIONAL_INFO)
		}
	} else if invStat.Status == REJECT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-commonUnitOfMeasurement-4", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == RETURN_TO_AP {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, "st-commonUnitOfMeasurement-2", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == APPROVE {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-budgetValidation-1", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	} else if invStat.Status == ALT_PO {
		po, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.PoNumber, invoice.DcDocumentData.DcHeader.ErpSystem, invoice.DcDocumentData.DcHeader.Client})
		if fetchErr != "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, "PO NUMBER INVALID OR DOES NOT EXIST", "", "st-commonUnitOfMeasurement-2", EMPTY_ADDITIONAL_INFO)
			return 2, errStr, invStat
		}
		myLogger.Debugf("entered alternoate PO loop=====================", po)
		//	PoBudgetRevert(stub, invoice, po)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_WAITING_INVOICE_FIX, invStat.ReasonCode, invStat.Comments, "ST0000", EMPTY_ADDITIONAL_INFO)
	} else if invStat.ReasonCode == BUYER_DELEGATION {
		errStr, invStat = BuyerDelegation(stub, contextObjPtr, invStat)
		if errStr == EMPTY {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, invStat.Status, invStat.ReasonCode, invStat.Comments, invStat.InternalStatus, EMPTY_ADDITIONAL_INFO)
		}
	} /* else if invStat.Status == ADDITIONAL_PO {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, "st-commonUnitOfMeasurement-2", EMPTY_ADDITIONAL_INFO)
	} */
	return 2, errStr, invStat
}
