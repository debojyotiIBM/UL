/*
   Copyright IBM Corp. 2018 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Lohit Krishnan
*/

package invoice

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	"github.com/ibm/p2p/grn"
	"github.com/ibm/p2p/po"
)

func PerformGRNSelection(stub shim.ChaincodeStubInterface, contextObjPtr *Context) (int, string, InvoiceStatus) {
	const NEXT_STEP = "st-DCGR-1"
	var errStr string
	var invStat InvoiceStatus

	var totalACValue float64

	for _, dcLine := range contextObjPtr.Invoice.DcDocumentData.AdditionalLineItems {
		totalACValue += dcLine.Amount
	}

	var totalInvoiceValue = contextObjPtr.Invoice.DcDocumentData.DcSwissHeader.TotalNet - totalACValue

	// 3.13.1 : Check for 2-way/3-way.

	result, err := checkForNwayMatch(stub, &contextObjPtr.Invoice)
	if err != "" {
		switch err {
		case "DUPLICATE POLINES DETECTED":
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, err, "", "st-GRNSelection-ap1", EMPTY_ADDITIONAL_INFO)
		case "POLINES NOT FOUND FOR INVOICE":
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, err, "", "st-GRNSelection-ap2", EMPTY_ADDITIONAL_INFO)
		case "GOODSRECEIPT FLAG ISSUE":
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, err, "", "st-GRNSelection-ap3", EMPTY_ADDITIONAL_INFO)
		default:
			errStr = "UNKNOWN ERROR"
		}
		return 2, errStr, invStat
	}
	myLogger.Debugf("GS : After Nway match : ", result)

	switch result {
	case "2way":
		//3.13.48 : GRN selection step is not performed in this case.
		invStat = UpdateInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", NEXT_STEP, EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	case "both":
		// If invoice has lines with both 2-way and 3-way match, fail it to IBM-AP.
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_COMPLETED, EXIST_2WAY_3WAY_VGRN, "", "st-GRNSelection-end", EMPTY_ADDITIONAL_INFO)
		return 2, errStr, invStat
	default:
		break
	}

	// Function getPOxActivePOLines_NonActivePOLinesForInvoice is written in LineSelection.go
	mPOxActivePOLines_NonActivePOLines := getPOxActivePOLines_NonActivePOLinesForInvoice(stub, &contextObjPtr.Invoice)

	myLogger.Debugf("GS : Number of POs in the map = ", len(mPOxActivePOLines_NonActivePOLines))

	grnsActivePOLines := getGRNSForActivePOLines(stub, &mPOxActivePOLines_NonActivePOLines)

	myLogger.Debugf("GS : Number of GRNs for Active PO Lines = ", len(grnsActivePOLines))

	grnsAfterRefDocFilter := filterGRNBasedOnRefDocNumber(contextObjPtr, &grnsActivePOLines)

	myLogger.Debugf("GS : Number of GRNs after RefDoc filter = ", len(grnsAfterRefDocFilter))
	//3.13.3
	if len(grnsAfterRefDocFilter) >= 1 {
		//3.13.4
		AggregateGRNValue := GetGRNValue(stub, contextObjPtr, grnsAfterRefDocFilter)
		myLogger.Debugf("GS : AggregateGRN Value = ", AggregateGRNValue)
		myLogger.Debugf("TotalInvoiceValue = ", totalInvoiceValue)
		if AggregateGRNValue != totalInvoiceValue {
			//3.13.15 & 3.13.6
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, GRN_REF_BASED_VALUE_MISMATCH, "", "st-GRNSelection-ap5", EMPTY_ADDITIONAL_INFO)
			return 2, errStr, invStat
		}
		// TODO : Utilize the GRNs - Assign them as used up.
		// Do GRN Selection as per the grnsAfterRefDocFilter Set and go to next step.
		return reconstructInvoiceBasedOnGRNs(stub, contextObjPtr, grnsAfterRefDocFilter)

	}
	//3.13.16 - Whether Open GRN Available for the selected PO Lines.
	if isOpenGRNAvailableForSelectedPOLines(stub, &contextObjPtr.Invoice) == false {
		// 3.13.47 : Send to Buyer
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_BUYER, OPEN_GRNS_NOT_AVAILABLE, "", "st-GRNSelection-buyer1", EMPTY_ADDITIONAL_INFO)
		return 2, errStr, invStat
	}
	myLogger.Debugf("GS : Open GRNs available for PO Lines")
	//3.13.17 - Check whether UP is same for all invoice lines belonging to same PO Line.
	if isUPSameForAllInvoicePOLines(&contextObjPtr.Invoice) == false {
		// 3.13.46 : Send to AP
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, UP_NOT_MATCHING_FOR_SAME_PO_LINES, "", "st-GRNSelection-ap6", EMPTY_ADDITIONAL_INFO)
		return 2, errStr, invStat
	}

	myLogger.Debugf("GS : UP is same for all PO Lines")

	//3.13.19 : Full utilization of PO, Invoice by one GRN
	if len(grnsActivePOLines) == 1 &&
		len(mPOxActivePOLines_NonActivePOLines) == 1 {
		cnt := 0
		invoiceGRNQtyMatches := false
		for _, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {
			cnt = cnt + 1
			if dcLine.Quantity == getGRNResidualQuantity(contextObjPtr, grnsActivePOLines[0]) {
				invoiceGRNQtyMatches = true
			}
		}
		if cnt == 1 && invoiceGRNQtyMatches {
			for _, poLines := range mPOxActivePOLines_NonActivePOLines {
				for _, poLine := range poLines[0] {
					if len(poLines[0]) == 1 && poLine.Quantity == getGRNResidualQuantity(contextObjPtr, grnsActivePOLines[0]) {
						return reconstructInvoiceBasedOnGRNs(stub, contextObjPtr, grnsActivePOLines)
					}
				}
			}
		}
	}
	myLogger.Debugf("GS : Full Utilization of complete invoice with one GRN which consumes complete PO")

	grnsForInvoice := getGRNSForInvoicePOLines(stub, &contextObjPtr.Invoice)

	myLogger.Debugf("GS : GRNs for this invoice = ", grnsForInvoice)
	//3.13.21 : Sort all Open GRN in chronological order.
	//3.13.22 : Filter out all GRN with Ref Number
	//3.13.24 : Filter GRN with GRNDate < InvoiceDate - 5
	grnsAfterThreeFilter := filterGRNSOnThreeCondition(grnsForInvoice, contextObjPtr.Invoice.DcDocumentData.DcHeader.DocDate)

	myLogger.Debugf("GS : GRNs after the three filters. = ", grnsAfterThreeFilter)
	//3.13.25 : Are GRNs available after all the above filters?
	if len(grnsAfterThreeFilter) == 0 {
		// 3.13.44 : Send to Buyer
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_BUYER, OPEN_GRNS_NOT_AVAILABLE, "", "st-GRNSelection-buyer1", EMPTY_ADDITIONAL_INFO)
		return 2, errStr, invStat
	}
	//3.13.26
	// Single GRNNumber having exact same quantities as Invoice lines.
	// Group grns based on the grn numbers. Check if any group satisfies all the quantities exactly
	GRNNumberxGRNs := groupGRNSBasedOnGRNNumber(grnsAfterThreeFilter)

	resultFlag, GRNNumber := checkIfAnyCompleteGRNSatisfiesInvoice(contextObjPtr, GRNNumberxGRNs)
	if resultFlag == true {
		grnList := GRNNumberxGRNs[GRNNumber]
		return reconstructInvoiceBasedOnGRNs(stub, contextObjPtr, grnList)
	}
	myLogger.Debugf("GS : Complete GRN doesn't satisfy the invoice ")
	//3.13.27
	// Exact match in different grns.
	resultFlag, grnList := checkInvoiceLineExactMatchWithDifferentGRNS(contextObjPtr, grnsAfterThreeFilter)
	if resultFlag == true {
		return reconstructInvoiceBasedOnGRNs(stub, contextObjPtr, grnList)
	}
	myLogger.Debugf("GS : Exact Match with different GRN doesn't exist ")

	//3.13.29
	// subset sum of grns having exact match.
	// TODO : Finish this function
	resultFlag, grnList = checkInvoiceLineExactMatchWithSubsetSumGRNS(contextObjPtr, grnsAfterThreeFilter)
	if resultFlag == true {
		myLogger.Debugf("GS : Reconstructing using SubSet Sum result : grnList = ", grnList)
		return reconstructInvoiceBasedOnGRNs(stub, contextObjPtr, grnList)
	}
	myLogger.Debugf("GS : Subset Sum Match didn't work. Trying Best Effort Selection. ")

	//3.13.33
	// Try to utilize as much grn's as possible and then keep others residual.
	return PerformBestEffortGRNSelection(stub, contextObjPtr, grnsAfterThreeFilter)

}

