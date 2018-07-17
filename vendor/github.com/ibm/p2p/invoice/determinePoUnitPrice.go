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
	logging "github.com/op/go-logging"
)

func DeterminePOUnitPrice(stub shim.ChaincodeStubInterface, contextObjPtr *Context) (int, string, InvoiceStatus) {
	var myLogger = logging.MustGetLogger("DeterminePOUnitPrice")
	var invoice *Invoice = &(contextObjPtr.Invoice)
	var invStat InvoiceStatus
	var errStr string
	cacheMapForPOLine := make(map[string] []po.POLineItem) // For Caching..
//	var toComparePOLine po.POLineItem // Dummy field......
	var poLineFromDB po.POLineItem
//	var err string
	myLogger.Debugf("********************************* DeterminePOUnitPrice *********************************************************")
	for _, InvoiceLineItem := range invoice.DcDocumentData.DcLines {
		key := invoice.DcDocumentData.DcHeader.ErpSystem + "-" + InvoiceLineItem.PoNumber 

		myLogger.Debugf("Inside For Invoice BCIID = ", invoice.BCIID, " key", key)
		myLogger.Debugf("Length  = ", len((cacheMapForPOLine[key])))
		
		if (len((cacheMapForPOLine[key])) == 0 ) {
			var allPOLineItems []po.POLineItem
			allPOLineItems = po.GetAllPOLineItemsByPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, InvoiceLineItem.PoNumber})
			
			if len(allPOLineItems) == 0 {
				myLogger.Debugf("PO Line not present in the ref Table : key = ", key)
				invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, "PO Line Does not Exists in DB", invStat.Comments, "st-commonUnitOfMeasurement-2", EMPTY_ADDITIONAL_INFO)
				return 2, errStr, invStat
			}
			cacheMapForPOLine[key] = allPOLineItems
			for _, poLineFromDB = range allPOLineItems {
				// Compute Unit Price. Update POLine is condition is satisfied -- Starts
				netOrderValue := poLineFromDB.NetOrderValue
				myLogger.Debugf("netOrderValue : ", netOrderValue)
				if netOrderValue != 1 {
					noOfUnits := poLineFromDB.Per
					poLineUnitPrice := netOrderValue / noOfUnits
					poLineFromDB.UnitPrice = poLineUnitPrice
				} else {
					poLineFromDB.UnitPrice = netOrderValue
				}
				myLogger.Debugf("UnitPrice :::: ", poLineFromDB.UnitPrice, poLineFromDB.PONumber, poLineFromDB.ERPSystem, poLineFromDB.Client, poLineFromDB.LineItemNumber)
				po.AddPOLineItems(stub, poLineFromDB.PONumber, poLineFromDB.ERPSystem, poLineFromDB.Client, poLineFromDB.LineItemNumber, string(util.MarshalToBytes(poLineFromDB)))
			myLogger.Debugf("Successfully added :::::::::::::::::::::::::::")
				// Compute Unit Price -- Ends
			} // End of FOR loop
		} else {
			myLogger.Debugf("PO Line already computed : key = ", key)
		} // End of For
	}

	invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", invStat.Comments, "st-commonUnitOfMeasurement-1", EMPTY_ADDITIONAL_INFO)
	return 1, errStr, invStat
}
