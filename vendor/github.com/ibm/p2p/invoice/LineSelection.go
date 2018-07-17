/*
   Copyright IBM Corp. 2018 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Santosh Penubothula
*/

package invoice

import (
	"encoding/json"
	"strings"

	"github.com/ibm/p2p/grn"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	"github.com/ibm/p2p/po"
)

//const lINE_SELECTION_NEXT_STEP = "st-budgetValidation-1"
const lINE_SELECTION_NEXT_STEP = "st-determinePOUnitPrice-1"

func PerformLineSelection(stub shim.ChaincodeStubInterface, contextObjPtr *Context) (int, string, InvoiceStatus) {

	var errStr string
	var invStat InvoiceStatus
	var totalACValue float64

	for _, dcLine := range contextObjPtr.Invoice.DcDocumentData.AdditionalLineItems {
		// if util.StrippedLowerCase(dcLine.PoNumber) == "ac" || util.StrippedLowerCase(dcLine.PoNumber) == "additionalcost" {
		// 	totalACValue += dcLine.Amount
		// 	continue
		// }
		totalACValue += dcLine.Amount
	}

	mPOxActivePOLines_NonActivePOLines := getPOxActivePOLines_NonActivePOLinesForInvoice(stub, &contextObjPtr.Invoice)

	var poNumbers []string
	singleLinePO := false
	for PONumber, POLines := range mPOxActivePOLines_NonActivePOLines {
		activePOLines := POLines[0]
		nonActivePOLines := POLines[1]
		if len(activePOLines)+len(nonActivePOLines) == 1 {
			myLogger.Debugf("LS : Single line PO with PONumber : ", PONumber)
			singleLinePO = true
		}
		if len(activePOLines) == 0 {
			myLogger.Debugf("LS : No active PO lines on PO : ", PONumber)
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_NO_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap2", AdditionalInfo{Type1: "PONumber", Value: PONumber})
			return 2, errStr, invStat
		}
		poNumbers = append(poNumbers, PONumber)
	}

	var totalInvoiceValue = contextObjPtr.Invoice.DcDocumentData.DcSwissHeader.TotalNet - totalACValue

	if len(contextObjPtr.Invoice.DcDocumentData.DcLines) == 0 {
		myLogger.Debugf("LS : Error : zero non AC invoice lines, cannot do line selection")
	}

	if len(poNumbers) == 1 && singleLinePO {
		if len(contextObjPtr.Invoice.DcDocumentData.DcLines) > 1 {
			myLogger.Debugf("LS : Single PO (having single line) Invoice with multiple DCLines")
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_NO_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap3", EMPTY_ADDITIONAL_INFO)
			return 2, errStr, invStat
		} else {
			// LINE SELECTION
			contextObjPtr.Invoice.DcDocumentData.DcLines[0].PoLine = mPOxActivePOLines_NonActivePOLines[poNumbers[0]][0][0].LineItemNumber
			contextObjPtr.StoreInvoice = true
			// LINE SELECTION
			invStat = UpdateInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", lINE_SELECTION_NEXT_STEP, EMPTY_ADDITIONAL_INFO)
			return 1, errStr, invStat
		}
	}

	if len(poNumbers) == 1 {
		grnsActivePOLines := getGRNSForActivePOLines(stub, &mPOxActivePOLines_NonActivePOLines)
		code, errStrSM, invStatSM := grnBasedInvoiceReconstruction(stub, contextObjPtr, &grnsActivePOLines, totalInvoiceValue)
		if code != 0 {
			return code, errStrSM, invStatSM
		}
	} else {
		myLogger.Debugf("LS : GRN based invoice reconstruction cannot be done as invoice has more than one PO")
	}

	mPOxPOLinesSelectedMaterial := make(map[string]map[int64]bool)
	dcLinesSelected := make(map[int]bool)
	atleastOneLineSelected := false
	//Material no. based match
	for i, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {

		var lineItemNumberSelected int64 = -1
		if dcLine.MatNumber != "" {
			for _, activePOLine := range mPOxActivePOLines_NonActivePOLines[dcLine.PoNumber][0] {
				if util.EqualsIgnoreCase(dcLine.MatNumber, activePOLine.MaterialNumber) {
					if lineItemNumberSelected == -1 {
						lineItemNumberSelected = activePOLine.LineItemNumber
					} else {
						//3.9.38
						myLogger.Debugf("LS : Multiple PO active lines with same Material number")
						if atleastOneLineSelected {
							invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_FEW_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap4", AdditionalInfo{Type1: "PONumber", Value: dcLine.PoNumber})
						} else {
							invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_NO_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap4", AdditionalInfo{Type1: "PONumber", Value: dcLine.PoNumber})
						}
						return 2, errStr, invStat
					}
				}
			}
		}
		if lineItemNumberSelected != -1 {
			//Material number of PO active and non-active lines matching check
			/*
				for _, nonActivePOLine := range mPOxActivePOLines_NonActivePOLines[dcLine.PoNumber][1] {
					if dcLine.MatNumber == nonActivePOLine.MaterialNumber {
						//3.9.38
						myLogger.Debugf("LS : Material number of PO active and non-active lines matching")
						if atleastOneLineSelected {
							invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_FEW_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap4", AdditionalInfo{Type1: "PONumber", Value: dcLine.PoNumber})
						} else {
							invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_NO_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap4", AdditionalInfo{Type1: "PONumber", Value: dcLine.PoNumber})
						}
						return 2, errStr, invStat
					}
				}
			*/
			_, present := mPOxPOLinesSelectedMaterial[dcLine.PoNumber]
			if present {
				_, present := mPOxPOLinesSelectedMaterial[dcLine.PoNumber][lineItemNumberSelected]
				if present {
					myLogger.Debugf("LS : Material Number based line selection : Same POLine selected for multiple invoice lines")
					//3.9.31
					if atleastOneLineSelected {
						invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_FEW_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap5", EMPTY_ADDITIONAL_INFO)
					} else {
						invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_NO_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap5", EMPTY_ADDITIONAL_INFO)
					}
					return 2, errStr, invStat
				}
				mPOxPOLinesSelectedMaterial[dcLine.PoNumber][lineItemNumberSelected] = true
			} else {
				mPOxPOLinesSelectedMaterial[dcLine.PoNumber] = make(map[int64]bool)
				mPOxPOLinesSelectedMaterial[dcLine.PoNumber][lineItemNumberSelected] = true
			}
			//LINE SELECTION
			contextObjPtr.Invoice.DcDocumentData.DcLines[i].PoLine = lineItemNumberSelected
			contextObjPtr.StoreInvoice = true
			atleastOneLineSelected = true
			dcLinesSelected[i] = true
			//LINE SELECTION
		}
	}

	//Description based line selection

	if len(contextObjPtr.Invoice.DcDocumentData.DcLines) != len(dcLinesSelected) {

		myLogger.Debugf("LS : Line selection for all lines didn't happen")
		myLogger.Debugf("LS : Trying Description based match for remaining")
		mPOxPOLinesSelectedDescription := make(map[string]map[int64]bool)

		for i, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {

			_, present := dcLinesSelected[i]
			if !present {
				var lineItemNumberSelected int64 = -1
				if dcLine.Description != "" {
					for _, activePOLine := range mPOxActivePOLines_NonActivePOLines[dcLine.PoNumber][0] {
						_, present := mPOxPOLinesSelectedMaterial[dcLine.PoNumber]
						if present {
							_, present := mPOxPOLinesSelectedMaterial[dcLine.PoNumber][activePOLine.LineItemNumber]
							if present {
								continue
							}
						}
						if util.EqualsIgnoreCase(dcLine.Description, activePOLine.Description) {
							if lineItemNumberSelected == -1 {
								lineItemNumberSelected = activePOLine.LineItemNumber
							} else {
								//3.9.37
								myLogger.Debugf("LS : Multiple PO active lines with same Description")
								if atleastOneLineSelected {
									invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_FEW_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap6", AdditionalInfo{Type1: "PONumber", Value: dcLine.PoNumber})
								} else {
									invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_NO_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap6", AdditionalInfo{Type1: "PONumber", Value: dcLine.PoNumber})
								}
								return 2, errStr, invStat
							}
						}
					}
				}
				if lineItemNumberSelected != -1 {
					//Description of PO active and non-active lines matching check
					/*
						for _, nonActivePOLine := range mPOxActivePOLines_NonActivePOLines[dcLine.PoNumber][1] {
							if strings.ToLower(dcLine.Description) == strings.ToLower(nonActivePOLine.Description) {
								//3.9.37
								myLogger.Debugf("LS : Description of PO active and non-active lines matching")
								if atleastOneLineSelected {
									invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_FEW_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap6", AdditionalInfo{Type1: "PONumber", Value: dcLine.PoNumber})
								} else {
									invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_NO_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap6", AdditionalInfo{Type1: "PONumber", Value: dcLine.PoNumber})
								}
								return 2, errStr, invStat
							}
						}
					*/
					for _, activePOLine := range mPOxActivePOLines_NonActivePOLines[dcLine.PoNumber][0] {
						if util.EqualsIgnoreCase(dcLine.Description, activePOLine.Description) {
							//3.9.37
							myLogger.Debugf("LS : Multiple PO active lines with same Description")
							if atleastOneLineSelected {
								invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_FEW_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap6", AdditionalInfo{Type1: "PONumber", Value: dcLine.PoNumber})
							} else {
								invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_NO_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap6", AdditionalInfo{Type1: "PONumber", Value: dcLine.PoNumber})
							}
							return 2, errStr, invStat
						}
					}
					_, present := mPOxPOLinesSelectedDescription[dcLine.PoNumber]
					if present {
						_, present := mPOxPOLinesSelectedDescription[dcLine.PoNumber][lineItemNumberSelected]
						if present {
							myLogger.Debugf("LS : Description based line selection : Same POLine selected for multiple invoice lines")
							//3.9.31
							if atleastOneLineSelected {
								invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_FEW_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap5", EMPTY_ADDITIONAL_INFO)
							} else {
								invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_NO_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap5", EMPTY_ADDITIONAL_INFO)
							}
							return 2, errStr, invStat
						} else {
							mPOxPOLinesSelectedDescription[dcLine.PoNumber][lineItemNumberSelected] = true
						}
					} else {
						mPOxPOLinesSelectedDescription[dcLine.PoNumber] = make(map[int64]bool)
						mPOxPOLinesSelectedDescription[dcLine.PoNumber][lineItemNumberSelected] = true
					}
					//LINE SELECTION
					contextObjPtr.Invoice.DcDocumentData.DcLines[i].PoLine = lineItemNumberSelected
					contextObjPtr.StoreInvoice = true
					atleastOneLineSelected = true
					dcLinesSelected[i] = true
					//LINE SELECTION
				}
			}
		}
	}

	if len(contextObjPtr.Invoice.DcDocumentData.DcLines) != len(dcLinesSelected) {
		//UOM check followed by Quantity and UP based match
	}

	if len(contextObjPtr.Invoice.DcDocumentData.DcLines) != len(dcLinesSelected) {
		if atleastOneLineSelected {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_FEW_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap7", EMPTY_ADDITIONAL_INFO)
		} else {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_NO_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap7", EMPTY_ADDITIONAL_INFO)
		}
		return 2, errStr, invStat
	}

	mPOxPOLinesSelected := make(map[string]map[int64]bool)

	for _, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {

		_, present := mPOxPOLinesSelected[dcLine.PoNumber]
		if present {
			_, present := mPOxPOLinesSelected[dcLine.PoNumber][dcLine.PoLine]
			if present {
				myLogger.Debugf("LS : Overall line selection : Same POLine selected for multiple invoice lines")
				//3.9.31
				if atleastOneLineSelected {
					invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_FEW_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap5", EMPTY_ADDITIONAL_INFO)
				} else {
					invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, INV_STATUS_PENDING_AP, INV_RS_LS_NO_LINES_SELECTED, invStat.Comments, "st-LineSelection-ap5", EMPTY_ADDITIONAL_INFO)
				}
				return 2, errStr, invStat
			} else {
				mPOxPOLinesSelected[dcLine.PoNumber][dcLine.PoLine] = true
			}
		} else {
			mPOxPOLinesSelected[dcLine.PoNumber] = make(map[int64]bool)
			mPOxPOLinesSelected[dcLine.PoNumber][dcLine.PoLine] = true
		}
	}

	invStat = UpdateInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", lINE_SELECTION_NEXT_STEP, EMPTY_ADDITIONAL_INFO)
	return 1, errStr, invStat
}

