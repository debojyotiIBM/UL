/*
   Copyright IBM Corp. 2018 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Santosh Penubothula
*/

package invoice

import (
	"encoding/json"
	"regexp"
	"time"
	"unicode"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
)

const dUPLICATE_CHECK_NEXT_STEP = "st-postfacto-1"

type DuplicateInvoiceCache struct {
	InvoiceNumber string
	Total_Amount  float64
	Vendor_ID     string
	CurrencyCode  string
	DocDate       util.BCDate
}

func PerformDuplicateInvoiceCheck(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invNumChanged bool) (int, string, InvoiceStatus) {

	var errStr string
	var invStat InvoiceStatus
	inv_num_wo_sc := getInvoiceNumberWithoutSpecialCharactersAndLeadingZeros(contextObjPtr.Invoice.DcDocumentData.DcHeader.InvoiceNumber)

	var code int

	code, errStr, invStat = checkForPotentialDuplicate(stub, contextObjPtr.Invoice.DcDocumentData.DcHeader.InvoiceNumber, contextObjPtr)
	if code != 0 {
		return code, errStr, invStat
	}

	if !invNumChanged {

		code, errStr, invStat = checkForPotentialDuplicate(stub, inv_num_wo_sc, contextObjPtr)
		if code != 0 {
			return code, errStr, invStat
		}

		if isFirstAlpha(inv_num_wo_sc) {

			code, errStr, invStat = checkForPotentialDuplicate(stub, inv_num_wo_sc[1:], contextObjPtr)
			if code != 0 {
				return code, errStr, invStat
			}

			if isLastAlpha(inv_num_wo_sc) {

				code, errStr, invStat = checkForPotentialDuplicate(stub, inv_num_wo_sc[1:len(inv_num_wo_sc)-1], contextObjPtr)
				if code != 0 {
					return code, errStr, invStat
				}

			}
		}
		if isLastAlpha(inv_num_wo_sc) {

			code, errStr, invStat = checkForPotentialDuplicate(stub, inv_num_wo_sc[:len(inv_num_wo_sc)-1], contextObjPtr)
			if code != 0 {
				return code, errStr, invStat
			}

		}
	}
	myLogger.Debugf("Potential Duplicate not found\n")
	invStat = UpdateInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", dUPLICATE_CHECK_NEXT_STEP, EMPTY_ADDITIONAL_INFO)
	return 1, errStr, invStat
}

func checkForPotentialDuplicate(stub shim.ChaincodeStubInterface, inv_num string, contextObjPtr *Context) (int, string, InvoiceStatus) {

	var scanid_bciidXdupInvCache map[string]DuplicateInvoiceCache
	scanid_bciidXdupInvCache_temp, present := contextObjPtr.DupInvoiceCache[inv_num]
	if !present {
		data, _ := db.TableStruct{Stub: stub, TableName: util.TAB_DUPLICATE_INVOICE_CACHE, PrimaryKeys: []string{inv_num}}.GetBytes()
		if len(data) > 0 {
			json.Unmarshal(data, &scanid_bciidXdupInvCache)
		} else {
			return 0, "", InvoiceStatus{}
		}
		contextObjPtr.DupInvoiceCache[inv_num] = scanid_bciidXdupInvCache
	} else {
		scanid_bciidXdupInvCache = scanid_bciidXdupInvCache_temp
	}

	dirtyCKeys := []string{}
	dupFound := false
	for cKey, dupInvCache := range scanid_bciidXdupInvCache {
		myLogger.Debugf("%v %v %v\n", cKey, " ::: ", string(util.MarshalToBytes(dupInvCache)))
		scan_id, bci_array, _ := stub.SplitCompositeKey(cKey)
		bci_id := bci_array[0]
		//date check and mark removal if older than 1 year
		if !isWithinOneYear(dupInvCache.DocDate) || !isInvoiceNumberPotentialDuplicate(inv_num, dupInvCache.InvoiceNumber) {
			//myLogger.Debugf("Cached invoice not within one year or changed inv_num, bci_id : %v %v %v %v %v %v %v\n", bci_id, " scan_id : ", scan_id, " inv_num : ", dupInvCache.InvoiceNumber, " date : ", dupInvCache.DocDate.String())
			dirtyCKeys = append(dirtyCKeys, cKey)
			continue
		}
		if bci_id == contextObjPtr.Invoice.BCIID && scan_id == contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID {
			continue
		}
		//myLogger.Debugf("Invoice details ::: %v, %v, %v\n", contextObjPtr.Invoice.DcDocumentData.DcHeader.VendorID, contextObjPtr.Invoice.DcDocumentData.DcHeader.CurrencyCode, contextObjPtr.Invoice.DcDocumentData.DcHeader.TotalAmount)
		if util.EqualsIgnoreCase(dupInvCache.Vendor_ID, contextObjPtr.Invoice.DcDocumentData.DcHeader.VendorID) && util.EqualsIgnoreCase(dupInvCache.CurrencyCode, contextObjPtr.Invoice.DcDocumentData.DcHeader.CurrencyCode) && contextObjPtr.Invoice.DcDocumentData.DcHeader.TotalAmount == dupInvCache.Total_Amount {
			myLogger.Debugf("Potential Duplicate with invoice BCIID : %v %v %v %v %v\n", bci_id, " Scan ID : ", scan_id, " Inv no. : ", dupInvCache.InvoiceNumber)
			dupFound = true
		}
	}

	if len(dirtyCKeys) > 0 {
		for _, cKey := range dirtyCKeys {
			delete(scanid_bciidXdupInvCache, cKey)
			myLogger.Debugf("Deleting old/changed invoice %v from duplicate cache key %v\n", cKey, inv_num)
		}
		contextObjPtr.DupInvoiceCache[inv_num] = scanid_bciidXdupInvCache
		db.TableStruct{Stub: stub, TableName: util.TAB_DUPLICATE_INVOICE_CACHE, PrimaryKeys: []string{inv_num}}.AddBytes(util.MarshalToBytes(scanid_bciidXdupInvCache))
	}
	if dupFound {
		invStat, errStr := SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_DIC_DUPLICATE_FOUND, "", "st-DuplicateInvoiceCheck-ap1", EMPTY_ADDITIONAL_INFO)
		return 2, errStr, invStat
	} else {
		myLogger.Debugf(":::: NO HIT :::: in duplicate cache %v\n", inv_num)
	}
	return 0, "", InvoiceStatus{}
}