type timeSlice []grn.GRN

func (p timeSlice) Len() int {
	return len(p)
}

func (p timeSlice) Less(i, j int) bool {
	return p[i].DocPostDate.Time().Before(p[j].DocPostDate.Time())
}

func (p timeSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func filterGRNSOnThreeCondition(grns []grn.GRN, InvoiceDate util.BCDate) []grn.GRN {
	var result []grn.GRN
	for _, GRN := range grns {
		// TODO : Change to Business Days.
		invDateMinusFive := InvoiceDate.Time().AddDate(0, 0, -5)
		if GRN.RefDocNumber == "" && GRN.DocPostDate.Time().After(invDateMinusFive) {
			result = append(result, GRN)
		}
	}
	date_sorted_grns := make(timeSlice, 0, len(result))
	for _, d := range result {
		date_sorted_grns = append(date_sorted_grns, d)
	}
	sort.Sort(date_sorted_grns)
	return date_sorted_grns

}

func getAllSubsets(inputSet [][]int) [][]int {

	setLen := len(inputSet)
	if setLen == 1 {
		return inputSet
	}

	var outputSet [][]int

	firstElement := inputSet[0]

	inputSet = inputSet[1:]

	oldSet := getAllSubsets(inputSet)

	outputSet = append(outputSet, firstElement)

	for idx, _ := range oldSet {
		outputSet = append(outputSet, oldSet[idx])

		tempSet := make([]int, len(oldSet[idx]))
		copy(tempSet, oldSet[idx])
		tempSet = append(tempSet, firstElement[0])

		outputSet = append(outputSet, tempSet)

	}
	return outputSet

}

func PerformSubsetSum(set []float64, sum float64) (bool, []int) {

	inputSet := make([][]int, len(set))

	for i := 0; i < len(set); i++ {
		inputSet[i] = make([]int, 1)
		inputSet[i][0] = i
	}

	outputSet := getAllSubsets(inputSet)

	fmt.Println("\n")
	for _, val := range outputSet {
		fmt.Println(" ", val, " ")
	}

	for i, _ := range outputSet {
		curSum := 0.0
		for j, _ := range outputSet[i] {
			curSum = curSum + set[outputSet[i][j]]
		}
		if curSum == sum {
			fmt.Println("Got the Subset sum solution ", outputSet[i])
			return true, outputSet[i]
		}
	}
	fmt.Println("NO SOLUTION FOUND")
	return false, outputSet[0]

}

func checkInvoiceLineExactMatchWithSubsetSumGRNS(contextObjPtr *Context, grns []grn.GRN) (bool, []grn.GRN) {
	var outputGRNList []grn.GRN
	//return false, outputGRNList

	contextObjPtr.StoreInvoice = true
	mInvLinexGRNS := make(map[int][]grn.GRN)
	selectedGRNIdx := make(map[int]bool)

	for invLineIdx, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {
		currInvLineMatched := false
		var grnList []grn.GRN
		invLineQty := dcLine.Quantity

		var selectedGRNQty float64
		for grnIdx, GRN := range grns {
			if selectedGRNIdx[grnIdx] == true {
				continue
			}
			if GRN.PONumber == dcLine.PoNumber && GRN.POItemNumber == dcLine.PoLine {
				if getGRNResidualQuantity(contextObjPtr, GRN)+selectedGRNQty < invLineQty {
					//Update selected GRN Qty
					selectedGRNQty = selectedGRNQty + getGRNResidualQuantity(contextObjPtr, GRN)

					grnList = append(grnList, GRN)
					selectedGRNIdx[grnIdx] = true

				} else {
					currInvLineMatched = true
					selectedGRNIdx[grnIdx] = true
					grnList = append(grnList, GRN)
					//break
				}
			}
		}
		if currInvLineMatched == false {
			// Need to forward to the buyer..
			//FwdToBuyerFlag = true
			return false, outputGRNList
		}
		mInvLinexGRNS[invLineIdx] = grnList
	}

	for i, grns := range mInvLinexGRNS {
		invLineQty := contextObjPtr.Invoice.DcDocumentData.DcLines[i].Quantity
		grnQtyArr := make([]float64, len(grns))
		for grnIdx, GRN := range grns {
			grnQtyArr[grnIdx] = getGRNResidualQuantity(contextObjPtr, GRN)
		}
		myLogger.Debugf("GS : For Invoice Line no = ", i)
		myLogger.Debugf("SubSet Sum on grnQtyArr = ", grnQtyArr)
		result, arrayIndices := PerformSubsetSum(grnQtyArr, invLineQty)
		myLogger.Debugf("SubSet Sum result ArrayIndices = ", arrayIndices)
		if result == false {
			return false, outputGRNList
		}
		for grnIdx, GRN := range grns {
			for _, arrIdx := range arrayIndices {
				if grnIdx == arrIdx {
					outputGRNList = append(outputGRNList, GRN)
					break
				}
			}
		}
	}
	return true, outputGRNList

}

func checkInvoiceLineExactMatchWithDifferentGRNS(contextObjPtr *Context, grns []grn.GRN) (bool, []grn.GRN) {

	var grnList []grn.GRN
	mGRNUtilized := make(map[int]bool)
	for _, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {
		currInvLineMatched := false
		for grnIdx, GRN := range grns {
			if mGRNUtilized[grnIdx] == false && GRN.PONumber == dcLine.PoNumber && GRN.POItemNumber == dcLine.PoLine && dcLine.Quantity == getGRNResidualQuantity(contextObjPtr, GRN) {
				mGRNUtilized[grnIdx] = true
				currInvLineMatched = true
				grnList = append(grnList, GRN)
				break
			}
		}
		if currInvLineMatched == false {
			return false, grnList
		}
	}
	return true, grnList
}

func checkIfAnyCompleteGRNSatisfiesInvoice(contextObjPtr *Context, grnMap map[string][]grn.GRN) (bool, string) {

	for GRNNumber, GRNList := range grnMap {
		mInvoiceLinesUtilized := make(map[int]bool, len(contextObjPtr.Invoice.DcDocumentData.DcLines))
		currGRNMatched := false
		for _, GRN := range GRNList {
			currGRNMatched = false
			for invIdx, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {
				if GRN.PONumber == dcLine.PoNumber && GRN.POItemNumber == dcLine.PoLine && dcLine.Quantity == getGRNResidualQuantity(contextObjPtr, GRN) {
					currGRNMatched = true
					mInvoiceLinesUtilized[invIdx] = true
					break
				}
			}
			if !currGRNMatched {
				break
			}
		}
		result := isAllTrue(mInvoiceLinesUtilized, len(contextObjPtr.Invoice.DcDocumentData.DcLines))
		if result == true && currGRNMatched == true {
			return true, GRNNumber
		}
	}
	return false, ""
}

func isAllTrue(boolMap map[int]bool, expectedCnt int) bool {
	curCnt := 0
	for _, r := range boolMap {
		if r == false {
			return false
		} else {
			curCnt = curCnt + 1
		}
	}
	myLogger.Debugf("==== Expected Count = ", expectedCnt)
	myLogger.Debugf("==== Actual Count = ", curCnt)
	if expectedCnt == curCnt {
		return true
	} else {
		return false
	}
}

func groupGRNSBasedOnGRNNumber(grns []grn.GRN) map[string][]grn.GRN {
	mGRNNumberxGRNs := make(map[string][]grn.GRN)

	for _, GRN := range grns {
		cGRNNumber := GRN.GrnNumber
		mGRNNumberxGRNs[cGRNNumber] = append(mGRNNumberxGRNs[cGRNNumber], GRN)
	}
	return mGRNNumberxGRNs
}

func getGRNSForInvoicePOLines(stub shim.ChaincodeStubInterface, invoice *Invoice) []grn.GRN {

	var grns []grn.GRN
	erpSystem := invoice.DcDocumentData.DcHeader.ErpSystem
	for _, dcLine := range invoice.DcDocumentData.DcLines {
		cPONumber := dcLine.PoNumber
		cPOLine := dcLine.PoLine
		grnRecords, _ := db.TableStruct{Stub: stub, TableName: util.TAB_GRN, PrimaryKeys: []string{erpSystem, cPONumber, util.GetStringFromInt(cPOLine)}, Data: ""}.GetAll()
		for _, grnRow := range grnRecords {
			var currentGRN grn.GRN
			json.Unmarshal([]byte(grnRow), &currentGRN)
			grns = append(grns, currentGRN)
		}

	}
	return grns
}

func isOpenGRNAvailableForSelectedPOLines(stub shim.ChaincodeStubInterface, invoice *Invoice) bool {
	erpSystem := invoice.DcDocumentData.DcHeader.ErpSystem
	for _, dcLine := range invoice.DcDocumentData.DcLines {
		cPONumber := dcLine.PoNumber
		cPOLine := dcLine.PoLine
		grnRecords, _ := db.TableStruct{Stub: stub, TableName: util.TAB_GRN, PrimaryKeys: []string{erpSystem, cPONumber, util.GetStringFromInt(cPOLine)}, Data: ""}.GetAll()
		if len(grnRecords) == 0 {
			return false
		}
	}
	return true
}

func isUPSameForAllInvoicePOLines(invoice *Invoice) bool {
	mPO_POLinexUP := make(map[string]float64)
	for _, dcLine := range invoice.DcDocumentData.DcLines {
		cPONumber := dcLine.PoNumber
		cPOLine := dcLine.PoLine
		key := cPONumber + util.GetStringFromInt(cPOLine)
		if mPO_POLinexUP[key] == 0 {
			mPO_POLinexUP[key] = dcLine.UnitPrice
		} else {
			if mPO_POLinexUP[key] != dcLine.UnitPrice {
				return false
			}
		}
	}
	return true
}

func PerformBestEffortGRNSelection(stub shim.ChaincodeStubInterface, contextObjPtr *Context, grns []grn.GRN) (int, string, InvoiceStatus) {
	const NEXT_STEP = "st-DCGR-1"
	contextObjPtr.StoreInvoice = true
	mInvLinexGRNS := make(map[int][]grn.GRN)
	selectedGRNIdx := make(map[int]bool)
	FwdToBuyerFlag := false
	for invLineIdx, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {
		currInvLineMatched := false
		var grnList []grn.GRN
		invLineQty := dcLine.Quantity

		var selectedGRNQty float64
		for grnIdx, GRN := range grns {
			if selectedGRNIdx[grnIdx] == true {
				continue
			}
			if GRN.PONumber == dcLine.PoNumber && GRN.POItemNumber == dcLine.PoLine {
				if getGRNResidualQuantity(contextObjPtr, GRN)+selectedGRNQty < invLineQty {
					//Update selected GRN Qty
					selectedGRNQty = selectedGRNQty + getGRNResidualQuantity(contextObjPtr, GRN)

					grnList = append(grnList, GRN)
					selectedGRNIdx[grnIdx] = true

				} else {
					currInvLineMatched = true
					selectedGRNIdx[grnIdx] = true
					grnList = append(grnList, GRN)
					break
				}
			}
		}
		if currInvLineMatched == false {
			// Need to forward to the buyer..
			FwdToBuyerFlag = true
		}
		mInvLinexGRNS[invLineIdx] = grnList
	}
	// Reconstruct the invoice and update the residual quantities

	newDCLines := []DCLine{}
	taxPercentage := -1
	commonDeliveryNote := ""
	for _, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {
		if commonDeliveryNote == "" {
			commonDeliveryNote = dcLine.DeliveryNote
		}
		if commonDeliveryNote != dcLine.DeliveryNote {
			commonDeliveryNote = "NOT COMMON"
		}
		if taxPercentage == -1 {
			taxPercentage = int(dcLine.TaxAmount * 10000 / dcLine.Amount)
		} else if taxPercentage == 0 || taxPercentage != int(dcLine.TaxAmount*10000/dcLine.Amount) {
			taxPercentage = -2
		}
	}
	for i, grns := range mInvLinexGRNS {
		invLineQty := contextObjPtr.Invoice.DcDocumentData.DcLines[i].Quantity
		var selectedGRNQty float64
		for _, GRN := range grns {
			var newLineQty float64
			if getGRNResidualQuantity(contextObjPtr, GRN)+selectedGRNQty < invLineQty {
				newLineQty = getGRNResidualQuantity(contextObjPtr, GRN)
			} else {
				newLineQty = invLineQty - selectedGRNQty
			}
			selectedGRNQty = selectedGRNQty + newLineQty
			newDCLine := DCLine{}
			newDCLine.PoNumber = GRN.PONumber
			newDCLine.PoLine = GRN.POItemNumber
			newDCLine.Amount = GRN.GRNValue
			newDCLine.Quantity = getGRNResidualQuantity(contextObjPtr, GRN)
			if taxPercentage == -1 {
				newDCLine.TaxAmount = -1
				newDCLine.TaxPercent = -1
			} else {
				newDCLine.TaxAmount = newDCLine.Amount * (float64(taxPercentage) / 10000)
				newDCLine.TaxPercent = float64(taxPercentage) / 100
			}
			poLineItemsRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS, PrimaryKeys: []string{GRN.ErpSystem, GRN.PONumber, util.GetStringFromInt(GRN.POItemNumber)}, Data: ""}.GetAll()
			var POLine po.POLineItem
			for _, poLineItemRow := range poLineItemsRec {
				json.Unmarshal([]byte(poLineItemRow), &POLine)
				break
			}
			newDCLine.UnitPrice = POLine.UnitPrice
			newDCLine.MatNumber = POLine.MaterialNumber
			newDCLine.Description = POLine.Description
			newDCLine.GrnMatch = append(newDCLine.GrnMatch, GRN)

			scanIDSuffix := getScanIDSuffix(contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID)
			if GRN.RefDocNumber != contextObjPtr.Invoice.DcDocumentData.DcHeader.InvoiceNumber && GRN.RefDocNumber != scanIDSuffix {
				newDCLine.DeliveryNote = GRN.RefDocNumber
			} else {
				if commonDeliveryNote != "NOT COMMON" {
					newDCLine.DeliveryNote = commonDeliveryNote
				}
			}

			newDCLines = append(newDCLines, newDCLine)

			updateGRNResidualQuantity(contextObjPtr, GRN, invLineQty)
			if selectedGRNQty == invLineQty {
				break
			}
		}
		if selectedGRNQty < invLineQty {
			//One more line item without GRN tagging.
			FwdToBuyerFlag = false
			newDCLine := DCLine{}
			newDCLine.PoNumber = contextObjPtr.Invoice.DcDocumentData.DcLines[i].PoNumber
			newDCLine.PoLine = contextObjPtr.Invoice.DcDocumentData.DcLines[i].PoLine
			newDCLine.Quantity = (invLineQty - selectedGRNQty)

			if taxPercentage == -1 {
				newDCLine.TaxAmount = -1
				newDCLine.TaxPercent = -1
			} else {
				newDCLine.TaxAmount = newDCLine.Amount * (float64(taxPercentage) / 10000)
				newDCLine.TaxPercent = float64(taxPercentage) / 100
			}
			ERPSystem := contextObjPtr.Invoice.DcDocumentData.DcHeader.ErpSystem
			PONumber := contextObjPtr.Invoice.DcDocumentData.DcLines[i].PoNumber
			POLineItemNumber := contextObjPtr.Invoice.DcDocumentData.DcLines[i].PoLine

			poLineItemsRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS, PrimaryKeys: []string{ERPSystem, PONumber, util.GetStringFromInt(POLineItemNumber)}, Data: ""}.GetAll()
			var POLine po.POLineItem
			for _, poLineItemRow := range poLineItemsRec {
				json.Unmarshal([]byte(poLineItemRow), &POLine)
				break
			}
			newDCLine.UnitPrice = POLine.UnitPrice
			newDCLine.Amount = newDCLine.UnitPrice * newDCLine.Quantity
			newDCLine.MatNumber = POLine.MaterialNumber
			newDCLine.Description = POLine.Description

			newDCLine.DeliveryNote = contextObjPtr.Invoice.DcDocumentData.DcLines[i].DeliveryNote

			newDCLines = append(newDCLines, newDCLine)
		}
	}
	contextObjPtr.Invoice.DcDocumentData.DcLines = newDCLines

	if FwdToBuyerFlag == true {
		invStat, errStr := SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_BUYER, OPEN_GRNS_NOT_AVAILABLE, "", "st-GRNSelection-buyer1", EMPTY_ADDITIONAL_INFO)
		return 2, errStr, invStat
	} else {
		invStat := UpdateInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", NEXT_STEP, EMPTY_ADDITIONAL_INFO)
		return 1, "", invStat
	}
	// END
}

