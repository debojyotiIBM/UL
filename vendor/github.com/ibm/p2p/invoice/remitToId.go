/*

   Copyright IBM Corp. 2017 All Rights Reserved.

   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");

   @author : Saptaswa Sarkar

*/

package invoice

import (
	"encoding/json"

	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"

	"github.com/ibm/p2p/po"

	vmd "github.com/ibm/p2p/vmd"

	logging "github.com/op/go-logging"
)

var logger = logging.MustGetLogger("RemitToId")

func RemitToId(stub shim.ChaincodeStubInterface, contextObjPtr *Context) (int, string, InvoiceStatus) {

	var invoice *Invoice = &(contextObjPtr.Invoice)

	var errStr string

	var invStat InvoiceStatus

	logger.Debugf("invoice in RemitToId ================", invoice)

	removeSpecialChar(invoice)

	var matchedVendor vmd.Vendor = getMatchedVendor(stub, invoice)

	if len(matchedVendor.VendorID) == 0 {

		//IBM AP

		logger.Debugf("Matched Vendor not found")

		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_VENDOR_NOT_EXIST_IN_VMD, "", "st-remitToId-01", EMPTY_ADDITIONAL_INFO)

		return 2, errStr, invStat

	}

	vendID := invoice.DcDocumentData.DcHeader.VendorID

	// Vendor object already found, then processing

	var isVendorActiveFlag bool

	//to check if the bank account is active

	if strings.EqualFold(matchedVendor.IsDeletedFlag, "No") || strings.EqualFold(matchedVendor.IsDeletedFlag, "") {

		isVendorActiveFlag = true

	} else {

		// to check if vendor is not active from Compnay Code

		vendCompanyDetailsMap := getMatchedCompanyCodeDetails(stub, invoice.DcDocumentData.DcHeader.CompanyCode)

		var isDeleted string

		var cPostingBlock string

		if vendCompanyDetailsMap != nil && len(vendCompanyDetailsMap) > 0 {

			isDeleted = vendCompanyDetailsMap["isdeletedflag"]

			cPostingBlock = vendCompanyDetailsMap["cpostingblock"]

		} else {

			logger.Debugf("Vendor ID in VMD - ", matchedVendor.VendorID, "is Not Active. Company code Needs Extension")

			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_COMPANY_CODE_MATCH, "", "st-remitToId-01", EMPTY_ADDITIONAL_INFO)

			return 2, errStr, invStat

		}

		if len(isDeleted) != 0 && strings.EqualFold(isDeleted, "X") {

			//ui dynmanic decisioning

			logger.Debugf("Vendor ID in VMD - ", matchedVendor.VendorID, "is Not Active. IsDeletedFlag is ON for CompanyCode ")

			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_VENDOR_NOT_ACTIVE_COMPANYCODE_DELETEFLAG_ON, "", "st-remitToId-01", EMPTY_ADDITIONAL_INFO)

			return 2, errStr, invStat

		} else {

			if len(cPostingBlock) != 0 && strings.EqualFold(cPostingBlock, "X") {

				//ui dynmanic decisioning

				logger.Debugf("Vendor ID in VMD - ", matchedVendor.VendorID, "is Not Active. PostingBlockFlag is ON for CompanyCode")

				invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_VENDOR_NOT_ACTIVE_COMPANYCODE_POSTINGFLAG_ON, "", "st-remitToId-01", EMPTY_ADDITIONAL_INFO)

				return 2, errStr, invStat

			} else {

				isVendorActiveFlag = true

			}

		}

	}

	if isVendorActiveFlag {

		filteredBandkDetailsRecords := getBankDetailsForAVendor(stub, vendID)

		if len(invoice.DcDocumentData.DcHeader.BankAccount) != 0 {

			//to check if the bank accnt details match for vmd and invoice

			//get the bank details for the given vendor id and iterate

			flag := false

			for _, bankdDet := range filteredBandkDetailsRecords {

				if bankdDet["bankaccountorig"] == invoice.DcDocumentData.DcHeader.BankAccount {

					flag = true

					logger.Debugf("BankDetailsRecords - ", flag)

					break

				}

			}

			if flag {

				logger.Debugf("Bank Account matched with same in Invoice ")

				invStat = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", "st-DuplicateInvoiceCheck-start", EMPTY_ADDITIONAL_INFO)

				return 1, errStr, invStat

			} else {

				//IBM AP UI // INV_RS_BANK_AC_NOT_MATCH_VMD

				logger.Debugf("Bank Account does not match with the same in Invoice")

				invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID,

					invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_BANK_AC_NOT_MATCH_VMD, "", "st-remitToId-02", EMPTY_ADDITIONAL_INFO)

				return 2, errStr, invStat

			}

		} else {

			//If bank account is not present in Invoice, then check the same in VMD

			//to check the bank details in VMD for the matched vendorId

			logger.Debugf("No Bankaccount found in Invoice: checking bank details in Vendor details")

			if len(filteredBandkDetailsRecords) > 0 {

				if len(filteredBandkDetailsRecords) == 1 {

					//default to Vendor Id - //continue to UI
					// also to check if the there is Fiscal address available in Invoice
					if len(getFiscalAddress(stub, invoice, matchedVendor)) > 0 {

						logger.Debugf("Single Bank found for the Vendor Id ", matchedVendor.VendorID)

						invStat = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", "st-DuplicateInvoiceCheck-start", EMPTY_ADDITIONAL_INFO)

						return 1, errStr, invStat

					}

				} else {

					//IBM AP  UI

					logger.Debugf("Multiple Bank Details found")

					invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_MULTIPLE_BANKS_AC_FOUND, "", "st-remitToId-03", EMPTY_ADDITIONAL_INFO)

					return 2, errStr, invStat

				}

			} else {

				//FISCAL addess check
				fiscalAddrs := getFiscalAddress(stub, invoice, matchedVendor)

				if len(fiscalAddrs) == 0 {
					logger.Debugf("No Fiscal Address found in invoice")

					invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_NO_FISCAL_ADDRESS_FOUND, "", "st-remitToId-03", EMPTY_ADDITIONAL_INFO)

					return 2, errStr, invStat
				}

				if fiscalAddressMatch(stub, invoice, matchedVendor) {

					logger.Debugf("Fiscal Address match successful")

					invStat = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", "st-DuplicateInvoiceCheck-start", EMPTY_ADDITIONAL_INFO)

					return 1, errStr, invStat

				} else {

					//IBM AP UI

					logger.Debugf("Fiscal Address does not match ")

					invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_FISCAL_ADD_NOT_MATCH, "", "st-remitToId-04", EMPTY_ADDITIONAL_INFO)

					return 2, errStr, invStat

				}

			}

		}

	} else {

		//VENDOR is INACTIVE - IBM AP Action

		//IBM AP

		logger.Debugf("Vendor is Inactive")

		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_VENDOR_INACTIVE, "", "st-remitToId-01", EMPTY_ADDITIONAL_INFO)

		return 2, errStr, invStat

	}

	return 1, errStr, invStat

}

