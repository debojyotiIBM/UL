/*
   Copyright IBM Corp. 2018 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Lohit Krishnan
*/

package invoice

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func PerformDCGR(stub shim.ChaincodeStubInterface, contextObjPtr *Context) (int, string, InvoiceStatus) {
	invStat, errStr := SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_COMPLETED, TEMP_DCGR_STATUS, "", "st-DCGR-end", EMPTY_ADDITIONAL_INFO)
	return 2, errStr, invStat
}
