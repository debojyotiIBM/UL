/*
    Copyright IBM Corp. 2017 All Rights Reserved.
    Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
    @author : Pushpalatha M Hiremath
*/

package audit

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	"github.com/op/go-logging"
)

/*
	This class is maintaining thr audit trail of invoice status. As of now the audit is attached to the invoice itself.
	In future for enhancement we use this class to maintain audit if needed.
*/


var myLogger = logging.MustGetLogger("Procure-To-Pay Audit")

func AddAuditLog(stub shim.ChaincodeStubInterface, keys []string, auditLog string) {
	db.TableStruct{Stub: stub, TableName:util.TAB_INVOICE_STATUS, PrimaryKeys: keys, Data: auditLog}.Add()
}

func GetAudit(stub shim.ChaincodeStubInterface, keys []string) (pb.Response){
	data, _ := db.TableStruct{Stub: stub, TableName:util.TAB_INVOICE_STATUS, PrimaryKeys: keys, Data: ""}.Get()
	return shim.Success(util.MarshalToBytes(data))
}

func GetAuditLog(stub shim.ChaincodeStubInterface, bciId string, invoiceNumber string) (pb.Response) {
	log, _:= db.TableStruct{Stub: stub, TableName:util.TAB_INVOICE_STATUS, PrimaryKeys:[]string{bciId, invoiceNumber}, Data: ""}.GetHistory()
	return shim.Success(util.MarshalToBytes(log))
}

func GetInvoiceDetailsLog(stub shim.ChaincodeStubInterface, bciId string, invoiceNumber string) (pb.Response) {
	log, _:= db.TableStruct{Stub: stub, TableName:util.TAB_INVOICE, PrimaryKeys:[]string{bciId, invoiceNumber}, Data: ""}.GetHistory()
	return shim.Success(util.MarshalToBytes(log))
}

