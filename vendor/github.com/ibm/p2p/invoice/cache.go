/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package invoice

import (
	//"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	//pb "github.com/hyperledger/fabric/protos/peer"
	//"github.com/ibm/db"
	util "github.com/ibm/p2p"
	grn "github.com/ibm/p2p/grn"
	po "github.com/ibm/p2p/po"
)

type GenericType struct {
	AlphaKey int64   `json:"AlphaKey"`
	BetaKey  float64 `json:"betaKey"`
	GammaKey string  `json:"gammaKey"`
}

func updateGRNResidualQuantity(contextObjPtr *Context, GRN grn.GRN, usedQty float64) {

	key := GRN.ErpSystem + GRN.PONumber + util.GetStringFromInt(GRN.POItemNumber) + GRN.GrnNumber + util.GetStringFromInt(GRN.GRNLineItem)
	if _, ok := contextObjPtr.UpdatedGRNs[key]; ok {
		derivedGRN := contextObjPtr.UpdatedGRNs[key]
		derivedGRN.ResidualQuantity = derivedGRN.ResidualQuantity - usedQty
		contextObjPtr.UpdatedGRNs[key] = derivedGRN
		return
	}
	GRN.ResidualQuantity = GRN.ResidualQuantity - usedQty
	contextObjPtr.UpdatedGRNs[key] = GRN

}

func getGRNResidualQuantity(contextObjPtr *Context, GRN grn.GRN) float64 {
	key := GRN.ErpSystem + GRN.PONumber + util.GetStringFromInt(GRN.POItemNumber) + GRN.GrnNumber + util.GetStringFromInt(GRN.GRNLineItem)
	if _, ok := contextObjPtr.UpdatedGRNs[key]; ok {
		derivedGRN := contextObjPtr.UpdatedGRNs[key]
		return derivedGRN.ResidualQuantity
	}
	return GRN.ResidualQuantity
}

func StoreGRNResiduals(stub shim.ChaincodeStubInterface, contextObjPtr *Context) {
	myLogger.Debugf("-----------------------------------------------------------------------------------")
	myLogger.Debugf("UPDATING RESIDUAL QUANTITIES FOR GRN . . . ")
	for _, GRN := range contextObjPtr.UpdatedGRNs {
		myLogger.Debugf("GRN Updated - ", string(util.MarshalToBytes(GRN)))
		poitemNum := util.GetStringFromInt(GRN.POItemNumber)
		seq := util.GetStringFromInt(GRN.Sequence)
		matdocyear := util.GetStringFromInt(GRN.MatDocYear)
		grnlineitem := util.GetStringFromInt(GRN.GRNLineItem)
		grn.AddGRN(stub, []string{GRN.ErpSystem, GRN.PONumber, poitemNum, GRN.GrnNumber, grnlineitem, seq, GRN.TransType, matdocyear, GRN.Client}, string(util.MarshalToBytes(GRN)))
	}
}

func StorePOResiduals(stub shim.ChaincodeStubInterface, contextObjPtr *Context) {
	myLogger.Debugf("-----------------------------------------------------------------------------------")
	myLogger.Debugf("UPDATING RESIDUAL QUANTITIES FOR PO . . .")

	for _, PO := range contextObjPtr.UpdatedPOs {
		myLogger.Debugf("PO Updated - ", string(util.MarshalToBytes(PO)))
		po.AddPO(stub, PO.PONumber, PO.ERPSystem, PO.Client, string(util.MarshalToBytes(PO)))
	}

	for _, poLine := range contextObjPtr.UpdatedPOLines {
		myLogger.Debugf("POLine Updated - ", string(util.MarshalToBytes(poLine)))
		po.AddPOLineItems(stub, poLine.PONumber, poLine.ERPSystem, poLine.Client, poLine.LineItemNumber, string(util.MarshalToBytes(poLine)))
	}
}

func RevertGRNResidualForCompleteInvoice(stub shim.ChaincodeStubInterface, contextObjPtr *Context) {

	for i, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {
		GRNsToBeReverted := dcLine.GrnMatch
		for _, GRN := range GRNsToBeReverted {
			key := GRN.ErpSystem + GRN.PONumber + util.GetStringFromInt(GRN.POItemNumber) + GRN.GrnNumber + util.GetStringFromInt(GRN.GRNLineItem)
			if _, ok := contextObjPtr.UpdatedGRNs[key]; ok {
				derivedGRN := contextObjPtr.UpdatedGRNs[key]
				derivedGRN.ResidualQuantity = derivedGRN.ResidualQuantity + dcLine.Quantity
				contextObjPtr.UpdatedGRNs[key] = derivedGRN
			}
		}
		contextObjPtr.Invoice.DcDocumentData.DcLines[i].GrnMatch = nil
	}
	contextObjPtr.StoreInvoice = true
	StoreGRNResiduals(stub, contextObjPtr)
}