//Input :
//  1. ContextObj
//  2. GRN (based on which you should reconstruct)
//  3. nonAcDCLines
//Output :
//  Output of stage (status, errStr, invStat)
func reconstructInvoiceBasedOnGRNs(stub shim.ChaincodeStubInterface, contextObjPtr *Context, grns []grn.GRN) (int, string, InvoiceStatus) {
	contextObjPtr.StoreInvoice = true
	const NEXT_STEP = "st-DCGR-1"
	newDCLines := []DCLine{}
	taxPercentage := -1
	commonDeliveryNote := ""
	for _, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {
		if commonDeliveryNote == "" {
			commonDeliveryNote = dcLine.DeliveryNote
		}
		if commonDeliveryNote != dcLine.DeliveryNote {
			commonDeliveryNote = "NOT COMMON"
		}
		if taxPercentage == -1 {
			taxPercentage = int(dcLine.TaxAmount * 10000 / dcLine.Amount)
		} else if taxPercentage == 0 || taxPercentage != int(dcLine.TaxAmount*10000/dcLine.Amount) {
			taxPercentage = -2
		}
	}

	for _, GRN := range grns {

		newDCLine := DCLine{}
		invLineQty := getGRNResidualQuantity(contextObjPtr, GRN)
		newDCLine.PoNumber = GRN.PONumber
		newDCLine.PoLine = GRN.POItemNumber
		newDCLine.Amount = GRN.GRNValue
		newDCLine.Quantity = invLineQty
		if taxPercentage == -1 {
			newDCLine.TaxAmount = -1
			newDCLine.TaxPercent = -1
		} else {
			newDCLine.TaxAmount = newDCLine.Amount * (float64(taxPercentage) / 10000)
			newDCLine.TaxPercent = float64(taxPercentage) / 100
		}
		poLineItemsRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS, PrimaryKeys: []string{GRN.ErpSystem, GRN.PONumber, util.GetStringFromInt(GRN.POItemNumber)}, Data: ""}.GetAll()
		var POLine po.POLineItem
		for _, poLineItemRow := range poLineItemsRec {
			json.Unmarshal([]byte(poLineItemRow), &POLine)
			break
		}
		newDCLine.UnitPrice = POLine.UnitPrice
		newDCLine.MatNumber = POLine.MaterialNumber
		newDCLine.Description = POLine.Description
		newDCLine.GrnMatch = append(newDCLine.GrnMatch, GRN)

		scanIDSuffix := getScanIDSuffix(contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID)
		if GRN.RefDocNumber != contextObjPtr.Invoice.DcDocumentData.DcHeader.InvoiceNumber && GRN.RefDocNumber != scanIDSuffix {
			newDCLine.DeliveryNote = GRN.RefDocNumber
		} else {
			if commonDeliveryNote != "NOT COMMON" {
				newDCLine.DeliveryNote = commonDeliveryNote
			}
		}

		newDCLines = append(newDCLines, newDCLine)
		updateGRNResidualQuantity(contextObjPtr, GRN, invLineQty)
	}

	contextObjPtr.Invoice.DcDocumentData.DcLines = newDCLines
	myLogger.Debugf("LS : DCline delivery note no., GRN based reconstruction and line selection")
	invStat := UpdateInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", NEXT_STEP, EMPTY_ADDITIONAL_INFO)
	return 1, "", invStat

}