func DuplicateInvoiceCheck_IBM_AP_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {

	var errStr string
	if invStat.Status == REJECT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-DuplicateInvoiceCheck-end", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == UPDATED_INVOICE_NUMBER {
		invStat = UpdateInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, "st-DuplicateInvoiceCheck-InvNumChanged", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	} else if invStat.Status == "" || invStat.Status == CONTINUE || invStat.Status == SUBMIT {
		invStat = UpdateInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, dUPLICATE_CHECK_NEXT_STEP, EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	}
	return 2, errStr, invStat
}

func isInvoiceNumberPotentialDuplicate(keyInvNum string, InvoiceNumber string) bool {

	inv_num_wo_sc := getInvoiceNumberWithoutSpecialCharactersAndLeadingZeros(InvoiceNumber)

	if keyInvNum == inv_num_wo_sc {
		return true
	}
	if isFirstAlpha(inv_num_wo_sc) {
		if keyInvNum == inv_num_wo_sc[1:] {
			return true
		}
		if isLastAlpha(inv_num_wo_sc) {
			if keyInvNum == inv_num_wo_sc[1:len(inv_num_wo_sc)-1] {
				return true
			}
		}
	}
	if isLastAlpha(inv_num_wo_sc) {
		if keyInvNum == inv_num_wo_sc[:len(inv_num_wo_sc)-1] {
			return true
		}
	}
	return false
}

func AddInvoiceForDuplicateCheck(stub shim.ChaincodeStubInterface, contextObjPtr *Context) {

	if !isWithinOneYear(contextObjPtr.Invoice.DcDocumentData.DcHeader.DocDate) {
		myLogger.Debugf("Add invoice not within one year, bci_id : %v %v %v %v %v %v %v\n", contextObjPtr.Invoice.BCIID, " scan_id : ", contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, " inv_num : ", contextObjPtr.Invoice.DcDocumentData.DcHeader.InvoiceNumber, " date : ", contextObjPtr.Invoice.DcDocumentData.DcHeader.DocDate.String())
		return
	}
	inv_num_wo_sc := getInvoiceNumberWithoutSpecialCharactersAndLeadingZeros(contextObjPtr.Invoice.DcDocumentData.DcHeader.InvoiceNumber)
	dupInvCacheEntry := createDuplicateInvCacheEntry(&contextObjPtr.Invoice)
	updateCacheEntryForInvoiceNumber(stub, contextObjPtr, contextObjPtr.Invoice.DcDocumentData.DcHeader.InvoiceNumber, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, contextObjPtr.Invoice.BCIID, dupInvCacheEntry)
	updateCacheEntryForInvoiceNumber(stub, contextObjPtr, inv_num_wo_sc, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, contextObjPtr.Invoice.BCIID, dupInvCacheEntry)
	if isFirstAlpha(inv_num_wo_sc) {
		updateCacheEntryForInvoiceNumber(stub, contextObjPtr, inv_num_wo_sc[1:], contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, contextObjPtr.Invoice.BCIID, dupInvCacheEntry)
		if isLastAlpha(inv_num_wo_sc) {
			updateCacheEntryForInvoiceNumber(stub, contextObjPtr, inv_num_wo_sc[1:len(inv_num_wo_sc)-1], contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, contextObjPtr.Invoice.BCIID, dupInvCacheEntry)
		}
	}
	if isLastAlpha(inv_num_wo_sc) {
		updateCacheEntryForInvoiceNumber(stub, contextObjPtr, inv_num_wo_sc[:len(inv_num_wo_sc)-1], contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, contextObjPtr.Invoice.BCIID, dupInvCacheEntry)
	}
}

