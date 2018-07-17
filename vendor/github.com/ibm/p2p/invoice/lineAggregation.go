/*
   Copyright IBM Corp. 2018 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Sudip Dutta ( suddutt1@in.ibm.com)
*/

package invoice

import (
	"fmt"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	util "github.com/ibm/p2p"
	logging "github.com/op/go-logging"
)

var _lc_la_logger = logging.MustGetLogger("Line-Aggregation-Log")

//Local constant . To hide from the rest of the world added _lc at the begining
const _lc_GO_TO_NEXT_STEP = "st-LineSelection-start"

// LineAggregration method perform the step 3.8.
// Aggregates the duplication lines items of an invoice
func LineAggregration(stub shim.ChaincodeStubInterface, trxnContext *Context) (int, string, InvoiceStatus) {

	var invStat InvoiceStatus
	invoice := trxnContext.Invoice
	lineItems := invoice.DcDocumentData.DcLines
	_lc_la_logger.Debugf("Size of line items : %d", len(lineItems))

	aggregrationMap := make(map[string]DCLine)
	aggregatedLineItems := make([]DCLine, 0)
	ignoreTaxAtLineLevel := (invoice.DcDocumentData.DcHeader.TaxAmount > 0)
	for _, lineItem := range lineItems {
		isNormalLI := performAggregration(aggregrationMap, lineItem, ignoreTaxAtLineLevel)
		if !isNormalLI {
			//This entry is line tax, total etc. so adding in the slice.
			aggregatedLineItems = append(aggregatedLineItems, lineItem)
		}
	}
	//Now add the aggregreated items
	for _, aggrLineItem := range aggregrationMap {
		aggregatedLineItems = append([]DCLine{aggrLineItem}, aggregatedLineItems...)
	}
	//Set back the original items in the trxn context object
	invoice.DcDocumentData.DcLines = aggregatedLineItems
	_lc_la_logger.Debugf("Size of line items after aggregration: %d", len(aggregatedLineItems))
	trxnContext.Invoice = invoice
	AddInvoice(stub, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, string(util.MarshalToBytes(invoice)), false)
	additonalInfo := AdditionalInfo{Value: fmt.Sprintf("Size of line items after aggregration: %d", len(invoice.DcDocumentData.DcLines)), Type1: "string"}

	invStat, _ = SetInvoiceStatus(stub, trxnContext, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", _lc_GO_TO_NEXT_STEP, additonalInfo)

	return 1, "", invStat
}

//Aggregreate the line items into a map with material number, unit price and description as key
func performAggregration(dataMap map[string]DCLine, item DCLine, ignoreTax bool) bool {
	key := ""
	poNumber := normalizeString(item.PoNumber)
	if strings.EqualFold("AC", poNumber) || strings.EqualFold("ADDITIONAL COST", poNumber) {
		//This is an addition cost item. Do not aggregrate
		return false
	}
	materialNumber := normalizeString(item.MatNumber)
	desc := normalizeString(item.Description)
	if len(materialNumber) > 0 {
		if !ignoreTax {
			key = fmt.Sprintf("%s_%s_%s_%.4f_%.4f", poNumber, materialNumber, "_", item.UnitPrice, item.TaxPercent)
		} else {
			key = fmt.Sprintf("%s_%s_%s_%.4f", poNumber, materialNumber, "_", item.UnitPrice)
		}
		_lc_la_logger.Debugf("M-> Key generated: %s", key)
	} else if len(desc) > 0 && item.UnitPrice > 0 {
		if !ignoreTax {
			key = fmt.Sprintf("%s_%s_%s_%.4f_%.4f", poNumber, "_", desc, item.UnitPrice, item.TaxPercent)
		} else {
			key = fmt.Sprintf("%s_%s_%s_%.4f", poNumber, "_", desc, item.UnitPrice)
		}
		_lc_la_logger.Debugf("D-> Key generated: %s", key)
	}
	if len(key) > 0 {

		if existingEntry, entryExists := dataMap[key]; entryExists {
			existingEntry.Quantity = existingEntry.Quantity + item.Quantity
			existingEntry.Amount = existingEntry.Quantity * existingEntry.UnitPrice
			existingEntry.TaxAmount = existingEntry.TaxAmount + item.TaxAmount
			dataMap[key] = existingEntry
		} else {
			dataMap[key] = item
		}
		return true

	}
	return false
}
func normalizeString(input string) string {
	return strings.ToUpper(strings.Replace(input, " ", "", -1))
}