func getMatchedVendor(stub shim.ChaincodeStubInterface, invoice *Invoice) vmd.Vendor {

	logger.Debug("From getMatchedVendor : ERP :", invoice.DcDocumentData.DcHeader.ErpSystem)

	logger.Debug("From getMatchedVendor : Vendor Id :", invoice.DcDocumentData.DcHeader.VendorID)

	logger.Debug("From getMatchedVendor : Client Id :", invoice.DcDocumentData.DcHeader.Client)

	vend, _ := vmd.GetVendor(stub, invoice.DcDocumentData.DcHeader.ErpSystem,

		invoice.DcDocumentData.DcHeader.VendorID, invoice.DcDocumentData.DcHeader.Client)

	logger.Debug("Matched vendor  is ", vend.VendorID)

	return vend

}

//This method extract the CompanyCode details for a given Company Code from Invoice
func getMatchedCompanyCodeDetails(stub shim.ChaincodeStubInterface, companyCode string) map[string]string {

	resp := vmd.GetAllVendorsCompanyCode(stub)

	vendorsJSONBytes := resp.GetPayload()

	vendors := make([]map[string]string, 0)

	err := json.Unmarshal(vendorsJSONBytes, &vendors)

	logger.Debugf("Error is ", err)

	if err != nil {

	}

	logger.Debugf("vendors - data : ", vendors)

	for _, vendorDetalsMap := range vendors {

		if vendorDetalsMap["companycode"] == companyCode {

			return vendorDetalsMap

		}

	}

	return nil

}

//This method matches the required Fiscal address check of vendor' address with the SupplierAddress of invor

func fiscalAddressMatch(stub shim.ChaincodeStubInterface, invoice *Invoice, matchedVendor vmd.Vendor) bool {

	swshdr := invoice.DcDocumentData.DcSwissHeader

	fAddress := getFiscalAddress(stub, invoice, matchedVendor)

	var flag bool

	//TODO to change the logic for matching the address

	if swshdr.SupplierName == matchedVendor.VendorName && fAddress == matchedVendor.Address {

		flag = true

	}

	return flag

}

func getFiscalAddress(stub shim.ChaincodeStubInterface, invoice *Invoice, matchedVendor vmd.Vendor) string {

	swshdr := invoice.DcDocumentData.DcSwissHeader

	var fiscalAddress = [10]string{swshdr.SupplierAddress1, swshdr.SupplierAddress2,

		swshdr.SupplierAddress3, swshdr.SupplierAddress4, swshdr.SupplierAddress5,

		swshdr.SupplierAddress6, swshdr.SupplierAddress7, swshdr.SupplierAddress8,

		swshdr.SupplierAddress9, swshdr.SupplierAddress10}

	var fAddress string

	for _, FA := range fiscalAddress {

		fAddress += FA

	}

	return fAddress

}