/*
func UpdateResidualPOBudget(po po.PO, usedBudget float64) {
	po.SetPoBudget(po.PoBudget - usedBudget)
	UPDATED_PO_FOR_BUDGET[ po.PONumber] = po
}

func GetResidualPOBudget(po po.PO) float64 {
	if _, ok := UPDATED_PO_FOR_BUDGET[ po.PONumber]; ok {
		derived_po := UPDATED_PO_FOR_BUDGET[ po.PONumber]
		return derived_po.PoBudget
	}
	return po.PoBudget
}

func UpdatePOResidual(po po.PO, lineNumber int64, quantityUsed float64) {
	var lineItems []po.POLineItem
	for _, line := range *po.LineItems() {
		if line.PoLine() == lineNumber {
			line.SetResidualQuantity(line.ResidualQuantity() - quantityUsed)
		}
		lineItems = append(lineItems, line)
	}
	po.SetLineItems(lineItems)
	contextObjPtr.UpdatedPOs[ po.PONumber] = po
}

func StorePOResiduals(stub shim.ChaincodeStubInterface) {
	myLogger.Debugf("-----------------------------------------------------------------------------------")
	myLogger.Debugf("UPDATING RESIDUAL QUANTITIES FOR PO . . .")
	var updatedPOForBudgets []string
	for _, po := range contextObjPtr.UpdatedPOs {
		if _, ok := UPDATED_PO_FOR_BUDGET[ po.PONumber]; ok {
			derived_po := UPDATED_PO_FOR_BUDGET[ po.PONumber]
			po.PoBudget = derived_po.PoBudget
			updatedPOForBudgets = append(updatedPOForBudgets,  po.PONumber)
		}
		myLogger.Debugf("PO Updated - ", string(util.MarshalToBytes(po)))
		po.AddPO(stub,  po.PONumber, po.ERPSystem, string(util.MarshalToBytes(po)))
	}

	for poNumber, po := range UPDATED_PO_FOR_BUDGET {
		if !(util.StringArrayContains(updatedPOForBudgets, poNumber)) {
			myLogger.Debugf("PO Updated - ", string(util.MarshalToBytes(po)))
			po.AddPO(stub,  po.PONumber, po.ERPSystem, string(util.MarshalToBytes(po)))
		}
	}
}

func GetUpdatedResidualPOQty(po po.PO, poLine po.POLineItem) float64 {
	var poResQty float64
	if _, ok := contextObjPtr.UpdatedPOs[ po.PONumber]; ok {
		derived_po := UPDATED_PO_FOR_BUDGET[ po.PONumber]
		for _, line := range *derived_po.LineItems() {
			if line.PoLine() == poLine.PoLine() {
				poResQty = line.ResidualQuantity()
			}
		}
	} else {
		for _, line := range *po.LineItems() {
			if line.PoLine() == poLine.PoLine() {
				poResQty = line.ResidualQuantity()
			}
		}
	}
	return poResQty
}

func UpdateGrnResidualQty(stub shim.ChaincodeStubInterface,contextObjPtr Context* grn grn.GRN, quantityUsed float64, invNumber string, invLineNumber int64) {
	myLogger.Debugf(">>>>>>>>>InvNumber:", invNumber)
	myLogger.Debugf(">>>>>>>>>InvLineNumber:", invLineNumber)
	(*grn.LineItems())[0].SetResidualQuantity((*grn.LineItems())[0].ResidualQuantity() - quantityUsed)
	myLogger.Debugf("grn - ", grn)
	if invNumber != "" && invLineNumber != 0 {
		var grnInvLineItems []grn.InvLine

		for _, inLineItem := range *grn.InvLineItems() {
			grnInvLineItems = append(grnInvLineItems, inLineItem)
		}

		var latestInvLineItem grn.InvLine
		latestInvLineItem.SetInvNumber(invNumber)
		latestInvLineItem.SetInvLineNumber(invLineNumber)
		grnInvLineItems = append(grnInvLineItems, latestInvLineItem)
		grn.SetInvLineItems(grnInvLineItems)

		myLogger.Debugf(">>>>>>>>>>>>>>>>>>Printing Array Values")
		for index, _ := range *grn.InvLineItems() {
			myLogger.Debugf(">>>> Arr[InvNo]", grnInvLineItems[index].InvNumber())
			myLogger.Debugf(">>>> Arr[LineNo]", grnInvLineItems[index].InvLineNumber())
		}

	}
	contextObjPtr.UpdatedGRNs[grn.GrnNumber()] = grn
	StoreGRNResiduals(stub)
}


func RevertGRNResidualQuantity(stub shim.ChaincodeStubInterface, invoice Invoice) {
	var INV_GRNMAP map[string]GenericType
	INV_GRNMAP = make(map[string]GenericType)

	var po_num = invoice.DcDocumentData.DcHeader.PoNumber
	var scanId = invoice.DcDocumentData.DcHeader.ScanID
	for idx, lineItem := range *invoice.DcLines() {
		if len(*lineItem.GrnMatch()) != 0 {
			for _, g := range *lineItem.GrnMatch() {
				var inv_grn_value GenericType
				var revertQuantity float64
				if lineItem.Quantity() == 1 {
					revertQuantity = lineItem.Amount()
				} else {
					revertQuantity = lineItem.Quantity()
				}
				inv_grn_value.SetAlphaKey(lineItem.PoLine())
				inv_grn_value.SetBetaKey(revertQuantity)
				inv_grn_value.SetGammaKey(scanId)
				INV_GRNMAP[grn.GrnNumber()] = inv_grn_value
				break
			}
		}
		(*invoice.DcLines())[idx].SetGrnMatch(nil)
	}
	myLogger.Debugf(">>>>>>>>>>>>>>>>>>Printing INV_GRNMAP Values")
	for grnNum, values := range INV_GRNMAP {
		myLogger.Debugf("Map Values:", grnNum, values.AlphaKey(), values.BetaKey(), values.GammaKey())
	}

	invoice_po_grns := grn.GetGrnsByPO(stub, po_num)
	for _, grn := range invoice_po_grns {
		grn_generic := INV_GRNMAP[grn.GrnNumber()]
		if grn_generic.AlphaKey() != 0 && grn_generic.BetaKey() != 0.0 {
			for idx, line := range *grn.LineItems() {
				if line.PoLineItemNumber() == grn_generic.AlphaKey() {
					var residualQty float64
					revertQty := grn_generic.BetaKey()
					residualQty = line.ResidualQuantity()
					residualQty = residualQty + revertQty
					(*grn.LineItems())[idx].SetResidualQuantity(residualQty)
				}
			}
			var newGrnInvLineItems []grn.InvLine
			for _, invLine := range *grn.InvLineItems() {
				if invLine.InvNumber() != grn_generic.GammaKey() {
					newGrnInvLineItems = append(newGrnInvLineItems, invLine)
				}
			}
			grn.SetInvLineItems(newGrnInvLineItems)
		}
		contextObjPtr.UpdatedGRNs[grn.GrnNumber()] = grn
	}
	StoreGRNResiduals(stub)
	myLogger.Debugf(">>>>>>>>>>>EXIT Func<<<<<<<<<<<<")
}

func StoreGRNResiduals(stub shim.ChaincodeStubInterface) {
	myLogger.Debugf("-----------------------------------------------------------------------------------")
	myLogger.Debugf("UPDATING RESIDUAL QUANTITIES FOR GRN . . . ")
	for _, grn := range contextObjPtr.UpdatedGRNs {
		myLogger.Debugf("GRN Updated - ", string(util.MarshalToBytes(grn)))
		grn.AddGRN(stub, []string{grn.BillOfLading(), grn.GrnNumber()}, string(util.MarshalToBytes(grn)))
	}
}

func GetUpdatedResidualGRNQty(grn grn.GRN) float64 {

	return (*grn.LineItems())[0].ResidualQuantity()
}

func PoBudgetRevert(stub shim.ChaincodeStubInterface, invoice Invoice, po po.PO) {
	myLogger.Debugf(" PoBudgetRevert-----------------------------------------------------------------------------------")
	myLogger.Debugf("Po==========", po)
	totalPoBudget := GetResidualPOBudget(po)
	myLogger.Debugf("totalPoBudget==========", totalPoBudget)
	myLogger.Debugf("RevicedTotalAmount string value=====", invoice.RevisedTotalAmount())
	myLogger.Debugf("RevicedTotalAmount==========", invoice.RevisedTotalAmount())
	revertBudget := totalPoBudget + invoice.RevisedTotalAmount()
	myLogger.Debugf("revertBudget==========", revertBudget)
	myLogger.Debugf("util.GetFloatFromString(revertBudget)", revertBudget)
	po.SetPoBudget(revertBudget)
	UPDATED_PO_FOR_BUDGET[ po.PONumber] = po
	myLogger.Debugf("Before storing residual=============", po)
	StorePOResiduals(stub)
}

*/
func UpdatePO(stub shim.ChaincodeStubInterface, contextObjPtr *Context, po po.PO) {
	myLogger.Debugf("Updating PO value==========", po)
	myLogger.Debugf("Before updation==========", contextObjPtr.UpdatedPOs[po.PONumber])
	contextObjPtr.UpdatedPOs[po.PONumber] = po
	myLogger.Debugf("After updation==========", contextObjPtr.UpdatedPOs[po.PONumber])
	//	StorePOResiduals(stub)
}

