/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : NK
*/

package report

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	util "github.com/ibm/p2p"
	"github.com/ibm/p2p/invoice"
	"github.com/ibm/p2p/vmd"
	logging "github.com/op/go-logging" 
	"time"
)

type WhoApprovedWhat struct {
	ScanID          string      `json:"scanID"`
	DocNumber       string      `json:"docNumber"`
	VendorName      string      `json:"vendorName"`
	VendorID        string      `json:"vendorID"`
	Total           float64     `json:"total"`
	Currency        string      `json:"currency"`
	DocDate         util.BCDate `json:"docDate"`
	ApprovedBy      string      `json:"approvedBy"`
	ApproverAccount string      `json:"approverAccount"`
	ERPPosting      string      `json:"eRPPosting"`
}

var myLogger = logging.MustGetLogger("Procure-To-Pay CompanyCode")

const bcdLayout = "20060102"

func GetReport(stub shim.ChaincodeStubInterface, status string, startDate string, endDate string) pb.Response {


	var detailedInvoiceArr []invoice.DetailedInvoice
	detailedInvoiceArr = invoice.GetInvoiceByStatus(stub, status)

	//========== Iterate and assigend to new struct- which is exact format of Report required.
	var whoApprovedWhatArr []WhoApprovedWhat
	for _, detailedInvoice := range detailedInvoiceArr {
		invoice := detailedInvoice.Invoice
		invStatArr := detailedInvoice.InvoiceStatus
		var whoApprovedWhat WhoApprovedWhat

		for _, invStat := range invStatArr {
			if invStat.Status == status {

				//convert string to time for comaprision
				//Time format in invstat 20180614
				tstart, err := ShortDateFromString(startDate)
				fmt.Errorf("cannot parse startdate: %v", err)
				tend, err := ShortDateFromString(endDate)
				fmt.Errorf("cannot parse endDate: %v", err)
				incStateDate := invStat.Time.Time()

				//Comparision Logic- invStat.Time >= startDate && invStat.Time <= endDate
				if incStateDate.After(tstart) || incStateDate.Equal(tstart) && incStateDate.Before(tend) || incStateDate.Equal(tend) {

					// Get data from Invoice
					whoApprovedWhat.ScanID = invoice.DcDocumentData.DcHeader.ScanID
					whoApprovedWhat.DocNumber = invoice.DcDocumentData.DcHeader.InvoiceNumber
					whoApprovedWhat.VendorID = invoice.DcDocumentData.DcHeader.VendorID
					whoApprovedWhat.VendorName = GetVendorName(stub, invoice.DcDocumentData.DcHeader.VendorID)
					whoApprovedWhat.Total = invoice.DcDocumentData.DcHeader.TotalAmount
					whoApprovedWhat.Currency = invoice.DcDocumentData.DcHeader.CurrencyCode
					whoApprovedWhat.DocDate = invoice.DcDocumentData.DcHeader.DocDate

					// geta data from other entity/model
					whoApprovedWhat.ApprovedBy = invStat.BuyerId
					whoApprovedWhat.ApproverAccount = invStat.UserId
					whoApprovedWhat.ERPPosting = "Unknown" // TODO
					whoApprovedWhatArr = append(whoApprovedWhatArr, whoApprovedWhat)

				}
			}
		}

	}

	myLogger.Debugf("ERROR parsing detail invoice :", whoApprovedWhatArr)
	return shim.Success(util.MarshalToBytes(whoApprovedWhatArr))
}

/**
Method to return vendor name by Vendor Id

**/
func GetVendorName(stub shim.ChaincodeStubInterface, vendorId string) string {
	vendor, fetchErr := vmd.GetVendor(stub, "", vendorId, "")
	if fetchErr != "" {
		myLogger.Debugf("Error in fetching vendor record for vendor ID : ", vendorId)
		return ""
	}
	return vendor.VendorName
}

//ShortDateFromString parse shot date from string
func ShortDateFromString(ds string) (time.Time, error) {
	t, err := time.Parse(bcdLayout, ds)
	if err != nil {
		return t, err
	}
	return t, nil
}

//CheckDataBoundariesStr checks is startdate <= enddate
func CheckDataBoundariesStr(startdate, enddate string) (bool, error) {

	tstart, err := ShortDateFromString(startdate)
	if err != nil {
		return false, fmt.Errorf("cannot parse startdate: %v", err)
	}
	tend, err := ShortDateFromString(enddate)
	if err != nil {
		return false, fmt.Errorf("cannot parse enddate: %v", err)
	}

	if tstart.After(tend) {
		return false, fmt.Errorf("startdate > enddate - please set proper data boundaries")
	}
	return true, err
}