//This method extracts the matched Bank details from Bank Details Table
func getBankDetailsForAVendor(stub shim.ChaincodeStubInterface, vendID string) []map[string]string {

	//should return a array of bankdetails for a given VenodrId

	resp := vmd.GetAllVendorsBank(stub)

	vendorsBankDetsJSONBytes := resp.GetPayload()

	logger.Debugf("vendorsBankDetsJSONBytes :", string(vendorsBankDetsJSONBytes))

	logger.Debugf("vendID ::::: ", vendID)

	filteredRecords := make([]map[string]string, 0)

	records := make([]map[string]string, 0)

	json.Unmarshal(vendorsBankDetsJSONBytes, &records)

	logger.Debugf("length of records :", len(records))

	for _, record := range records {

		logger.Debugf("vendor id from bank data ", record["vendorid"])

		if record["vendorid"] == vendID {

			filteredRecords = append(filteredRecords, record)

		}

	}

	return filteredRecords

}

//Select_Remit_TO_ID_IBM_AP_Action is to accept the invoice status
func Select_Remit_TO_ID_IBM_AP_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {

	var errStr string

	if invStat.Status == FWD_TO_BUYER {

		var invoice *Invoice = &(contextObjPtr.Invoice)

		invStat = ForwardToBuyer(stub, invoice, invStat)

		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-remitToId-21", EMPTY_ADDITIONAL_INFO)

	} else if invStat.Status == BUYER_DELEGATION {

		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-remitToId-22", EMPTY_ADDITIONAL_INFO)

	} else if invStat.Status == RETURN_TO_AP {

		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-remitToId-23", EMPTY_ADDITIONAL_INFO)

		return 2, errStr, invStat

	} else if invStat.Status == REJECT {

		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-remitToId-24", EMPTY_ADDITIONAL_INFO)

	} else if invStat.Status == PROCESS_OUTSIDE_BC {

		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, COMPLETED_OUTSIDE_BC, invStat.ReasonCode, invStat.Comments, "st-remitToId-25", EMPTY_ADDITIONAL_INFO)
	}

	return 2, errStr, invStat

}

//Select_Remit_TO_ID_Buyer_Action is to accept buyer action
func Select_Remit_TO_ID_Buyer_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {

	var invoice *Invoice = &(contextObjPtr.Invoice)

	var errStr string

	if invStat.Status == FWD_TO_OTHER_BUYER {

		logger.Debugf("Entered  into FWD_TO_OTHER_BUYER")

		errStr, invStat = ForwardToOtherBuyer(stub, invoice, invStat)

		if errStr == "" {

			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, "st-remitToId-31", EMPTY_ADDITIONAL_INFO)

		}

	} else if invStat.Status == BUYER_DELEGATION {

		logger.Debugf("Entered  into Buyer_Action - BUYER_DELEGATION")

		errStr, invStat = BuyerDelegation(stub, contextObjPtr, invStat)

		if errStr == "" {

			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, invStat.Status, invStat.ReasonCode, invStat.Comments, invStat.InternalStatus, EMPTY_ADDITIONAL_INFO)

		}

	} else if invStat.Status == REJECT {

		logger.Debugf("Entered  into Buyer_Action - REJECT")

		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-remitToId-32", EMPTY_ADDITIONAL_INFO)

	} else if invStat.Status == RETURN_TO_AP {

		logger.Debugf("Entered  into Buyer_Action - RETURN_TO_AP")

		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, "st-remitToId-33", EMPTY_ADDITIONAL_INFO)

	} else if invStat.Status == ALT_PO {

		po, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.PoNumber, invoice.DcDocumentData.DcHeader.ErpSystem, invoice.DcDocumentData.DcHeader.Client})

		if fetchErr != "" {

			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, INV_RS_INVALID_PO, "", "ST0000", EMPTY_ADDITIONAL_INFO)

			return 2, errStr, invStat

		}

		logger.Debugf("Entered  into Buyer_Action - ALT_PO", po)

		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_WAITING_INVOICE_FIX, invStat.ReasonCode, invStat.Comments, "st-remitToId-34", EMPTY_ADDITIONAL_INFO)

	}

	return 2, errStr, invStat

}

//removeSpecialChar will remove '-' character, if any, from Bank Account from Invoice and after that invoice will have bank accountnumber with out '-'
func removeSpecialChar(invoice *Invoice) {

	bAcc := invoice.DcDocumentData.DcHeader.BankAccount

	str := strings.Replace(bAcc, "-", "", -1)

	invoice.DcDocumentData.DcHeader.BankAccount = str
}