func isWithinOneYear(date util.BCDate) bool {
	time1 := date.Time()
	time2 := time.Now()
	timeL := time.Date(time2.Year()-1, time2.Month(), time2.Day(), 0, 0, 0, 0, time2.Location())
	timeL = timeL.AddDate(0, 0, 1)
	timeU := time.Date(time2.Year()+1, time2.Month(), time2.Day(), 0, 0, 0, 0, time2.Location())
	timeU = timeU.AddDate(0, 0, -1)
	if time1.Before(timeL) || time1.After(timeU) {
		return false
	}
	return true
}

func updateCacheEntryForInvoiceNumber(stub shim.ChaincodeStubInterface, contextObjPtr *Context, inv_num string, Scan_ID string, BCIID string, dupInvCacheEntry DuplicateInvoiceCache) {

	var scanid_bciidXdupInvCache map[string]DuplicateInvoiceCache
	scanid_bciidXdupInvCache_temp, present := contextObjPtr.DupInvoiceCache[inv_num]
	if !present {
		data, _ := db.TableStruct{Stub: stub, TableName: util.TAB_DUPLICATE_INVOICE_CACHE, PrimaryKeys: []string{inv_num}}.GetBytes()
		if len(data) > 0 {
			json.Unmarshal(data, &scanid_bciidXdupInvCache)
		} else {
			scanid_bciidXdupInvCache = make(map[string]DuplicateInvoiceCache)
		}
		contextObjPtr.DupInvoiceCache[inv_num] = scanid_bciidXdupInvCache
	} else {
		scanid_bciidXdupInvCache = scanid_bciidXdupInvCache_temp
	}
	cKey, _ := stub.CreateCompositeKey(Scan_ID, []string{BCIID})
	scanid_bciidXdupInvCache[cKey] = dupInvCacheEntry
	contextObjPtr.DupInvoiceCache[inv_num] = scanid_bciidXdupInvCache
	db.TableStruct{Stub: stub, TableName: util.TAB_DUPLICATE_INVOICE_CACHE, PrimaryKeys: []string{inv_num}}.AddBytes(util.MarshalToBytes(scanid_bciidXdupInvCache))
}

func createDuplicateInvCacheEntry(invoice *Invoice) DuplicateInvoiceCache {
	dupInvCache := DuplicateInvoiceCache{}
	dupInvCache.InvoiceNumber = invoice.DcDocumentData.DcHeader.InvoiceNumber
	dupInvCache.Total_Amount = invoice.DcDocumentData.DcHeader.TotalAmount
	dupInvCache.Vendor_ID = invoice.DcDocumentData.DcHeader.VendorID
	dupInvCache.CurrencyCode = invoice.DcDocumentData.DcHeader.CurrencyCode
	dupInvCache.DocDate = invoice.DcDocumentData.DcHeader.DocDate
	return dupInvCache
}

func isFirstAlpha(inv_num string) bool {
	for _, c := range inv_num {
		return unicode.IsLetter(c)
	}
	return false
}

func isLastAlpha(inv_num string) bool {
	ret_val := false
	for _, c := range inv_num {
		ret_val = unicode.IsLetter(c)
	}
	return ret_val
}

func getInvoiceNumberWithoutSpecialCharactersAndLeadingZeros(inv_num string) string {
	reg_sc, _ := regexp.Compile("[^a-zA-Z0-9]+")
	reg_lead_zeros, _ := regexp.Compile("^0+")
	return reg_lead_zeros.ReplaceAllString(reg_sc.ReplaceAllString(inv_num, ""), "")
}