// This function returns (result, error) where
// result can be either (2way, 3way or both)
// If poLine.GoodsReceiptFlag == 'X' then it is 3 way, else it is 2 way.
//If there are both 2way and 3way in different po lines for the same invoice, then return "both"
func checkForNwayMatch(stub shim.ChaincodeStubInterface, invoice *Invoice) (string, string) {
	var result = ""
	for _, dcLine := range invoice.DcDocumentData.DcLines {
		cPONumber := dcLine.PoNumber
		cPOLine := util.GetStringFromInt(dcLine.PoLine)
		cERPSystem := invoice.DcDocumentData.DcHeader.ErpSystem
		poLineItemsRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS, PrimaryKeys: []string{cERPSystem, cPONumber, cPOLine}, Data: ""}.GetAll()
		if len(poLineItemsRec) > 1 {
			// Error : There are two entries with the same <ERPSystem, PO, POLine> combination
			// Reason : "Duplicate POLines detected", Status : "IBM-AP Action Pending"
			return "", DUPLICATE_POLINES
		}
		if len(poLineItemsRec) == 0 {
			// Error : There are no entries with the same <ERPSystem, PO, POLine> combination
			// Reason : "No POLines found for the InvoiceLines", Status : "IBM-AP Action Pending"
			return "", POLINES_NOT_FOUND
		}
		for _, poLineItemRow := range poLineItemsRec {
			var currentPOLineItem po.POLineItem
			json.Unmarshal([]byte(poLineItemRow), &currentPOLineItem)
			if currentPOLineItem.GoodsReceiptFlag == "X" {
				if result == "2way" {
					result = "both"
					return result, ""
				}
				result = "3way"
			} else if currentPOLineItem.GoodsReceiptFlag == "" {
				if result == "3way" {
					result = "both"
					return result, ""
				}
				result = "2way"
			} else {
				return "", GOODSRECEIPT_FLAG_ISSUE
			}
		}
	}
	return result, ""
}