/*
func PartialGrn(stub shim.ChaincodeStubInterface, invoiceRecArr string) pb.Response {

	var invoices []Invoice
	var invoicesBySupplier map[string]string
	invoicesBySupplier = make(map[string]string)
	var invStat InvoiceStatus
	var errStr string
	var invoicesByBuyer map[string]string
	invoicesByBuyer = make(map[string]string)
	var li_number_counter int64 = 1
	var waitingForGrn = true
	var linesInfo string

	err := json.Unmarshal([]byte(invoiceRecArr), &invoices)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input invoice array:", err)
	}

	myLogger.Debugf("Chceking partial GRN cases===================")

	for _, invoice := range invoices {
		invoice = ManageDCLineItem(stub, invoice, false)
		myLogger.Debugf("check bciid=====================>", invoice.BCIID)
		var scanId = invoice.DcDocumentData.DcHeader.ScanID
		for idx, line := range *invoice.DcLines() {
			line.SetInvoiceLine(li_number_counter)
			li_number_counter = li_number_counter + 1
			var originalGRNMatch = line.GrnMatch()
			var TagGRNMatch []grn.GRN = *line.TagGRNMatch()
			var UnTagGRNMatch = *line.UnTagGRNMatch()
			myLogger.Debugf("originalGRNMatch,TagGRNMatch,UnTagGRNMatch", originalGRNMatch, TagGRNMatch, UnTagGRNMatch)
			if len(UnTagGRNMatch) != 0 {
				myLogger.Debugf("Entered untag condition================")
				RevertGRNResidualQuantityUnTag(stub, line, invoice)
				line.SetGrnMatch(nil)
				line.SetUnTagGRNMatch(nil)
				(*invoice.DcLines())[idx] = line
			}
			if len(TagGRNMatch) != 0 {
				line.SetGrnMatch(TagGRNMatch)
				var selectedGrns []grn.GRN

				for _, g := range *line.GrnMatch() {
					myLogger.Debugf("g ---- > ", string(util.MarshalToBytes(g)))
					if line.Quantity() == 1.0 {
						UpdateGrnResidualQty(stub, g, line.UnitPrice(), scanId, line.InvoiceLine())
					} else {
						UpdateGrnResidualQty(stub, g, line.Quantity(), scanId, line.InvoiceLine())
					}

					selectedGrns = append(selectedGrns, g)
					myLogger.Debugf("grn updated", g)
					line.SetGrnMatch(selectedGrns)
					myLogger.Debugf("line.GRNMatch=========", line.GrnMatch())
					line.SetTagGRNMatch(nil)
				}

				(*invoice.DcLines())[idx] = line

			}

			if line.GrnMatch() == nil {

				waitingForGrn = false
			}
		}
		//Updating the invoice lines
		myLogger.Debugf("Invoice before updating===================>", invoice)
		AddInvoice(stub, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, string(util.MarshalToBytes(invoice)), false)
		invoiceKey := invoice.BCIID + "~" + invoice.DcDocumentData.DcHeader.ScanID

		if invoice.DcDocumentData.DcHeader.VendorID != "" {
			if invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] != "" {
				invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] = invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] + "|" + invoiceKey
			} else {
				invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] = invoiceKey
			}
		}

		// collect invoices by buyer
		var po po.PO
		poRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO, PrimaryKeys: []string{invoice.DcDocumentData.DcHeader.PoNumber}, Data: ""}.Get()
		err := json.Unmarshal([]byte(poRecord), &po)
		if err != nil {
			myLogger.Debugf("ERROR in parsing po record:", err)
		}
		if po.BuyerID != "" {
			if invoicesByBuyer[po.BuyerID] != "" {
				invoicesByBuyer[po.BuyerID] = invoicesByBuyer[po.BuyerID] + "|" + invoiceKey
			} else {
				invoicesByBuyer[po.BuyerID] = invoiceKey
			}
		}

		//Updating Details in supplier table
		for supplierId, val := range invoicesBySupplier {
			util.UpdateReferenceData(stub, util.TAB_INVOICE_BY_SUPPLIER, []string{supplierId}, val)
		}

		//Updating Details in Buyer table
		for buyerId, val := range invoicesByBuyer {
			util.UpdateReferenceData(stub, util.TAB_INVOICE_BY_BUYER, []string{buyerId}, val)
		}

		//Set invoice for further processing
		SetInvoiceForProcessing(invoice)
		STORE_INVOICE = true
		if waitingForGrn == true {
			invStat = UpdateInvoiceStatus(stub, contextObjPtr, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PROCESSING, "", "", "ST01101", EMPTY_ADDITIONAL_INFO)
			checkStatus(stub, 1, errStr, invStat)
		} else {
			invStat, errStr = SetInvoiceStatus(stub, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_WAITING_FOR_GRN, INV_RS_GRN_INV_HOLD, "", "ST0909", CreateAdditionalInfo("INVOICE LINES MISSING GRN", linesInfo))
			checkStatus(stub, 2, errStr, invStat)
		}
		myLogger.Debugf("Final Invoice===================", invoice)

	}
	return shim.Success(nil)
}

func RevertGRNResidualQuantityUnTag(stub shim.ChaincodeStubInterface, lineItem DCLine, invoice Invoice) {

	var INV_GRNMAP map[string]GenericType
	INV_GRNMAP = make(map[string]GenericType)
	myLogger.Debugf("lineItem.UnTagGRNMatch===========", lineItem.UnTagGRNMatch())
	var scanID = invoice.DcDocumentData.DcHeader.ScanID
	var po_num = invoice.DcDocumentData.DcHeader.PoNumber
	if len(*lineItem.UnTagGRNMatch()) != 0 {
		for _, g := range *lineItem.UnTagGRNMatch() {
			var inv_grn_value GenericType
			var revertQuantity float64
			if lineItem.Quantity() == 1 {
				revertQuantity = lineItem.Amount()
			} else {
				revertQuantity = lineItem.Quantity()
			}
			inv_grn_value.SetAlphaKey(lineItem.PoLine())
			inv_grn_value.SetBetaKey(revertQuantity)
			inv_grn_value.SetGammaKey(scanID)
			INV_GRNMAP[grn.GrnNumber()] = inv_grn_value
			break
		}
	}
	myLogger.Debugf(">>>>>>>>>>>>>>>>>>Printing INV_GRNMAP Values")
	for grnNum, values := range INV_GRNMAP {
		myLogger.Debugf("Map Values:", grnNum, values.AlphaKey(), values.BetaKey(), values.GammaKey())
	}

	myLogger.Debugf("Po number==============", po_num)

	invoice_po_grns := grn.GetGrnsByPO(stub, po_num)
	myLogger.Debugf("invoice_po_grns=-==============", invoice_po_grns)
	for _, grn := range invoice_po_grns {
		grn_generic := INV_GRNMAP[grn.GrnNumber()]
		myLogger.Debugf("grn_generic.AlphaKey()", grn_generic.AlphaKey(), grn_generic.AlphaKey())
		if grn_generic.AlphaKey() != 0 && grn_generic.BetaKey() != 0.0 {
			for idx, line := range *grn.LineItems() {
				myLogger.Debugf(" line.PoLineItemNumber============", line.PoLineItemNumber(), grn_generic.AlphaKey())
				if line.PoLineItemNumber() == grn_generic.AlphaKey() {
					myLogger.Debugf("Entered if condition==================")
					var residualQty float64
					revertQty := grn_generic.BetaKey()
					myLogger.Debugf(" residualQty Quantity available in map=================", revertQty)
					residualQty = line.ResidualQuantity()
					residualQty = residualQty + revertQty
					myLogger.Debugf("second residualQty Quantity available in map=================", revertQty)
					(*grn.LineItems())[idx].SetResidualQuantity(residualQty)
					myLogger.Debugf("Final residualQty Quantity available in map=================", (*grn.LineItems())[idx].ResidualQuantity())
				}
			}
			var newGrnInvLineItems []grn.InvLine
			for _, invLine := range *grn.InvLineItems() {
				if invLine.InvNumber() != grn_generic.GammaKey() {
					newGrnInvLineItems = append(newGrnInvLineItems, invLine)
				}
			}
			grn.SetInvLineItems(newGrnInvLineItems)
		}
		myLogger.Debugf("Updated GRN=====================", grn)
		contextObjPtr.UpdatedGRNs[grn.GrnNumber()] = grn
	}
	StoreGRNResiduals(stub)
	myLogger.Debugf(">>>>>>>>>>>EXIT Func<<<<<<<<<<<<")

}
*/
func BuyerDelegation(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (string, InvoiceStatus) {
	var invoice *Invoice = &(contextObjPtr.Invoice)
	var errStr string
	var invoicesBySupplier map[string]string
	invoicesBySupplier = make(map[string]string)
	var invoicesByBuyer map[string]string
	invoicesByBuyer = make(map[string]string)
	po, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, invoice.DcDocumentData.DcHeader.PoNumber, invoice.DcDocumentData.DcHeader.Client})

	if fetchErr != "" {
		return errStr, invStat
		// return 0, "ERROR parsing input PO in stage 01", invStat
	}
	//invoiceKey := invoice.BCIID + "~" + invoice.DcDocumentData.DcHeader.ScanID
	invoiceKey := invoice.DcDocumentData.DcHeader.ScanID + "~" + invoice.BCIID

	var buyerIdOrg = po.BuyerID

	GetInvoicesByBuyerIDFilter(stub, buyerIdOrg, invoiceKey)

	myLogger.Debugf("buyer id and email id===========", invStat.BuyerId)
	myLogger.Debugf("buyer id and email id===========", invStat.BuyerEmailId)
	po.BuyerID = invStat.BuyerId
	po.BuyerEmailId = invStat.BuyerEmailId
	myLogger.Debugf("Po before updation============", po)
	UpdatePO(stub, contextObjPtr, po)
	myLogger.Debugf("Po beAfterfore updation============", po)

	if invoice.DcDocumentData.DcHeader.VendorID != "" {
		if invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] != "" {
			invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] = invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] + "|" + invoiceKey
		} else {
			invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] = invoiceKey
		}
	}

	myLogger.Debugf("Buyer Id for updation==============", po.BuyerID)
	if po.BuyerID != "" {
		if invoicesByBuyer[po.BuyerID] != "" {
			invoicesByBuyer[po.BuyerID] = invoicesByBuyer[po.BuyerID] + "|" + invoiceKey
			myLogger.Debugf("nuyer Id updation in table", invoicesByBuyer[po.BuyerID])
		} else {
			invoicesByBuyer[po.BuyerID] = invoiceKey
		}
	}

	//Updating Details in supplier table
	for supplierId, val := range invoicesBySupplier {
		util.UpdateReferenceData(stub, util.TAB_INVOICE_BY_SUPPLIER, []string{supplierId}, val)
	}

	//Updating Details in Buyer table
	for buyerId, val := range invoicesByBuyer {
		myLogger.Debugf("Buyer Id=-====", buyerId, val)
		util.UpdateReferenceData(stub, util.TAB_INVOICE_BY_BUYER, []string{buyerId}, val)
	}
	return errStr, invStat
}

