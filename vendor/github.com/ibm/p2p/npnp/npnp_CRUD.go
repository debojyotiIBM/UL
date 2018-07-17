/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package npnp

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	"github.com/op/go-logging"
)

var myLogger = logging.MustGetLogger("Procure-To-Pay")

type NPNP struct {
	ERPSystem          string `json:"erpsystem"`
	CompanyCode        string `json:"companycode"`
	VendorID           string `json:"vendorid"`
	IsInNoPOVendorList bool   `json:"isinnopovenderlist"`
	Client             string `json:"client"`
}

// Npnp

func AddNpnpRecords(stub shim.ChaincodeStubInterface, npnpRecArr string) pb.Response {
	var npnps []NPNP
	err := json.Unmarshal([]byte(npnpRecArr), &npnps)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input npnp array:", err)
	}

	for _, npnp := range npnps {
		db.TableStruct{Stub: stub, TableName: util.TAB_NPNP, PrimaryKeys: []string{npnp.ERPSystem, npnp.CompanyCode, npnp.VendorID, npnp.Client}, Data: string(util.MarshalToBytes(npnp))}.Add()
	}
	return shim.Success(nil)
}

/*
  Get company data from blockchain
*/
func GetNpnp(stub shim.ChaincodeStubInterface, erpsystem string, companycode string, vendorid string, client string) (NPNP, string) {

	ccRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_NPNP, PrimaryKeys: []string{erpsystem, companycode, vendorid, client}, Data: ""}.Get()
	var npnp NPNP

	err := json.Unmarshal([]byte(ccRecord), &npnp)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input npnp:", err, ccRecord)
		return npnp, "ERROR in parsing input npnp"
	}
	return npnp, ""
}