func GetGRNValue(stub shim.ChaincodeStubInterface, contextObjPtr *Context, grns []grn.GRN) float64 {
	var totalGRNValue float64
	totalGRNValue = 0
	for _, GRN := range grns {

		grnUnitPrice := getUnitPriceFromGRN(stub, GRN)
		myLogger.Debugf("grnUnitPrice = ", grnUnitPrice)

		residualQty := getGRNResidualQuantity(contextObjPtr, GRN)
		myLogger.Debugf("residualQty = ", residualQty)

		grnResidualValue := residualQty * grnUnitPrice
		myLogger.Debugf("grnResidualValue = ", grnResidualValue)

		totalGRNValue = totalGRNValue + grnResidualValue
	}
	return totalGRNValue
}

func getUnitPriceFromGRN(stub shim.ChaincodeStubInterface, GRN grn.GRN) float64 {
	myLogger.Debugf("function : getUnitPriceFromGRN")
	myLogger.Debugf("GRN.ErpSystem = ", GRN.ErpSystem)
	myLogger.Debugf("GRN.PONumber = ", GRN.PONumber)
	myLogger.Debugf("GRN.POItemNumber = ", GRN.POItemNumber)
	poLineItemsRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS, PrimaryKeys: []string{GRN.ErpSystem, GRN.PONumber, util.GetStringFromInt(GRN.POItemNumber)}, Data: ""}.GetAll()
	var POLine po.POLineItem
	for _, poLineItemRow := range poLineItemsRec {
		json.Unmarshal([]byte(poLineItemRow), &POLine)
		break
	}
	return POLine.UnitPrice
}