func grnBasedInvoiceReconstruction(stub shim.ChaincodeStubInterface, contextObjPtr *Context, grnsActivePOLines *[]grn.GRN, totalInvoiceValue float64) (int, string, InvoiceStatus) {

	var errStr string
	var invStat InvoiceStatus
	//3.9.9
	var totalMatchedGRNValue float64
	scanIDSuffix := getScanIDSuffix(contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID)

	grnsReferenceMatched := []grn.GRN{}

	for _, GRN := range *grnsActivePOLines {

		currGRNmatched := false
		if !util.EqualsIgnoreCase(GRN.RefDocNumber, contextObjPtr.Invoice.DcDocumentData.DcHeader.InvoiceNumber) && !util.EqualsIgnoreCase(GRN.RefDocNumber, scanIDSuffix) {
			for _, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {
				if util.EqualsIgnoreCase(dcLine.DeliveryNote, GRN.RefDocNumber) {
					currGRNmatched = true
					break
				}
			}
		} else {
			currGRNmatched = true
		}
		if !currGRNmatched {
			myLogger.Debugf("LS : Reference doc no. based match failed for grn with GRNNumber : ", GRN.GrnNumber)
		} else {
			totalMatchedGRNValue += GRN.GRNValue
			grnsReferenceMatched = append(grnsReferenceMatched, GRN)
		}
	}

	myLogger.Debugf("LS : totalMatchedGRNValue : ", totalMatchedGRNValue, " Invoice net value : ", totalInvoiceValue)

	if totalMatchedGRNValue == totalInvoiceValue {
		//3.9.9 yes 3.9.42
		//Invoice Number and ScanID suffix based match
		allGRNSMatched := true
		totalMatchedGRNValue = 0
		for _, GRN := range grnsReferenceMatched {
			if !util.EqualsIgnoreCase(GRN.RefDocNumber, contextObjPtr.Invoice.DcDocumentData.DcHeader.InvoiceNumber) && !util.EqualsIgnoreCase(GRN.RefDocNumber, scanIDSuffix) {
				myLogger.Debugf("LS : Invoice no. or ScanID suffix based match failed for grn with GRNNumber : ", GRN.GrnNumber)
				allGRNSMatched = false
				break
			}
			totalMatchedGRNValue += GRN.GRNValue
		}

		myLogger.Debugf("LS : Invoice no. and ScanID suffix based match :: allGRNSMatched : ", allGRNSMatched, " totalMatchedGRNValue : ", totalMatchedGRNValue, " Invoice net value : ", totalInvoiceValue)

		//3.9.42.{1,2}
		if allGRNSMatched && totalMatchedGRNValue == totalInvoiceValue {
			//3.9.42.3
			//RECONSTRUCT INVOICE AS PER GRN : Replace all DCLines with GRN lines
			newDCLines := []DCLine{}
			taxPercentage := -1
			deliveryNote := "=1"

			for idx := range contextObjPtr.Invoice.DcDocumentData.DcLines {
				dcLine := contextObjPtr.Invoice.DcDocumentData.DcLines[idx]
				if deliveryNote == "=1" {
					deliveryNote = dcLine.DeliveryNote
				} else if !util.EqualsIgnoreCase(dcLine.DeliveryNote, deliveryNote) {
					deliveryNote = ">1"
				}
				if taxPercentage == -1 {
					taxPercentage = int(dcLine.TaxAmount * 10000 / dcLine.Amount)
				} else if taxPercentage != int(dcLine.TaxAmount*10000/dcLine.Amount) {
					taxPercentage = -2
				}
			}

			for _, GRN := range grnsReferenceMatched {

				newDCLine := DCLine{}
				newDCLine.PoNumber = GRN.PONumber
				newDCLine.PoLine = GRN.POItemNumber
				newDCLine.Amount = GRN.GRNValue
				newDCLine.Quantity = GRN.Quantity
				if taxPercentage < 0 {
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
				if deliveryNote != ">1" {
					newDCLine.DeliveryNote = deliveryNote
				}
				newDCLines = append(newDCLines, newDCLine)
			}
			//3.9.43
			//DO LINE SELECTION AS PER GRN's POLINE
			contextObjPtr.Invoice.DcDocumentData.DcLines = newDCLines
			contextObjPtr.Invoice.D_LinesReconstructed = "LineSelection"
			contextObjPtr.StoreInvoice = true
			myLogger.Debugf("LS : Invoice no. and ScanID suffix, GRN based reconstruction and line selection")
			invStat = UpdateInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", lINE_SELECTION_NEXT_STEP, EMPTY_ADDITIONAL_INFO)
			return 1, errStr, invStat
		}
		//3.9.42.4
		DeliveryNoteSet := make(map[string]bool)
		for _, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {

			cDeliveryNote := dcLine.DeliveryNote
			if cDeliveryNote == "" {
				myLogger.Debugf("LS : Empty Delivery note on invoice DCLine")
				continue
			}
			_, present := DeliveryNoteSet[cDeliveryNote]
			if !present {
				DeliveryNoteSet[cDeliveryNote] = true
			}
		}
		var grnsMatchDeliveryNoteSet []grn.GRN
		totalMatchedGRNValue = 0
		grnForAllDeliveryNote := true
		for cDeliveryNote := range DeliveryNoteSet {
			grnForCurrDeliveryNote := false
			for _, GRN := range grnsReferenceMatched {
				if util.EqualsIgnoreCase(GRN.RefDocNumber, cDeliveryNote) {
					grnsMatchDeliveryNoteSet = append(grnsMatchDeliveryNoteSet, GRN)
					totalMatchedGRNValue += GRN.GRNValue
					grnForCurrDeliveryNote = true
				}
			}
			if !grnForCurrDeliveryNote {
				myLogger.Debugf("LS : GRN doesn't exist for DCline with delivery note : ", cDeliveryNote)
				grnForAllDeliveryNote = false
			}
		}

		myLogger.Debugf("LS : grnForAllDeliveryNote : ", grnForAllDeliveryNote, " totalGRNValue : ", totalMatchedGRNValue, " Invoice net value : ", totalInvoiceValue)

		if grnForAllDeliveryNote && totalMatchedGRNValue == totalInvoiceValue {

			//3.9.42.{8, 12}
			//RECONSTRUCT INVOICE AS PER GRN : Replace all DCLines with GRN lines

			newDCLines := []DCLine{}
			taxPercentage := -1

			for _, dcLine := range contextObjPtr.Invoice.DcDocumentData.DcLines {

				if taxPercentage == -1 {
					taxPercentage = int(dcLine.TaxAmount * 10000 / dcLine.Amount)
				} else if taxPercentage != int(dcLine.TaxAmount*10000/dcLine.Amount) {
					taxPercentage = -2
				}
			}

			for _, GRN := range grnsMatchDeliveryNoteSet {

				newDCLine := DCLine{}
				newDCLine.PoNumber = GRN.PONumber
				newDCLine.PoLine = GRN.POItemNumber
				newDCLine.Amount = GRN.GRNValue
				newDCLine.Quantity = GRN.Quantity
				if taxPercentage < 0 {
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
				newDCLine.DeliveryNote = GRN.RefDocNumber
				newDCLines = append(newDCLines, newDCLine)

			}
			//3.9.43
			//DO LINE SELECTION AS PER GRN's POLINE
			contextObjPtr.Invoice.DcDocumentData.DcLines = newDCLines
			contextObjPtr.Invoice.D_LinesReconstructed = "LineSelection"
			contextObjPtr.StoreInvoice = true
			myLogger.Debugf("LS : DCline delivery note no., GRN based reconstruction and line selection")
			invStat = UpdateInvoiceStatus(stub, contextObjPtr, contextObjPtr.Invoice.BCIID, contextObjPtr.Invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", lINE_SELECTION_NEXT_STEP, EMPTY_ADDITIONAL_INFO)
			return 1, errStr, invStat
		} else {
			myLogger.Debugf("LS : GRN based reconstruction and line selection failed")
			return 0, errStr, invStat
		}
	}
	myLogger.Debugf("LS : GRN based reconstruction and line selection failed")
	return 0, errStr, invStat
}

func getScanIDSuffix(scanID string) string {
	tokens := strings.Split(scanID, "_")
	return tokens[1]
}

func getPOxActivePOLines_NonActivePOLinesForInvoice(stub shim.ChaincodeStubInterface, invoice *Invoice) map[string][][]po.POLineItem {

	mPOxPOLines := make(map[string][][]po.POLineItem)
	erpSystem := invoice.DcDocumentData.DcHeader.ErpSystem

	for _, dcLine := range invoice.DcDocumentData.DcLines {

		cPONumber := dcLine.PoNumber
		_, present := mPOxPOLines[cPONumber]
		if !present {
			mPOxPOLines[cPONumber] = make([][]po.POLineItem, 2)
			poLineItemsRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO_LINEITEMS, PrimaryKeys: []string{erpSystem, cPONumber}, Data: ""}.GetAll()
			for _, poLineItemRow := range poLineItemsRec {
				var currentPOLineItem po.POLineItem
				json.Unmarshal([]byte(poLineItemRow), &currentPOLineItem)
				if currentPOLineItem.PoStatus == "ACTIVE" {
					mPOxPOLines[cPONumber][0] = append(mPOxPOLines[cPONumber][0], currentPOLineItem)
				} else {
					mPOxPOLines[cPONumber][1] = append(mPOxPOLines[cPONumber][1], currentPOLineItem)
				}
			}
		}
	}
	return mPOxPOLines
}

func getGRNSForActivePOLines(stub shim.ChaincodeStubInterface, mPOxActivePOLines_NonActivePOLines *map[string][][]po.POLineItem) []grn.GRN {
	var grns []grn.GRN
	for _, POLines := range *mPOxActivePOLines_NonActivePOLines {
		for _, POLine := range POLines[0] {
			grnRecords, _ := db.TableStruct{Stub: stub, TableName: util.TAB_GRN, PrimaryKeys: []string{POLine.ERPSystem, POLine.PONumber, util.GetStringFromInt(POLine.LineItemNumber)}, Data: ""}.GetAll()
			for _, grnRow := range grnRecords {
				var currentGRN grn.GRN
				json.Unmarshal([]byte(grnRow), &currentGRN)
				grns = append(grns, currentGRN)
			}
		}
	}
	return grns
}

func LineSelection_IBM_AP_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {

	var errStr string
	if invStat.Status == FWD_TO_BUYER {
		invStat = ForwardToBuyer(stub, &contextObjPtr.Invoice, invStat)
		tokens := strings.Split(invStat.InternalStatus, "ap")
		newInternalStatus := "st-LineSelection-buy" + tokens[len(tokens)-1]
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, newInternalStatus, EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == REJECT {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-LineSelection-end", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == PROCESS_OUTSIDE_BC {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, COMPLETED_OUTSIDE_BC, invStat.ReasonCode, invStat.Comments, "st-LineSelection-end", EMPTY_ADDITIONAL_INFO)
	} else if invStat.Status == "" || invStat.Status == CONTINUE {
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, CONTINUE, invStat.ReasonCode, invStat.Comments, lINE_SELECTION_NEXT_STEP, EMPTY_ADDITIONAL_INFO)
	}
	return 2, errStr, invStat
}

func LineSelection_Buyer_Action(stub shim.ChaincodeStubInterface, contextObjPtr *Context, invStat InvoiceStatus) (int, string, InvoiceStatus) {

	// UI STATUS TO BE HANDLED
	// 1) AWAITING BUYER ACTION HOLD INVOICE
	var errStr string
	if invStat.Status == FWD_TO_OTHER_BUYER {

		errStr, invStat = ForwardToOtherBuyer(stub, &contextObjPtr.Invoice, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_BUYER, invStat.ReasonCode, invStat.Comments, invStat.InternalStatus, EMPTY_ADDITIONAL_INFO)
		}
	} else if invStat.Status == REJECT {

		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, invStat.ReasonCode, invStat.Comments, "st-LineSelection-end", EMPTY_ADDITIONAL_INFO)

	} else if invStat.Status == RETURN_TO_AP {

		tokens := strings.Split(invStat.InternalStatus, "buy")
		newInternalStatus := "st-LineSelection-ap" + tokens[len(tokens)-1]
		invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_PENDING_AP, invStat.ReasonCode, invStat.Comments, newInternalStatus, EMPTY_ADDITIONAL_INFO)

	} else if invStat.Status == BUYER_DELEGATION {

		errStr, invStat = BuyerDelegation(stub, contextObjPtr, invStat)
		if errStr == "" {
			invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, invStat.Status, invStat.ReasonCode, invStat.Comments, invStat.InternalStatus, EMPTY_ADDITIONAL_INFO)
		}
	} else if invStat.Status == ALT_PO {

		// _, fetchErr := po.GetPO(stub, []string{contextObjPtr.Invoice.DcDocumentData.DcHeader.PoNumber, contextObjPtr.Invoice.DcDocumentData.DcHeader.ErpSystem, contextObjPtr.Invoice.DcDocumentData.DcHeader.Client})
		// if fetchErr != "" {
		// 	invStat, errStr = SetInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, INV_STATUS_REJECTED, INV_RS_INVALID_PO, "", "st-LineSelection-end", EMPTY_ADDITIONAL_INFO)
		// 	return 2, errStr, invStat
		// }
		//	myLogger.Debugf("LS : entered alternate PO loop=====================", po)
		//	PoBudgetRevert(stub, contextObjPtr.Invoice, po)
		invStat = UpdateInvoiceStatus(stub, contextObjPtr, invStat.BciId, invStat.ScanID, ALT_PO, invStat.ReasonCode, invStat.Comments, "ST0000", EMPTY_ADDITIONAL_INFO)
		return 1, errStr, invStat
	}
	return 2, errStr, invStat
}
