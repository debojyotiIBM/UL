/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Sharath Chandra R
*/

package invoice

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	util "github.com/ibm/p2p"
	po "github.com/ibm/p2p/po"
)

func PostfactoPo(stub shim.ChaincodeStubInterface, contextObjPtr *Context) (int, string, InvoiceStatus) {
	var invoice *Invoice = &(contextObjPtr.Invoice)
	var invStat InvoiceStatus
	var errStr string
	cacheMapForPO := make(map[string]po.PO) // For Caching..
	var poFromDB po.PO
	var err string
	myLogger.Debugf("********************************* PostfactoPo *********************************************************", invoice)
	for i, InvoiceLineItem := range invoice.DcDocumentData.DcLines {
		myLogger.Debugf("Inside For Invoice BCIID = ", invoice.BCIID)
		key := invoice.DcDocumentData.DcHeader.ErpSystem + "-" + InvoiceLineItem.PoNumber + "-" + invoice.DcDocumentData.DcHeader.Client
		if len((cacheMapForPO[key]).ERPSystem) == 0 {
			poFromDB, err = po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, InvoiceLineItem.PoNumber, invoice.DcDocumentData.DcHeader.Client})
			if err != "" {
				myLogger.Debugf("PO not present in the ref Table : composite key = ", key)
				invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, "ST0000", invStat.Comments, "st-determinePOUnitPrice-2", EMPTY_ADDITIONAL_INFO)
				return 1, errStr, invStat
			}
			cacheMapForPO[key] = poFromDB
		} else {
			myLogger.Debugf(" Fetching from cache for key ", key)
			poFromDB = cacheMapForPO[key]
		}
		creationDate := (poFromDB.CreationDate).Time()
		myLogger.Debugf("Po creation date in PostfactoPo ======== ", creationDate)
		invDate := (invoice.DcDocumentData.DcHeader.DocDate).Time()
		myLogger.Debugf("Invoice date  in PostfactoPo ======== ", invDate)

		if creationDate.After(invDate) {
			myLogger.Debugf(" Its a Post facto PO ================")
			contextObjPtr.StoreInvoice = true
			invoice.DcDocumentData.DcLines[i].D_postFacto = true
		}
	} // End of For

	myLogger.Debugf(" postFactoFlag ================ contextObjPtr.StoreInvoice = ", contextObjPtr.StoreInvoice)
	if contextObjPtr.StoreInvoice {
		invoice.DcDocumentData.DcHeader.D_postFacto = true
		AddInvoice(stub, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, string(util.MarshalToBytes(invoice)), false)
		//		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, "CONTINUE", "", invStat.Comments, "st-postFactoPO-1", EMPTY_ADDITIONAL_INFO)
		//		return 1, errStr, invStat
	}

	invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", invStat.Comments, "st-currencyValidation-1", EMPTY_ADDITIONAL_INFO)
	return 1, errStr, invStat
}
