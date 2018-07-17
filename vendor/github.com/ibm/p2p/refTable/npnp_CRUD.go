/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package refTable

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	"github.com/op/go-logging"
	"errors"
)

var myLogger = logging.MustGetLogger("Procure-To-Pay")

type NPNP struct {
	ERPSystem          string `json:"erpsystem"`
	CompanyCode        string `json:"companycode"`
	VendorID           string `json:"vendorid"`
	IsInNoPOVendorList bool   `json:"isinnopovendorlist"`
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
func GetNpnp(stub shim.ChaincodeStubInterface, erpsystem string, companycode string, vendorid string, client string) (NPNP, error) {

	var npnp NPNP
	ccRecord, err := db.TableStruct{Stub: stub, TableName: util.TAB_NPNP, PrimaryKeys: []string{erpsystem, companycode, vendorid, client}, Data: ""}.GetAll()
	if(err != nil) {
		return npnp, err
	}
	var isRecordExists bool
	for _, ccRow := range ccRecord {
		json.Unmarshal([]byte(ccRow), &npnp)
		isRecordExists = true
		break;
	}
	if (!isRecordExists) {
		return npnp, errors.New("Record does not exists")
	}
	return npnp, nil
}

// GetALL method to get all data
func GetAllNpnp(stub shim.ChaincodeStubInterface) []NPNP {
	var allNPNPs []NPNP
	NPNPsRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_NPNP, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, grnRow := range NPNPsRec {
		var currentNPNP NPNP
		json.Unmarshal([]byte(grnRow), &currentNPNP)
		allNPNPs = append(allNPNPs, currentNPNP)
	}

	return allNPNPs
}