/*
func DynamicGRNUpdate(stub shim.ChaincodeStubInterface, invoiceRecArr string, scanID string, poNumber string) pb.Response {
	myLogger.Debugf("Entered grnDynamicRevert=====================", invoiceRecArr, scanID, poNumber)

	var invoices []Invoice

	var updatedGrn grn.GRN

	contextObjPtr.UpdatedGRNs = make(map[string]grn.GRN)

	err := json.Unmarshal([]byte(invoiceRecArr), &invoices)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input invoice array:", err)
	}
	myLogger.Debugf("Line=======================", invoiceRecArr)
	for _, invoice := range invoices {

		for _, line := range *invoice.DcLines() {

			myLogger.Debugf("Entered if condition=======================>", line)
			updatedGrn = grnDynamicUpdate(stub, line, scanID)
		}
	}

	return shim.Success(util.MarshalToBytes(updatedGrn))
	//	return shim.Success(nil)
}

func grnDynamicUpdate(stub shim.ChaincodeStubInterface, line DCLine, inv_no string) grn.GRN {

	var selectedGrns []grn.GRN
	var updatedGrn grn.GRN
	for _, g := range *line.TagGRNMatch() {

		myLogger.Debugf("g ---- > ", string(util.MarshalToBytes(g)))
		if line.Quantity() == 1 {
			UpdateGrnResidualQty(stub, g, line.UnitPrice(), inv_no, line.InvoiceLine())
		} else {
			UpdateGrnResidualQty(stub, g, line.Quantity(), inv_no, line.InvoiceLine())
		}

		selectedGrns = append(selectedGrns, g)
		myLogger.Debugf("grn updated", g)
		line.SetGrnMatch(selectedGrns)
		myLogger.Debugf("line.GRNMatch=========", line.GrnMatch())
		line.SetTagGRNMatch(nil)
		updatedGrn = g
	}
	return updatedGrn
}

func DynamicGRN(stub shim.ChaincodeStubInterface, invoiceRecArr string, scanID string, poNumber string) pb.Response {
	myLogger.Debugf("Entered grnDynamicRevert=====================", invoiceRecArr, scanID, poNumber)

	var invoices []Invoice

	var finalgrn grn.GRN

	err := json.Unmarshal([]byte(invoiceRecArr), &invoices)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input invoice array:", err)
	}
	myLogger.Debugf("Line=======================", invoiceRecArr)
	for _, invoice := range invoices {

		for _, line := range *invoice.DcLines() {

			myLogger.Debugf("Entered if condition=======================>", line)
			finalgrn = RevertGRNResidualQuantityDynamic(stub, line, scanID, poNumber)

		}
	}

	return shim.Success(util.MarshalToBytes(finalgrn))
}

func RevertGRNResidualQuantityDynamic(stub shim.ChaincodeStubInterface, lineItem DCLine, scanID string, po_num string) grn.GRN {
	var INV_GRNMAP map[string]GenericType
	INV_GRNMAP = make(map[string]GenericType)
	var finalgrn grn.GRN
	var grnnumber string
	myLogger.Debugf("lineItem.GrnMatch()===========", lineItem.GrnMatch())
	//var inv_no = invoice.InvoiceNumber()
	//var po_num = invoice.DcDocumentData.DcHeader.PoNumber
	if len(*lineItem.GrnMatch()) != 0 {
		for _, g := range *lineItem.GrnMatch() {
			myLogger.Debugf("Number of grns===================")
			var inv_grn_value GenericType
			var revertQuantity float64
			if lineItem.Quantity() == 1 {
				revertQuantity = lineItem.Amount()
			} else {
				revertQuantity = lineItem.Quantity()
			}
			inv_grn_value.SetAlphaKey(lineItem.PoLine())
			inv_grn_value.SetBetaKey(revertQuantity)
			inv_grn_value.SetGammaKey(scanID)
			INV_GRNMAP[grn.GrnNumber()] = inv_grn_value
			grnnumber = grn.GrnNumber()
			break
		}
	}

	myLogger.Debugf(">>>>>>>>>>>>>>>>>>Printing INV_GRNMAP Values")
	for grnNum, values := range INV_GRNMAP {
		myLogger.Debugf("Map Values:", grnNum, values.AlphaKey(), values.BetaKey(), values.GammaKey())
	}

	myLogger.Debugf("Po number==============", po_num)

	invoice_po_grns := grn.GetGrnsByPO(stub, po_num)
	myLogger.Debugf("invoice_po_grns=-==============", invoice_po_grns)

	for _, grn := range invoice_po_grns {
		if grn.GrnNumber() == grnnumber {
			finalgrn = grn
			grn_generic := INV_GRNMAP[finalgrn.GrnNumber()]
			myLogger.Debugf("INV_GRNMAP[grn.GrnNumber()].AlphaKey()", grn_generic.AlphaKey(), grn_generic.AlphaKey())
			if grn_generic.AlphaKey() != 0 && grn_generic.BetaKey() != 0.0 {
				for idx, line := range *finalgrn.LineItems() {
					myLogger.Debugf(" line.PoLineItemNumber============", line.PoLineItemNumber(), grn_generic.AlphaKey())
					if line.PoLineItemNumber() == grn_generic.AlphaKey() {
						myLogger.Debugf("Entered if condition==================")
						var residualQty float64
						revertQty := grn_generic.BetaKey()
						myLogger.Debugf(" residualQty Quantity available in map=================", revertQty)
						residualQty = line.ResidualQuantity()
						residualQty = residualQty + revertQty
						myLogger.Debugf("second residualQty Quantity available in map=================", revertQty)
						(*finalgrn.LineItems())[idx].SetResidualQuantity(residualQty)
						myLogger.Debugf("Final residualQty Quantity available in map=================", (*finalgrn.LineItems())[idx].ResidualQuantity())
					}
				}
				var newGrnInvLineItems []grn.InvLine
				for _, invLine := range *finalgrn.InvLineItems() {
					if invLine.InvNumber() != grn_generic.GammaKey() {
						newGrnInvLineItems = append(newGrnInvLineItems, invLine)
					}
				}
				finalgrn.SetInvLineItems(newGrnInvLineItems)
			}
			myLogger.Debugf("Updated GRN=====================", finalgrn)
			myLogger.Debugf("contextObjPtr.UpdatedGRNs==============Before ", contextObjPtr.UpdatedGRNs)

			//contextObjPtr.UpdatedGRNs[grn.GrnNumber()] = grn
			grn.AddGRN(stub, []string{finalgrn.BillOfLading(), finalgrn.GrnNumber()}, string(util.MarshalToBytes(finalgrn)))
		}
	}
	//StoreGRNResiduals(stub)
	myLogger.Debugf("Response Result:", finalgrn)
	//grnFromDB1,_:=grn.GetGRN(stub, []string{"","",grnnumber})

	return finalgrn
}
*/
func ForwardToOtherBuyer(stub shim.ChaincodeStubInterface, invoice *Invoice, invStat InvoiceStatus) (string, InvoiceStatus) {
	var errStr string
	var invoicesBySupplier map[string]string
	invoicesBySupplier = make(map[string]string)
	var invoicesByBuyer map[string]string
	invoicesByBuyer = make(map[string]string)
	//	po, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.PoNumber,invoice.DcDocumentData.DcHeader.ErpSystem})
	/*	po, fetchErr := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, invoice.DcDocumentData.DcHeader.PoNumber, invoice.DcDocumentData.DcHeader.Client})
		if fetchErr != "" {
			return errStr, invStat
			// return 0, "ERROR parsing input PO in stage 01", invStat
		}
	*/
	//	invoiceKey := invoice.BCIID + "~" + invoice.DcDocumentData.DcHeader.ScanID
	invoiceKey := invoice.DcDocumentData.DcHeader.ScanID + "~" + invoice.BCIID
	/*	var buyerIdOrg = po.BuyerID

		myLogger.Debugf("Old buyer ID==================", buyerIdOrg)
	*/
	//	GetInvoicesByBuyerIDFilter(stub,buyerIdOrg,invoiceKey)

	myLogger.Debugf("buyer id and email id===========", invStat.BuyerId)
	myLogger.Debugf("buyer id and email id===========", invStat.BuyerEmailId)
	//	po.BuyerID=invStat.BuyerId
	//	po.BuyerEmailId()=invStat.BuyerEmailId
	//	myLogger.Debugf("Po before updation============", po)
	//	UPdatePO(stub,po)
	//	myLogger.Debugf("Po beAfterfore updation============", po)

	if invoice.DcDocumentData.DcHeader.VendorID != "" {
		if invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] != "" {
			invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] = invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] + "|" + invoiceKey
		} else {
			invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] = invoiceKey
		}
	}

	//	myLogger.Debugf("Buyer Id for updation==============", po.BuyerID)
	if invStat.BuyerId != "" {
		if invoicesByBuyer[invStat.BuyerId] != "" {
			invoicesByBuyer[invStat.BuyerId] = invoicesByBuyer[invStat.BuyerId] + "|" + invoiceKey
			myLogger.Debugf("nuyer Id updation in table", invoicesByBuyer[invStat.BuyerId])
		} else {
			invoicesByBuyer[invStat.BuyerId] = invoiceKey
		}
	}

	//Updating Details in supplier table
	for supplierId, val := range invoicesBySupplier {
		util.UpdateReferenceData(stub, util.TAB_INVOICE_BY_SUPPLIER, []string{supplierId}, val)
	}

	//Updating Details in Buyer table
	for buyerId, val := range invoicesByBuyer {
		myLogger.Debugf("Buyer Id=-====", buyerId, val)
		util.UpdateReferenceData(stub, util.TAB_INVOICE_BY_BUYER, []string{invStat.BuyerId}, val)
	}
	return errStr, invStat
}

func ForwardToBuyer(stub shim.ChaincodeStubInterface, invoice *Invoice, invStat InvoiceStatus) InvoiceStatus {
	var invoicesByBuyer map[string]string
	invoicesByBuyer = make(map[string]string)
	invoiceKey := invoice.DcDocumentData.DcHeader.ScanID + "~" + invoice.BCIID
	if invStat.BuyerId != "" {
		if invoicesByBuyer[invStat.BuyerId] != "" {
			invoicesByBuyer[invStat.BuyerId] = invoicesByBuyer[invStat.BuyerId] + "|" + invoiceKey
			myLogger.Debugf("buyer Id updation in table", invoicesByBuyer[invStat.BuyerId])
		} else {
			invoicesByBuyer[invStat.BuyerId] = invoiceKey
		}
		//Updating Details in Buyer table
		for buyerId, val := range invoicesByBuyer {
			myLogger.Debugf("Buyer Id=-====", buyerId, val)
			util.UpdateReferenceData(stub, util.TAB_INVOICE_BY_BUYER, []string{invStat.BuyerId}, val)
		}

	}
	return invStat

}