func filterGRNBasedOnRefDocNumber(contextObjPtr *Context, grns *[]grn.GRN) []grn.GRN {

	var result []grn.GRN
	scanIDSuffix := getScanIDSuffix(contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID)

	for _, GRN := range *grns {
		currGRNmatched := false
		if GRN.RefDocNumber == "" {
			continue
		}
		if GRN.RefDocNumber != contextObjPtr.Invoice.DcDocumentData.DcHeader.InvoiceNumber && GRN.RefDocNumber != scanIDSuffix {
			for _, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {
				if dcLine.DeliveryNote == GRN.RefDocNumber {
					currGRNmatched = true
					break
				}
			}
		} else {
			currGRNmatched = true
		}
		if currGRNmatched {
			result = append(result, GRN)
		}
	}
	return result
}

func GRNSelection_IBM_AP_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {

	const NEXT_STEP = "st-DCGR-1"

	var errStr string
	if invStat.Status == FWD_TO_BUYER {
		invStat = ForwardToBuyer(stub, &contextObjPtr.Invoice, invStat)
		/*
			tokens := strings.Split(invStat.InternalStatus, "ap")
			newInternalStatus := "st-GRNSelection-buy" + tokens[len(tokens)-1]
		*/
		newInternalStatus := "st-GRNSelection-buyer2"
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, newInternalStatus, EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == REJECT {
		RevertGRNResidualForCompleteInvoice(stub, contextObjPtr)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-GRNSelection-end", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == PROCESS_OUTSIDE_BC {
		RevertGRNResidualForCompleteInvoice(stub, contextObjPtr)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_COMPLETED, COMPLETED_OUTSIDE_BC, invStat.Comments, "st-GRNSelection-end", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == "" || invStat.Status == CONTINUE {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, NEXT_STEP, EMPTY_ADDITIONAL_INFO)
	}
	return 2, errStr, invStat
}

func GRNSelection_Buyer_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {

	// UI STATUS TO BE HANDLED
	// 1) AWAITING BUYER ACTION HOLD INVOICE
	var errStr string
	if invStat.Status == FWD_TO_OTHER_BUYER {

		errStr, invStat = ForwardToOtherBuyer(stub, &contextObjPtr.Invoice, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, invStat.InternalStatus, EMPTY_ADDITIONAL_INFO)
		}
	} else if invStat.Status == REJECT {
		RevertGRNResidualForCompleteInvoice(stub, contextObjPtr)
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-GRNSelection-end", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == GRN_CREATED {

		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_WAITING_FOR_GRN, invStat.ReasonCode, invStat.Comments, "st-GRNSelection-1", EMPTY_ADDITIONAL_INFO)

	} else if invStat.Status == RETURN_TO_AP {
		newInternalStatus := "st-GRNSelection-ap7"
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, newInternalStatus, EMPTY_ADDITIONAL_INFO)

	} else if invStat.Status == BUYER_DELEGATION {

		errStr, invStat = BuyerDelegation(stub, contextObjPtr, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, invStat.Status, invStat.ReasonCode, invStat.Comments, invStat.InternalStatus, EMPTY_ADDITIONAL_INFO)
		}
	} else if invStat.Status == ADDITIONAL_PO {
		RevertGRNResidualForCompleteInvoice(stub, contextObjPtr)
		myLogger.Debugf("GRN Selection : entered Additional PO loop=====================")

		invStat = UpdateInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, ALT_PO, invStat.ReasonCode, invStat.Comments, "ST0000", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat

	} else if invStat.Status == ALT_PO {
		RevertGRNResidualForCompleteInvoice(stub, contextObjPtr)

		myLogger.Debugf("GRN Selection : entered alternate PO loop=====================")
		invStat = UpdateInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, ALT_PO, invStat.ReasonCode, invStat.Comments, "ST0000", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	}
	return 2, errStr, invStat
}
