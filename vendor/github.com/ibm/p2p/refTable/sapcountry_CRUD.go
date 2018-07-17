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
	//"github.com/op/go-logging"
)

type SAPCountry struct {
	ERPSystem       string `json:"erpsystem"`
	Client          string `json:"client"`
	CountryKey      string `json:"countrykey"`
	LanguageKey     string `json:"languagekey"`
	CountryISOCode  string `json:"countryisocode"`
	IsFromEU        string `json:"isfromeu"`
	Procedure       string `json:"procedure"`
	CountryCurrency string `json:"countrycurrency"`
}

func AddSAPCountryRecords(stub shim.ChaincodeStubInterface, sapcountryRecArr string) pb.Response {
	var sapcountrys []SAPCountry
	err := json.Unmarshal([]byte(sapcountryRecArr), &sapcountrys)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input sapcountry array:", err)
	}

	for _, sapcountry := range sapcountrys {
		db.TableStruct{Stub: stub, TableName: util.TAB_SAPCOUNTRY, PrimaryKeys: []string{sapcountry.ERPSystem, sapcountry.Client, sapcountry.CountryKey}, Data: string(util.MarshalToBytes(sapcountry))}.Add()
	}
	return shim.Success(nil)
}

/*
  Get company data from blockchain
*/
func GetSAPCountry(stub shim.ChaincodeStubInterface, erpsystem string, client string, countrykey string) (SAPCountry, string) {

	ccRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_SAPCOUNTRY, PrimaryKeys: []string{erpsystem, client, countrykey}, Data: ""}.Get()
	var sapcountry SAPCountry

	err := json.Unmarshal([]byte(ccRecord), &sapcountry)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input sapcountry:", err, ccRecord)
		return sapcountry, "ERROR in parsing input sapcountry"
	}
	return sapcountry, ""
}

// GetALL method to get all data
func GetAllSAPCountry(stub shim.ChaincodeStubInterface) []SAPCountry {
	var allSAPCountrys []SAPCountry
	SAPCountrysRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_SAPCOUNTRY, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, grnRow := range SAPCountrysRec {
		var currentSAPCountry SAPCountry
		json.Unmarshal([]byte(grnRow), &currentSAPCountry)
		allSAPCountrys = append(allSAPCountrys, currentSAPCountry)
	}

	return allSAPCountrys
}
