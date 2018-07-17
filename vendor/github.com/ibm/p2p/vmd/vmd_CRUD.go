/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package vmd

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	//"github.com/jinzhu/copier"
	"github.com/op/go-logging"
)

type Vendor struct {
	ErpSystem     string `json:"erpsystem"`
	Client        string `json:"client"`
	VendorID      string `json:"vendorid"`
	IsDeletedFlag string `json:"isdeletedflag"`
	CPostingBlock string `json:"cpostingblock"`
	PaymentMethod string `json:"paymentmethod"`

	/*	ChangeFlag_    string `json:"changeFlag"`
		AdoptScanDate_ string `json:"adoptscandate"`
		VendorEmail_   string `json:"vendoremail"`
		VendorName_    string `json:"vendorname"`
		VendorName1_         string `json:"name1"`
		VendorName2_         string `json:"name2"`
		VendorAddress_       string `json:"address"`*/
	AdoptScanDate string `json:"adoptscandate"`
	VendorEmail   string `json:"vendoremail"`
	VendorName    string `json:"vendorname"`

	Country         string `json:"country"`
	Name1           string `json:"name1"`
	Name2           string `json:"name2"`
	Name3           string `json:"name3"`
	Name4           string `json:"name4"`
	City            string `json:"city"`
	District        string `json:"district"`
	POBox           string `json:"pobox"`
	POBoxPostalCode string `json:"poboxpostalcode"`
	PostalCode      string `json:"postalcode"`
	Region          string `json:"region"`
	Address         string `json:"address"`
	TaxNumber1      string `json:"taxnumber1"`
	TaxNumber2      string `json:"taxnumber2"`
	VATNumber       string `json:"vatnumber"`
	CustomerID      string `json:"customerid"`
	MailAddress     string `json:"mailaddress"`
}

type vendorCompanyCode struct {
	ErpSystem          string `json:"erpsystem"`
	Client             string `json:"client"`
	VendorID           string `json:"vendorid"`
	CompanyCode        string `json:"companycode"`
	Name               string `json:"name"`
	InvoicingParty     string `json:"invoicingparty"`
	IsDeletedFlag      string `json:"isdeletedflag"`
	CPostingBlock      string `json:"cpostingblock"`
	PaymentBlocked     string `json:"paymentblocked"`
	AdoptScanDateFlag  string `json:"adoptscandateflag"`
	PaymentTerms       string `json:"paymentterms"`
	WithholdingTaxCode string `json:"withholdingtaxcode"`
}

func RemoveSpecialCharAndLeadingZeros(s string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")

	processedString := reg.ReplaceAllString(s, "")

	processedString = strings.TrimLeft(processedString, "0")
	return processedString
}

type vendorBankDetails struct {
	ErpSystem         string `json:"erpsystem"`
	Client            string `json:"client"`
	VendorID          string `json:"vendorid"`
	IsDeletedFlag     string `json:"isdeletedflag"`
	BankCountryCode   string `json:"bankcountrycode"`
	BankCode          string `json:"bankcodeorig"`
	BankAccount       string `json:"bankaccountorig"`
	CurrencyReference string `json:"currencycode"`
	Iban              string `json:"iban"`
	SwiftCodeOrig     string `json:"swiftcodeorig"`
}

type vendorVATCountry struct {
	ERPSystem string `json:"erpsystem"`
	VendorId  string `json:"vendorid"`
	Country   string `json:"country"`
	Client    string `json:"client"`
	VATNumber string `json:"vatnumber"`
}

var myLogger = logging.MustGetLogger("Procure-To-Pay : VMD")

func AddVendorRecords(stub shim.ChaincodeStubInterface, vendorRecArr string) pb.Response {
	var vendors []Vendor
	err := json.Unmarshal([]byte(vendorRecArr), &vendors)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input vendors array:", err)
	}
	for _, vendor := range vendors {
		db.TableStruct{Stub: stub, TableName: util.TAB_VENDOR, PrimaryKeys: []string{vendor.ErpSystem, vendor.VendorID, vendor.Client}, Data: string(util.MarshalToBytes(vendor))}.Add()
	}
	return shim.Success(nil)
}

func AddVendorBankRecords(stub shim.ChaincodeStubInterface, vendorRecArr string) pb.Response {
	var vendorsBank []vendorBankDetails
	err := json.Unmarshal([]byte(vendorRecArr), &vendorsBank)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input vendors array:", err)
	}
	for _, vendor := range vendorsBank {
		db.TableStruct{Stub: stub, TableName: util.TAB_VENDOR_BANK, PrimaryKeys: []string{vendor.ErpSystem, vendor.VendorID, vendor.BankAccount, vendor.BankCode, vendor.BankCountryCode, vendor.Client}, Data: string(util.MarshalToBytes(vendor))}.Add()
	}
	return shim.Success(nil)
}

func AddVendorCompanyCodeRecords(stub shim.ChaincodeStubInterface, vendorRecArr string) pb.Response {
	var vendorsCC []vendorCompanyCode
	err := json.Unmarshal([]byte(vendorRecArr), &vendorsCC)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input vendors array:", err)
	}
	for _, vendor := range vendorsCC {
		db.TableStruct{Stub: stub, TableName: util.TAB_VENDOR_COMPANY_CODE, PrimaryKeys: []string{vendor.ErpSystem, vendor.VendorID, vendor.CompanyCode, vendor.Client}, Data: string(util.MarshalToBytes(vendor))}.Add()
	}
	return shim.Success(nil)
}

func VendorCompanyCodeLineMatch(e1 vendorCompanyCode, e2 vendorCompanyCode) bool {
	return e1.CompanyCode == e2.CompanyCode
}

func VendorBankDetailLineMatch(e1 vendorBankDetails, e2 vendorBankDetails) bool {
	return (e1.BankCountryCode == e2.BankCountryCode &&
		e1.BankCode == e2.BankCode &&
		e1.BankAccount == e2.BankAccount)
}

func IsBitSet(flag string, pos int) bool {

	if len(flag) < pos {
		return false
	}

	if string(flag[pos-1]) == "1" {
		return true
	} else {
		return false
	}
}

func GetVendorRec(stub shim.ChaincodeStubInterface, erpsystem string, vendorId string, client string) pb.Response {
	vendor, _ := GetVendor(stub, erpsystem, vendorId, client)
	return shim.Success(util.MarshalToBytes(vendor))
}

func GetVendor(stub shim.ChaincodeStubInterface, erpsystem string, vendorId string, client string) (Vendor, string) {
	var vmd Vendor
	vrec, fetchErr := db.TableStruct{Stub: stub, TableName: util.TAB_VENDOR, PrimaryKeys: []string{erpsystem, vendorId, client}, Data: ""}.Get()
	if fetchErr != nil {
		myLogger.Debugf("Error in fetching vendor", fetchErr)
		return vmd, "Error in fetching vendor"
	}

	err := json.Unmarshal([]byte(vrec), &vmd)
	if err != nil {
		myLogger.Debugf("ERROR parsing vendor  :", err)
		return vmd, "ERROR parsing vendor"
	}
	return vmd, ""
}

func GetVendorBankDetails(stub shim.ChaincodeStubInterface, erpsystem string, vendorId string, client string, bacc string, bcode string, bcountrycode string) (Vendor, string) {
	var vmd Vendor
	vrec, fetchErr := db.TableStruct{Stub: stub, TableName: util.TAB_VENDOR_BANK, PrimaryKeys: []string{erpsystem, vendorId, bacc, bcode, bcountrycode, client}, Data: ""}.Get()
	if fetchErr != nil {
		myLogger.Debugf("Error in fetching vendor", fetchErr)
		return vmd, "Error in fetching vendor"
	}

	err := json.Unmarshal([]byte(vrec), &vmd)
	if err != nil {
		myLogger.Debugf("ERROR parsing vendor  :", err)
		return vmd, "ERROR parsing vendor"
	}
	return vmd, ""
}

func GetVendorCompanyCode(stub shim.ChaincodeStubInterface, erpsystem string, vendorId string, client string, companyCode string) (Vendor, string) {
	var vmd Vendor
	vrec, fetchErr := db.TableStruct{Stub: stub, TableName: util.TAB_VENDOR_COMPANY_CODE, PrimaryKeys: []string{erpsystem, vendorId, companyCode, client}, Data: ""}.Get()
	if fetchErr != nil {
		myLogger.Debugf("Error in fetching vendor", fetchErr)
		return vmd, "Error in fetching vendor"
	}

	err := json.Unmarshal([]byte(vrec), &vmd)
	if err != nil {
		myLogger.Debugf("ERROR parsing vendor  :", err)
		return vmd, "ERROR parsing vendor"
	}
	return vmd, ""
}

func GetAllVendors(stub shim.ChaincodeStubInterface) pb.Response {
	var vendors []Vendor
	var vmd Vendor
	vendorRecords, _ := db.TableStruct{Stub: stub, TableName: util.TAB_VENDOR, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, vendorRec := range vendorRecords {
		err := json.Unmarshal([]byte(vendorRec), &vmd)
		if err != nil {
			myLogger.Debugf("ERROR parsing Vendors", vendorRec, err)
			// return vendors, "ERROR parsing vendor"
			return shim.Error("ERROR parsing Vendors")
		}
		vendors = append(vendors, vmd)
	}
	// return vendors
	return shim.Success(util.MarshalToBytes(vendors))
}

func GetAllVendorsBank(stub shim.ChaincodeStubInterface) pb.Response {
	var vendors []vendorBankDetails
	var vmd vendorBankDetails
	vendorRecords, _ := db.TableStruct{Stub: stub, TableName: util.TAB_VENDOR_BANK, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, vendorRec := range vendorRecords {
		err := json.Unmarshal([]byte(vendorRec), &vmd)
		if err != nil {
			myLogger.Debugf("ERROR parsing Vendors", vendorRec, err)
			// return vendors, "ERROR parsing vendor"
			return shim.Error("ERROR parsing Vendors")
		}
		vendors = append(vendors, vmd)
	}
	// return vendors
	return shim.Success(util.MarshalToBytes(vendors))
}

func GetAllVendorsCompanyCode(stub shim.ChaincodeStubInterface) pb.Response {
	var vendors []vendorCompanyCode
	var vmd vendorCompanyCode
	vendorRecords, _ := db.TableStruct{Stub: stub, TableName: util.TAB_VENDOR_COMPANY_CODE, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, vendorRec := range vendorRecords {
		err := json.Unmarshal([]byte(vendorRec), &vmd)
		if err != nil {
			myLogger.Debugf("ERROR parsing Vendors", vendorRec, err)
			// return vendors, "ERROR parsing vendor"
			return shim.Error("ERROR parsing Vendors")
		}
		vendors = append(vendors, vmd)
	}
	// return vendors
	return shim.Success(util.MarshalToBytes(vendors))
}

/* ===== Vendor Status Check Functions ===== */

func (vendor *Vendor) IsPaymentBlocked() bool {
	return vendor.CPostingBlock == "X"
}

func (vendor *Vendor) IsDeleted() bool {
	return vendor.IsDeletedFlag == "X"
}
