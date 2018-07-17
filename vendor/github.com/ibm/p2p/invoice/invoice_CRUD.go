/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package invoice

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	"github.com/ibm/p2p/grn"
	"github.com/ibm/p2p/po"
	"github.com/ibm/p2p/vmd"
	"github.com/ibm/pme"
	logging "github.com/op/go-logging"
	//"github.com/ibm/p2p/companyCode"
	"fmt"
	"sort"
	"strconv"
)

type DetailedInvoice struct {
	Invoice       Invoice         `json:"Invoice"`
	InvoiceStatus []InvoiceStatus `json:"Invoice_Status"`
}

func RemoveSpecialCharAndLeadingZeros(s string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")

	processedString := reg.ReplaceAllString(s, "")

	processedString = strings.TrimLeft(processedString, "0")
	return processedString
}

type Invoice struct {
	DcDocumentData       DCDocumentData `json:"DCDocumentData"`
	BCIID                string         `json:"BciId"`
	Version              string         `json:"Version"`
	D_LinesReconstructed string
}

type DCDocumentData struct {
	DcLines             []DCLine      `json:"DC_Lines"`
	Xmlns               string        `json:"xmlns"`
	DcSwissLine         DCSwissLine   `json:"DC_Swiss_Line"`
	DcHeader            DCHeader      `json:"DC_Header"`
	DcSwissHeader       DCSwissHeader `json:"DC_Swiss_Header"`
	Comments            []string      `json:"Comments"`
	AdditionalLineItems []DCLine      `json:"Additional_Line_Items"`
}

type DCSwissLine struct {
	DescriptionOfSupply string `json:"Description_Of_Supply"`
}

type DCLine struct {
	LineID              string    `json:"Line_ID"`
	ScanID              string    `json:"Scan_ID"`
	Description         string    `json:"Description"`
	Amount              float64   `json:"Amount"`
	GlAccount           string    `json:"GLAccount"`
	InvoiceLine         int64     `json:"Invoice_Line"`
	Quantity            float64   `json:"Quantity"`
	InternalOrder       string    `json:"InternalOrder"`
	TaxAmount           float64   `json:"Tax_Amount"`
	TaxPercent          float64   `json:"Tax_Percent"`
	CostCenter          string    `json:"CostCenter"`
	PoRelease           int64     `json:"PO_Release"`
	DeliveryNote        string    `json:"Delivery_Note"`
	MatNumber           string    `json:"MAT_NUMBER"`
	TaxCode             string    `json:"Tax_Code"`
	UnitPrice           float64   `json:"Unit_Price"`
	PoLine              int64     `json:"PO_Line_Number"`
	PoNumber            string    `json:"PO_Number"`
	GrnMatch            []grn.GRN `json:"Grn_Match"`
	FinalPrice          string    `json:"Final_Price"`
	OrgInvoiceLine      string    `json:"Org_Invoice_Line"`
	D_postFacto         bool      `json:"D_postFacto"`
	D_currencyMissMatch bool      `json:"D_currencyMissMatch"`
	D_outOfBudget       bool      `json:"D_outOfBudget"`

	TagGRNMatch   []grn.GRN `json:"Tag_GRN_Match"`
	UnTagGRNMatch []grn.GRN `json:"UnTag_GRN_Match"`
}

type DCLineHistory struct {
	LineHistoryID string  `json:"Line_History_ID"`
	BciID         string  `json:"BciId"`
	InvoiceNumber string  `json:"Invoice_Num"`
	LineID        string  `json:"Line_ID"`
	Description   string  `json:"Description"`
	Amount        float64 `json:"Amount"`
	InvoiceLine   int64   `json:"Invoice_Line"`
	Quantity      float64 `json:"Quantity"`
	TaxAmount     float64 `json:"Tax_Amount"`
	TaxPercent    float64 `json:"Tax_Percent"`
	TaxCode       string  `json:"Tax_Code"`
	UnitPrice     float64 `json:"Unit_Price"`
	PoLine        int64   `json:"Po_Line"`

	Time    util.BCDate `json:"Time"`
	Message string      `json:"Message"`
	ScanID  string      `json:"Scan_ID"`
}

type DCHeader struct {
	InvoiceNumber       string      `json:"Invoice_Num"`
	Description         string      `json:"Description"`
	PoNumber            string      `json:"PO_Number"`
	Ob10Link            string      `json:"OB10_Link"`
	TotalAmount         float64     `json:"Total_Amount"`
	IrpfValue           float64     `json:"IRPF_Value"`
	BankCode            string      `json:"Bank_Code"`
	ExchangeRate        float64     `json:"Exch_Rate"`
	ImportStatus        string      `json:"Import_Status"`
	VatLocalAmount      float64     `json:"VAT_Local_Amount"`
	BuyerVATCode        string      `json:"Buyer_VAT_Code"`
	VendorVATCode       string      `json:"Vendor_VAT_Code"`
	PorReference        string      `json:"POR_Reference"`
	TaxAmount           float64     `json:"Tax_Amount"`
	OtherInformation    string      `json:"Other_Information"`
	BankAccount         string      `json:"Bank_Account"`
	PaymentTerm         string      `json:"Payment_Term"`
	CompanyCode         string      `json:"Company_Code"`
	ImageFile           string      `json:"Image_File"`
	BankOGM             string      `json:"Bank_OGM"`
	DiscountAmount      float64     `json:"Discount_Amount"`
	PlantCode           string      `json:"Plant_Code"`
	ScanDate            util.BCDate `json:"Scan_Date"`
	ScanID              string      `json:"Scan_ID"`
	DueDate             util.BCDate `json:"Due_Date"`
	InvoiceException    string      `json:"Invoice_Exception"`
	VendorID            string      `json:"Vendor_ID"`
	PaymentReference    string      `json:"Payment_Reference"`
	InvType             string      `json:"Inv_Type"`
	ImportDate          util.BCDate `json:"Import_Date"`
	TaxReportingCountry string      `json:"TaxReportingCountry"`
	DiscountDate        util.BCDate `json:"Discount_Date"`
	IndexerID           string      `json:"Indexer_Id"`
	AttentionOf         string      `json:"AttentionOf"`
	DocDate             util.BCDate `json:"Doc_Date"`
	NumberOfImages      int64       `json:"Nr_Of_Images"`
	ErpRefNumber        string      `json:"ERP_Ref_Number"`
	CurrencyCode        string      `json:"Currency_Code"`
	DiscountPercent     float64     `json:"Discount_Percent"`
	VatRegion           string      `json:"VATRegion"`
	TaxCode             string      `json:"Tax_Code"`
	RevisedTotalAmount  float64     `json:"Revised_Total_Amount"`
	DocType             string      `json:"doctype"`
	DocSource           string      `json:"docsource"`

	ErpSystem           string `json:"ErpSystem"`
	Client              string `json:"Client"`
	D_postFacto         bool   `json:"D_postFacto"`
	D_currencyMissMatch bool   `json:"D_currencyMissMatch"`
}

type DCSwissHeader struct {
	SupplierName      string `json:"Supplier_Name"`
	SupplierAddress1  string `json:"Supplier_Address1"`
	SupplierAddress2  string `json:"Supplier_Address2"`
	SupplierAddress3  string `json:"Supplier_Address3"`
	SupplierAddress4  string `json:"Supplier_Address4"`
	SupplierAddress5  string `json:"Supplier_Address5"`
	SupplierAddress6  string `json:"Supplier_Address6"`
	SupplierAddress7  string `json:"Supplier_Address7"`
	SupplierAddress8  string `json:"Supplier_Address8"`
	SupplierAddress9  string `json:"Supplier_Address9"`
	SupplierAddress10 string `json:"Supplier_Address10"`

	SupplierFiscalName      string `json:"Supplier_Fiscal_Name"`
	SupplierFiscalAddress1  string `json:"Supplier_Fiscal_Address1"`
	SupplierFiscalAddress2  string `json:"Supplier_Fiscal_Address2"`
	SupplierFiscalAddress3  string `json:"Supplier_Fiscal_Address3"`
	SupplierFiscalAddress4  string `json:"Supplier_Fiscal_Address4"`
	SupplierFiscalAddress5  string `json:"Supplier_Fiscal_Address5"`
	SupplierFiscalAddress6  string `json:"Supplier_Fiscal_Address6"`
	SupplierFiscalAddress7  string `json:"Supplier_Fiscal_Address7"`
	SupplierFiscalAddress8  string `json:"Supplier_Fiscal_Address8"`
	SupplierFiscalAddress9  string `json:"Supplier_Fiscal_Address9"`
	SupplierFiscalAddress10 string `json:"Supplier_Fiscal_Address10"`

	BuyerName      string `json:"Buyer_Name"`
	BuyerAddress1  string `json:"Buyer_Address1"`
	BuyerAddress2  string `json:"Buyer_Address2"`
	BuyerAddress3  string `json:"Buyer_Address3"`
	BuyerAddress4  string `json:"Buyer_Address4"`
	BuyerAddress5  string `json:"Buyer_Address5"`
	BuyerAddress6  string `json:"Buyer_Address6"`
	BuyerAddress7  string `json:"Buyer_Address7"`
	BuyerAddress8  string `json:"Buyer_Address8"`
	BuyerAddress9  string `json:"Buyer_Address9"`
	BuyerAddress10 string `json:"Buyer_Address10"`

	TotalNet          float64     `json:"Total_Net"`
	DateOfSupply      util.BCDate `json:"Date_Of_Supply"`
	Ship_To_Address   string      `json:"Ship_To_Address"`
	Ship_From_Address string      `json:"Ship_From_Address"`
}

type InvoiceStatus struct {
	UserId string `json:"User_Id"`
	BciId  string `json:"Bci_Id"`
	//InvoiceNumber       string         `json:"Invoice_Number"`
	Status              string         `json:"Status"`
	ReasonCode          string         `json:"Reason_Code"`
	Comments            string         `json:"Comments"`
	InternalStatus      string         `json:"Internal_Status"`
	InternalDescription string         `json:"Internal_Description"`
	Time                util.BCDate    `json:"Time"`
	ProcessMapStep      string         `json:"Process_Map_Step"`
	AdditionalInfo      AdditionalInfo `json:"Additional_Info"`
	BuyerId             string         `json:"buyerId"`
	BuyerEmailId        string         `json:"buyerEmailId"`
	ScanID              string         `json:"Scan_ID"`
	TurnAroundTime      util.BCDate    `json:"Turn_Around_Time"`
}

type AdditionalInfo struct {
	Type1 string `json:"Type1"`
	Value string `json:"Value"`
}

func GetAdditionalInfo(t string, v string) *AdditionalInfo {
	var AddInfo AdditionalInfo
	AddInfo.Type1 = t
	AddInfo.Value = v
	return &AddInfo
}

type BCIRejectedInvoice struct {
	Invoice    Invoice `json:"Invoice"`
	Status     string  `json:"Status"`
	ReasonCode string  `json:"ReasonCode"`
}

type InvoiceReminderEmail struct {
	InvoiceNum string      `json:"Invoice_Num"`
	BciID      string      `json:"BciId"`
	Status     string      `json:"Status"`
	ReasonCode string      `json:"Reason_Code"`
	NoOfEmails int64       `json:"No_Of_Emails"`
	Time       util.BCDate `json:"Time"`
}

type TurnAroundDetails struct {
	Zeroday           int64 `json:"zeroday"`
	MinusOneDays      int64 `json:"minusOneDays"`
	MinusTwoDays      int64 `json:"minusTwoDays"`
	MinusthreeDays    int64 `json:"minusthreeDays"`
	OneDay            int64 `json:"oneDay"`
	TwoDays           int64 `json:"twoDays"`
	ThreeDays         int64 `json:"threeDays"`
	ThreeDaysAndAbove int64 `json:"ThreeDaysAndAbove"`
}

type DueDateDetails struct {
	DueToday     int64 `json:"dueToday"`
	DueInOneDay  int64 `json:"dueInOneDay"`
	PastDue      int64 `json:"pastDue"`
	TwoToFive    int64 `json:"twoToFive"`
	MoreThanFive int64 `json:"moreThanFive"`
}

type InvoicesToReSubmit struct {
	InvoicesToSubmit []InvoiceSubmitRequest
}

var myLogger = logging.MustGetLogger("Procure-To-Pay Invoice")
var EMPTY_ADDITIONAL_INFO AdditionalInfo
var InternalStatusProcessing string

func AddRejectedInvoiceRecords(stub shim.ChaincodeStubInterface, invoiceRecArr string) pb.Response {

	var contextObj Context
	InitCache(&contextObj)

	var rejecedInvoices []BCIRejectedInvoice
	var invoicesByStatus string
	var invoicesBySupplier map[string]string
	invoicesBySupplier = make(map[string]string)
	var invoicesByBuyer map[string]string
	invoicesByBuyer = make(map[string]string)

	err := json.Unmarshal([]byte(invoiceRecArr), &rejecedInvoices)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input rejected invoice array:", err)
	}

	for _, rejInvoice := range rejecedInvoices {
		rejInvoice_inv := rejInvoice.Invoice
		bciId := rejInvoice_inv.BCIID
		scanID := rejInvoice_inv.DcDocumentData.DcHeader.ScanID
		db.TableStruct{Stub: stub, TableName: util.TAB_INVOICE, PrimaryKeys: []string{scanID, bciId}, Data: string(util.MarshalToBytes(rejInvoice_inv))}.Add()
		UpdateInvoiceStatus(stub, &contextObj, bciId, scanID, INV_STATUS_REJECTED, rejInvoice.ReasonCode, "", "ST9999", EMPTY_ADDITIONAL_INFO)
		db.TableStruct{Stub: stub, TableName: util.TAB_INVOICE_STATUS, PrimaryKeys: []string{scanID, bciId}, Data: string(util.MarshalToBytes(contextObj.StatusHistory))}.Add()
		contextObj.StatusHistory = contextObj.StatusHistory[:0]

		invoiceKey := scanID + "~" + bciId
		// collect invoices by supplier
		if invoicesByStatus != "" {
			invoicesByStatus = invoicesByStatus + "|" + invoiceKey
		} else {
			invoicesByStatus = invoiceKey
		}
		//}

		if rejInvoice_inv.DcDocumentData.DcHeader.VendorID != "" {
			if invoicesBySupplier[rejInvoice_inv.DcDocumentData.DcHeader.VendorID] != "" {
				invoicesBySupplier[rejInvoice_inv.DcDocumentData.DcHeader.VendorID] = invoicesBySupplier[rejInvoice_inv.DcDocumentData.DcHeader.VendorID] + "|" + invoiceKey
			} else {
				invoicesBySupplier[rejInvoice_inv.DcDocumentData.DcHeader.VendorID] = invoiceKey
			}
		}
		// collect invoices by buyer
		var po po.PO
		poRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO, PrimaryKeys: []string{rejInvoice_inv.DcDocumentData.DcHeader.PoNumber}, Data: ""}.Get()
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

	}
	util.UpdateReferenceData(stub, util.TAB_INVOICE_BY_STATUS, []string{INV_STATUS_REJECTED}, invoicesByStatus)

	for supplierId, val := range invoicesBySupplier {
		util.UpdateReferenceData(stub, util.TAB_INVOICE_BY_SUPPLIER, []string{supplierId}, val)
	}
	for buyerId, val := range invoicesByBuyer {
		util.UpdateReferenceData(stub, util.TAB_INVOICE_BY_BUYER, []string{buyerId}, val)
	}

	return shim.Success(nil)
}

func IdentifyAdditionalLineItems(invoice Invoice) Invoice {
	for idx, line := range invoice.DcDocumentData.DcLines {
		if util.StrippedLowerCase(line.PoNumber) == "ac" ||
			util.StrippedLowerCase(line.PoNumber) == "additionalcost" ||
			strings.Contains(pme.Standardize(line.Description), pme.Standardize("handling charges")) ||
			strings.Contains(pme.Standardize(line.Description), pme.Standardize("freight")) ||
			strings.Contains(pme.Standardize(line.Description), pme.Standardize("other")) ||
			strings.Contains(pme.Standardize(line.PoNumber), pme.Standardize("AC")) ||
			strings.Contains(pme.Standardize(line.PoNumber), pme.Standardize("ac")) {
			invoice.DcDocumentData.AdditionalLineItems = append(invoice.DcDocumentData.AdditionalLineItems, line)
			invoice.DcDocumentData.DcLines = append(invoice.DcDocumentData.DcLines[:idx], invoice.DcDocumentData.DcLines[idx+1:]...)
		}
	}
	return invoice
}

func ManageDCLineItem(stub shim.ChaincodeStubInterface, invoice Invoice, isAdd bool) Invoice {
	myLogger.Debugf("Line History Started")
	var line_index string
	if isAdd {
		for idx, line := range invoice.DcDocumentData.DcLines {
			myLogger.Debugf("Line ID>>>>>>>>", line.LineID)
			line_index = invoice.DcDocumentData.DcHeader.InvoiceNumber + "_" + strconv.Itoa(idx)
			invoice.DcDocumentData.DcLines[idx].LineID = line_index
		}
	} else {
		var INV_OLDLINEMAP map[string]DCLine
		INV_OLDLINEMAP = make(map[string]DCLine)
		curr_time := time.Now().Format("02/01/2006 15:04:05 MST")
		oldInvoice, _ := GetInvoice(stub, []string{invoice.DcDocumentData.DcHeader.ScanID, invoice.BCIID})
		for _, inv_line := range oldInvoice.DcDocumentData.DcLines {
			INV_OLDLINEMAP[inv_line.LineID] = inv_line
		}

		for idx, line := range invoice.DcDocumentData.DcLines {
			myLogger.Debugf("Line ID: ", line.LineID)
			var msg string
			var lineId string
			if line.LineID == "" {
				//For New Line Item is added
				line_index = invoice.DcDocumentData.DcHeader.InvoiceNumber + "_" + strconv.Itoa(GetHighestRankCount(invoice.DcDocumentData.DcLines)+1)
				myLogger.Debugf("Line_Index_Id", line_index)
				invoice.DcDocumentData.DcLines[idx].LineID = line_index
				lineId = line_index
				msg = fmt.Sprintf("%s %s %s %v~%v~%v", "User", "Added", "LineItem With", line.Description, line.Quantity, line.UnitPrice)
			} else {
				if value, exist := INV_OLDLINEMAP[line.LineID]; exist {
					//For Line Item is updated
					lineId = line.LineID
					oldDCLine := value
					if oldDCLine.Description != line.Description || oldDCLine.Quantity != line.Quantity || oldDCLine.UnitPrice != line.UnitPrice {
						msg = fmt.Sprintf("%s %s %s %v~%v~%v %s %v~%v~%v", "User", "Updated", "LineItem With", oldDCLine.Description, oldDCLine.Quantity, oldDCLine.UnitPrice, "To", line.Description, line.Quantity, line.UnitPrice)
					} else {
						continue
					}
				}
			}
			InsertInvLineHistoryRecord(stub, lineId, msg, curr_time, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, line)
		}
		//For Line Item Deleted
		for _, dline := range invoice.DcDocumentData.DcLines {
			//if INV_OLDLINEMAP[dline.LineID] != ""{
			if _, exist := INV_OLDLINEMAP[dline.LineID]; exist {
				delete(INV_OLDLINEMAP, dline.LineID)
			}
		}

		for delLineId, delLine := range INV_OLDLINEMAP {
			delMsg := fmt.Sprintf("%s %s %s %v~%v~%v", "User", "Deleted", "LineItem With", delLine.Description, delLine.Quantity, delLine.UnitPrice)
			InsertInvLineHistoryRecord(stub, delLineId, delMsg, curr_time, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, delLine)
		}

	}
	return invoice
}

func InsertInvLineHistoryRecord(stub shim.ChaincodeStubInterface, lineId string, msg string, curr_time string, bciId string, scanId string, line DCLine) {
	var dcLineHistory DCLineHistory
	dcLineHistory.LineHistoryID = util.GetUUID()
	dcLineHistory.LineID = lineId
	dcLineHistory.BciID = bciId
	dcLineHistory.ScanID = scanId
	//dcLineHistory.SetInvoiceNumber(invNo)
	dcLineHistory.Time = util.CreateDateObject(curr_time)
	dcLineHistory.Description = line.Description
	dcLineHistory.Amount = line.Amount
	dcLineHistory.InvoiceLine = line.InvoiceLine
	dcLineHistory.Quantity = line.Quantity
	dcLineHistory.TaxAmount = line.TaxAmount
	dcLineHistory.TaxPercent = line.TaxPercent
	dcLineHistory.TaxCode = line.TaxCode
	dcLineHistory.UnitPrice = line.UnitPrice
	dcLineHistory.PoLine = line.PoLine
	dcLineHistory.Message = msg
	myLogger.Debugf("DCLine Record>>>>>>>>>>>", dcLineHistory)
	db.TableStruct{Stub: stub, TableName: util.TAB_INVOICE_LINE_HISTORY, PrimaryKeys: []string{dcLineHistory.LineHistoryID}, Data: string(util.MarshalToBytes(dcLineHistory))}.Add()
}

func GetHighestRankCount(dcLines []DCLine) int {
	var rankArr []int
	for _, line := range dcLines {
		myLogger.Debugf("Line data", line.LineID)
		if line.LineID != "" {
			splitStr := strings.SplitAfter(line.LineID, "_")
			val, _ := strconv.Atoi(splitStr[1])
			rankArr = append(rankArr, val)
		}
	}
	sort.Ints(rankArr)
	myLogger.Debugf("Total Counting of LineId:", rankArr)
	if len(rankArr) <= 0 {
		return 0
	}
	return rankArr[len(rankArr)-1]
}

func AddInvoiceRecords(stub shim.ChaincodeStubInterface, invoiceRecArr string, isAdd bool) pb.Response {
	var invoices []Invoice
	var invoicesBySupplier map[string]string
	invoicesBySupplier = make(map[string]string)
	var invoicesByBuyer map[string]string
	invoicesByBuyer = make(map[string]string)
	var invoicesByStatus string

	err := json.Unmarshal([]byte(invoiceRecArr), &invoices)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input invoice array:", err)
	}
	var contextObj Context
	InitCache(&contextObj)
	for _, invoice := range invoices {
		// Identify additional lineitems from invoice lines and add them to AdditionalLineItems
		invoice = IdentifyAdditionalLineItems(invoice)
		invoice = ManageDCLineItem(stub, invoice, isAdd)
		contextObj.Invoice = invoice
		AddInvoiceForDuplicateCheck(stub, &contextObj)
		AddInvoice(stub, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, string(util.MarshalToBytes(invoice)), isAdd)
		invoiceKey := invoice.DcDocumentData.DcHeader.ScanID + "~" + invoice.BCIID
		// collect invoices by supplier
		if invoicesByStatus != "" {
			invoicesByStatus = invoicesByStatus + "|" + invoiceKey
		} else {
			invoicesByStatus = invoiceKey
		}
		if invoice.DcDocumentData.DcHeader.VendorID != "" {
			if invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] != "" {
				invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] = invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] + "|" + invoiceKey
			} else {
				invoicesBySupplier[invoice.DcDocumentData.DcHeader.VendorID] = invoiceKey
			}
		}

		// collect invoices by buyer
		var po po.PO
		poRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PO, PrimaryKeys: []string{invoice.DcDocumentData.DcHeader.ErpSystem, invoice.DcDocumentData.DcHeader.PoNumber, invoice.DcDocumentData.DcHeader.Client}, Data: ""}.Get()
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
	}

	for supplierId, val := range invoicesBySupplier {
		util.UpdateReferenceData(stub, util.TAB_INVOICE_BY_SUPPLIER, []string{supplierId}, val)
	}

	for buyerId, val := range invoicesByBuyer {
		util.UpdateReferenceData(stub, util.TAB_INVOICE_BY_BUYER, []string{buyerId}, val)
	}

	if isAdd {
		util.UpdateReferenceData(stub, util.TAB_INVOICE_BY_STATUS, []string{INV_STATUS_INIT}, invoicesByStatus)
	}

	return shim.Success(nil)
}

func GetInvoiceRecord(stub shim.ChaincodeStubInterface, bciId string, ScanId string) pb.Response {
	var invoiceDetails DetailedInvoice
	invoice, _ := GetInvoice(stub, []string{ScanId, bciId})
	invStat, _ := GetInvoiceStatus(stub, []string{ScanId, bciId})
	invoiceDetails.Invoice = invoice
	invoiceDetails.InvoiceStatus = invStat

	return shim.Success(util.MarshalToBytes(invoiceDetails))
}

func AddInvoice(stub shim.ChaincodeStubInterface, bciId string, scanID string, invoiceStr string, isAdd bool) {
	var contextObj Context
	InitCache(&contextObj)
	db.TableStruct{Stub: stub, TableName: util.TAB_INVOICE, PrimaryKeys: []string{scanID, bciId}, Data: invoiceStr}.Add()
	if isAdd {
		contextObj.StatusHistory = contextObj.StatusHistory[:0]
		UpdateInvoiceStatus(stub, &contextObj, bciId, scanID, INV_STATUS_INIT, "", "", "ST0000", EMPTY_ADDITIONAL_INFO)
		db.TableStruct{Stub: stub, TableName: util.TAB_INVOICE_STATUS, PrimaryKeys: []string{scanID, bciId}, Data: string(util.MarshalToBytes(contextObj.StatusHistory))}.Add()
	}
}

func GetInvoicesByVendorID(stub shim.ChaincodeStubInterface, vendorId string) pb.Response {
	invoices, _ := GetInvoicesBy(stub, vendorId, util.TAB_INVOICE_BY_SUPPLIER)
	return shim.Success(util.MarshalToBytes(invoices))
}

func GetInvoicesByBuyerID(stub shim.ChaincodeStubInterface, buyerId string) pb.Response {
	invoices, _ := GetInvoicesBy(stub, buyerId, util.TAB_INVOICE_BY_BUYER)
	return shim.Success(util.MarshalToBytes(invoices))
}

func GetInvoicesBy(stub shim.ChaincodeStubInterface, id string, tableName string) ([]DetailedInvoice, error) {
	record, fetchErr := db.TableStruct{Stub: stub, TableName: tableName, PrimaryKeys: []string{id}, Data: ""}.Get()
	if fetchErr != nil {
		myLogger.Debugf("ERROR fetching invoices By criteria :", fetchErr, record)
	}
	var invoices []DetailedInvoice
	if record != "" {
		invKeys := strings.Split(record, "|")
		for _, invKey := range invKeys {
			invPrimaryKeys := strings.Split(invKey, "~")
			var detailedInv DetailedInvoice
			inv, _ := GetInvoice(stub, []string{invPrimaryKeys[0], invPrimaryKeys[1]})
			invStat, _ := GetInvoiceStatus(stub, []string{invPrimaryKeys[0], invPrimaryKeys[1]})
			detailedInv.Invoice = inv
			detailedInv.InvoiceStatus = invStat
			invoices = append(invoices, detailedInv)
		}
	}
	return invoices, nil
}

func GetInvoice(stub shim.ChaincodeStubInterface, keys []string) (Invoice, string) {
	invoiceRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_INVOICE, PrimaryKeys: keys, Data: ""}.Get()
	var invoice Invoice
	err := json.Unmarshal([]byte(invoiceRecord), &invoice)
	if err != nil {
		myLogger.Debugf("ERROR parsing invoice :", invoiceRecord, err)
		return invoice, "ERROR parsing invoice"
	}
	return invoice, ""
}

func GetAllInvoices(stub shim.ChaincodeStubInterface) pb.Response {
	invoiceRecMap, _ := db.TableStruct{Stub: stub, TableName: util.TAB_INVOICE, PrimaryKeys: []string{}, Data: ""}.GetAll()
	var invoices []Invoice
	for _, invoiceRecord := range invoiceRecMap {
		var invoice Invoice
		err := json.Unmarshal([]byte(invoiceRecord), &invoice)
		if err != nil {
			myLogger.Debugf("ERROR parsing invoice :", invoiceRecord, err)
			return shim.Error("ERROR parsing invoice")
		}
		invoices = append(invoices, invoice)
	}
	return shim.Success(util.MarshalToBytes(invoices))
}

func SetInvoiceStatus(stub shim.ChaincodeStubInterface, contextObjPtr *Context, bciId string, scanID string, status string, reasonCode string, comments string, internalStatus string, info AdditionalInfo) (InvoiceStatus, string) {
	myLogger.Debugf("SetInvoiceStatus============>")
	invStat := UpdateInvoiceStatus(stub, contextObjPtr, bciId, scanID, status, reasonCode, comments, internalStatus, info)
	myLogger.Debugf("Updated status==============>", invStat.Status)
	StoreInvoiceStatusHistory(stub, contextObjPtr, bciId, scanID, status)
	invoice, _ := GetInvoice(stub, []string{scanID, bciId})
	myLogger.Debugf("Status received in this hitttt==============>", status)
	if status == "REJECTED" {
		myLogger.Debugf("Status is Rejected and hence inside==============>", invStat.Status, status)
		db.TableStruct{Stub: stub, TableName: util.TAB_INV_UNIQUE_KEYS, PrimaryKeys: []string{invoice.DcDocumentData.DcHeader.InvoiceNumber, util.GetStringFromFloat(invoice.DcDocumentData.DcHeader.TotalAmount), invoice.DcDocumentData.DcHeader.VendorID}, Data: ""}.Delete()
		myLogger.Debugf("Deleted record from Tab_inv_Unique_keys table=======================")
		myLogger.Debugf("Trigger to event for sending email for REJECTION=====================")
		//responsePayload := TriggerEvent(stub, invoice, "REJECTED")
		responsePayload := "{\"ResponseCode:200\",\"status:" + "success" + "\"}"
		myLogger.Debugf("Event Triggered for sending email-Rejected Invoice=========================", responsePayload)

	}
	if status == "AWAITING BUYER ACTION" {
		myLogger.Debugf("Trigger to event for sending email for buyer action=====================")
		//responsePayload := TriggerEvent(stub, invoice, "AWAITING BUYER ACTION")
		responsePayload := "{\"ResponseCode:200\",\"status:" + "success" + "\"}"
		myLogger.Debugf("Event Triggered for sending email-buyer action=========================", responsePayload)
		//Logic shud be written for inserting into invoice email table.

	}
	return invStat, ""
}

func CalculateTurnAroundDate(invStat InvoiceStatus) InvoiceStatus {
	myLogger.Debugf("Inside Calculate Due Date")
	time := invStat.Time
	var turnAroundDate util.BCDate
	turnAroundDate.SetTime(time.Time().AddDate(0, 0, 2))
	//turnAroundDate := time.Time().AddDate(0,0,2)
	invStat.TurnAroundTime = turnAroundDate
	return invStat
}

func GetTurnAroundDetail(stub shim.ChaincodeStubInterface, userType string) pb.Response {
	var turnAroundDetails TurnAroundDetails
	var zeroday, minusoneday, minustwoday, minusthreeday, oneday, twoday, threeday, threeabove int64
	var invoiceArr []DetailedInvoice
	myLogger.Debugf("Inside Get turn around details=================")
	var ct, turnAroundDate util.BCDate
	ct.SetTime(time.Now().Local())

	myLogger.Debugf("time now==", ct)
	myLogger.Debugf("UserType===============", userType)

	if userType == "buyer" {
		invoiceArr = GetInvoiceByStatus(stub, "AWAITING BUYER ACTION")
		//	invoiceArr, _ = GetInvoicesBy(stub, buyerId, util.TAB_INVOICE_BY_BUYER)
	} else if userType == "ibmap" {
		invoiceArr = GetInvoiceByStatus(stub, "AWAITING IBM AP ACTION")
	}
	for _, invoiceDetail := range invoiceArr {
		invStat := invoiceDetail.InvoiceStatus
		//time := invStat[len(invStat)-1].Time
		//turnAroundDate.SetTime(time.Time().AddDate(0,0,2))
		turnAroundDate = invStat[len(invStat)-1].TurnAroundTime

		myLogger.Debugf("Calculated Turn Around Time===================", turnAroundDate)

		duration := turnAroundDate.Time().Sub(ct.Time())
		diff := int(duration.Hours() / 24)
		myLogger.Debugf("Duration between two days=============", duration)
		myLogger.Debugf("Diff between two days=============", diff)

		switch diff {
		case 0:
			zeroday++
			myLogger.Debugf("case 0")
		case -1:
			minusoneday++
			myLogger.Debugf("case -1")
		case -2:
			minustwoday++
			myLogger.Debugf("case -2")
		case -3:
			minusthreeday++
			myLogger.Debugf("case -3")
		case 1:
			oneday++
			myLogger.Debugf("case 1")
		case 2:
			twoday++
			myLogger.Debugf("case 2")
		case 3:
			threeday++
			myLogger.Debugf("case 3")
		default:
			threeabove++
			myLogger.Debugf("case default")
		}
	}

	turnAroundDetails.Zeroday = zeroday
	turnAroundDetails.MinusOneDays = minusoneday
	turnAroundDetails.MinusTwoDays = minustwoday
	turnAroundDetails.MinusthreeDays = minusthreeday
	turnAroundDetails.OneDay = oneday
	turnAroundDetails.TwoDays = twoday
	turnAroundDetails.ThreeDays = threeday
	turnAroundDetails.ThreeDaysAndAbove = threeabove

	return shim.Success(util.MarshalToBytes(turnAroundDetails))
}

func GetDueDateDetail(stub shim.ChaincodeStubInterface, userType string) pb.Response {
	var dueDateDetails DueDateDetails
	var zeroday, oneday, twotofiveday, pastdue, fiveabove int64
	var invoiceArr []DetailedInvoice
	myLogger.Debugf("Inside Get due date details=================")
	var ct, dueDate util.BCDate
	ct.SetTime(time.Now().Local())

	myLogger.Debugf("time now==", ct)

	if userType == "buyer" {
		invoiceArr = GetInvoiceByStatus(stub, "AWAITING BUYER ACTION")
		//invoiceArr, _ = GetInvoicesBy(stub, buyerId, util.TAB_INVOICE_BY_BUYER)
	} else if userType == "ibmap" {
		invoiceArr = GetInvoiceByStatus(stub, "AWAITING IBM AP ACTION")
	}

	for _, invoiceDetail := range invoiceArr {
		//invStat :=invoiceDetail.InvoiceStatus
		//time := invStat[len(invStat)-1].Time
		//turnAroundDate.SetTime(time.Time().AddDate(0,0,2))
		dueDate = invoiceDetail.Invoice.DcDocumentData.DcHeader.DueDate

		myLogger.Debugf("DueDate in Invoice===================", dueDate)

		duration := dueDate.Time().Sub(ct.Time())
		diff := int(duration.Hours() / 24)
		myLogger.Debugf("Duration between two days=============", duration)
		myLogger.Debugf("Diff between two days=============", diff)

		switch {

		case diff == 0:
			zeroday++
			myLogger.Debugf("case 0")
		case diff == 1:
			oneday++
			myLogger.Debugf("case 1")
		case diff == 2:
			twotofiveday++
			myLogger.Debugf("case 2")
		case diff == 3:
			twotofiveday++
			myLogger.Debugf("case 3")
		case diff == 4:
			twotofiveday++
			myLogger.Debugf("case 4")
		case diff == 5:
			twotofiveday++
			myLogger.Debugf("case 5")
		case diff < 0:
			pastdue++
			myLogger.Debugf("case -1")
		case diff > 5:
			fiveabove++
			myLogger.Debugf("case diff")
		}
	}

	dueDateDetails.DueToday = zeroday
	dueDateDetails.DueInOneDay = oneday
	dueDateDetails.TwoToFive = twotofiveday
	dueDateDetails.MoreThanFive = fiveabove
	dueDateDetails.PastDue = pastdue

	return shim.Success(util.MarshalToBytes(dueDateDetails))
}
func GetInvoicesForIbmap(stub shim.ChaincodeStubInterface) pb.Response {
	var invoiceArr []DetailedInvoice
	var finalInvArr []DetailedInvoice
	statusArr := [9]string{"AWAITING IBM AP ACTION", "AWAITING BUYER ACTION", "PROCESSED", "REJECTED", "AWAITING VMD ACTION", "WAITING FOR GRN",
		"WAITING DB REFRESH FOR PO", "WAITING DB REFRESH FOR VENDOR", "WAITING FOR MANUAL POSTING"}
	for _, status := range statusArr {
		invoiceArr = GetInvoiceByStatus(stub, status)
		for _, invoiceDetail := range invoiceArr {
			finalInvArr = append(finalInvArr, invoiceDetail)
		}
	}
	return shim.Success(util.MarshalToBytes(finalInvArr))
}

func StoreInvoiceStatusHistory(stub shim.ChaincodeStubInterface, contextObjPtr *Context, bciId string, scanID string, status string) {
	myLogger.Debugf("StoreInvoiceStatusHistory============>")
	db.TableStruct{Stub: stub, TableName: util.TAB_INVOICE_STATUS, PrimaryKeys: []string{scanID, bciId}, Data: string(util.MarshalToBytes(contextObjPtr.StatusHistory))}.Add()
	util.UpdateReferenceData(stub, util.TAB_INVOICE_BY_STATUS, []string{status}, scanID+"~"+bciId)
}

func UpdateInvoiceStatus(stub shim.ChaincodeStubInterface, contextObjPtr *Context, bciId string, scanID string, status string, reasonCode string, comments string, internalStatus string, info AdditionalInfo) InvoiceStatus {
	myLogger.Debugf("bci Id ===========", bciId, scanID)
	invStat := CreateInvoiceStatus(stub, contextObjPtr, bciId, scanID, status, reasonCode, comments, internalStatus, info)
	myLogger.Debugf("invStat============", invStat)
	contextObjPtr.StatusHistory = append(contextObjPtr.StatusHistory, invStat)
	myLogger.Debugf("STATUS_HISTORY===========", contextObjPtr.StatusHistory)
	return invStat
}

func CreateInvoiceStatus(stub shim.ChaincodeStubInterface, contextObjPtr *Context, bciId string, scanId string, status string, reasonCode string, comments string, internalStatus string, info AdditionalInfo) InvoiceStatus {
	var invStat InvoiceStatus
	var ct util.BCDate
	ct.SetTime(time.Now().Local())
	myLogger.Debugf("Inside sreatestatus============")
	invStat.BciId = bciId
	//invStat.SetInvoiceNumber(invoiceNumber)
	invStat.ScanID = scanId
	invStat.Status = status
	invStat.ReasonCode = reasonCode
	invStat.Comments = comments
	invStat.InternalStatus = internalStatus
	invStat.InternalDescription = PROCESS_INTERNAL_STEP[invStat.InternalStatus]

	invStat.Time = ct
	invStat.AdditionalInfo = info
	invStat.UserId = contextObjPtr.userIDGlobal
	invStat.ProcessMapStep = PROCESS_MAP_STEP[invStat.InternalStatus]

	if status == "AWAITING BUYER ACTION" {
		invStat = CalculateTurnAroundDate(invStat)
	}
	if status == "AWAITING IBM AP ACTION" {
		myLogger.Debugf("Status is awaiting IBMAP action and hence inside==============>", invStat.Status, status)
		invStat = CalculateTurnAroundDate(invStat)
		myLogger.Debugf("Setting the turn around date =======================", invStat)
	}

	return invStat
}
func GetInvoiceByStatus(stub shim.ChaincodeStubInterface, invoiceStatus string) []DetailedInvoice {
	var invoicesArr []DetailedInvoice
	invoiceKeys := FetchInvoiceByStatus(stub, invoiceStatus)
	for _, invoiceKey := range invoiceKeys {
		primaryKeys := strings.Split(invoiceKey, "~")
		if len(primaryKeys) == 2 {
			var invoiceDetails DetailedInvoice
			invoice, _ := GetInvoice(stub, []string{primaryKeys[0], primaryKeys[1]})
			invStat, _ := GetInvoiceStatus(stub, []string{primaryKeys[0], primaryKeys[1]})
			invoiceDetails.Invoice = invoice
			invoiceDetails.InvoiceStatus = invStat
			invoicesArr = append(invoicesArr, invoiceDetails)
		}
	}
	return invoicesArr
}

func FetchInvoiceByStatus(stub shim.ChaincodeStubInterface, invoiceStatus string) []string {
	var invoiceKeys []string
	record, err := db.TableStruct{Stub: stub, TableName: util.TAB_INVOICE_BY_STATUS, PrimaryKeys: []string{invoiceStatus}, Data: ""}.Get()
	if err != nil {
		myLogger.Debugf("Error : Getting the invoice by status")
	} else {
		invoiceKeys = strings.Split(record, "|")
	}
	return invoiceKeys
}

func GetInvoiceStatus(stub shim.ChaincodeStubInterface, keys []string) ([]InvoiceStatus, string) {
	invStatRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_INVOICE_STATUS, PrimaryKeys: keys, Data: ""}.Get()

	var invStat []InvoiceStatus
	err := json.Unmarshal([]byte(invStatRec), &invStat)
	if err != nil {
		myLogger.Debugf("ERROR parsing invoice status:", err)
		return invStat, "ERROR parsing invoice status"
	}
	return invStat, ""
}

func GetSupplierEmail(stub shim.ChaincodeStubInterface, vendorId string) string {
	// Problem needs to be fixed
	vendor, fetchErr := vmd.GetVendor(stub, "", vendorId, "")
	if fetchErr != "" {
		myLogger.Debugf("Error in fetching vendor record for vendor ID : ", vendorId)
		return ""
	}
	return vendor.VendorEmail
}

func CreateAdditionalInfo(infoType string, infoValue string) AdditionalInfo {
	var additionalInfo AdditionalInfo
	additionalInfo = *GetAdditionalInfo(infoType, infoValue)
	return additionalInfo
}

func GetPosByIBMAP(stub shim.ChaincodeStubInterface) []po.PO {
	var uniquePONumbers []string
	var pos []po.PO
	invoiceKeys := FetchInvoiceByStatus(stub, INV_STATUS_PENDING_AP)
	for _, invoiceKey := range invoiceKeys {
		primaryKeys := strings.Split(invoiceKey, "~")
		if len(primaryKeys) == 2 {
			invoice, _ := GetInvoice(stub, []string{primaryKeys[0], primaryKeys[1]})
			if !(util.StringArrayContains(uniquePONumbers, invoice.DcDocumentData.DcHeader.PoNumber)) {
				uniquePONumbers = append(uniquePONumbers, invoice.DcDocumentData.DcHeader.PoNumber)
			}
		}
	}

	for _, poNumber := range uniquePONumbers {
		po, _ := po.GetPO(stub, []string{poNumber})
		pos = append(pos, po)
	}
	return pos
}

func GetInvoicesByBuyerIDFilter(stub shim.ChaincodeStubInterface, buyerId string, invoiceKey string) {
	invoices, _ := GetInvoicesByFilter(stub, buyerId, util.TAB_INVOICE_BY_BUYER, invoiceKey)
	myLogger.Debugf("values-===========", invoices)

}

func GetInvoicesByFilter(stub shim.ChaincodeStubInterface, buyerId string, tableName string, invoiceKey string) ([]DetailedInvoice, error) {
	record, fetchErr := db.TableStruct{Stub: stub, TableName: tableName, PrimaryKeys: []string{buyerId}, Data: ""}.Get()
	util.RemoveRecord(stub, util.TAB_INVOICE_BY_BUYER, []string{buyerId}, invoiceKey)
	if fetchErr != nil {
		myLogger.Debugf("ERROR fetching invoices By criteria :", fetchErr, record)
	}
	var invoices []DetailedInvoice
	var invoicesByBuyer map[string]string
	invoicesByBuyer = make(map[string]string)
	if record != "" {
		invKeys := strings.Split(record, "|")
		for _, invKey := range invKeys {
			invPrimaryKeys := strings.Split(invKey, "~")
			invoicePrimaryKeys := strings.Split(invoiceKey, "~")
			myLogger.Debugf("invPrimaryKeys=======", invPrimaryKeys)
			myLogger.Debugf("invoicePrimaryKeys=======", invoicePrimaryKeys)
			var detailedInv DetailedInvoice
			if invoicePrimaryKeys[0] != invPrimaryKeys[0] && invoicePrimaryKeys[1] != invPrimaryKeys[1] {
				myLogger.Debugf("Entered if condition")
				inv, _ := GetInvoice(stub, []string{invPrimaryKeys[0], invPrimaryKeys[1]})
				invStat, _ := GetInvoiceStatus(stub, []string{invPrimaryKeys[0], invPrimaryKeys[1]})
				detailedInv.Invoice = inv
				detailedInv.InvoiceStatus = invStat
				invoices = append(invoices, detailedInv)
				if buyerId != "" {
					if invoicesByBuyer[buyerId] != "" {
						invoicesByBuyer[buyerId] = invoicesByBuyer[buyerId] + "|" + invKey
					} else {
						invoicesByBuyer[buyerId] = invKey
					}
				}
			}
		}

		myLogger.Debugf("Invoices to be loaded again===================", invoicesByBuyer)

		for buyerId, val := range invoicesByBuyer {

			recs3 := db.TableStruct{Stub: stub, TableName: tableName, PrimaryKeys: []string{buyerId}, Data: val}.Add()
			myLogger.Debugf("recs3===========", recs3)
		}

	}
	return invoices, nil
}

func GetInvoiceLineHistory(stub shim.ChaincodeStubInterface, bciId string, invoiceNo string) pb.Response {
	var lineHistory []DCLineHistory
	var line DCLineHistory
	records, _ := db.TableStruct{Stub: stub, TableName: util.TAB_INVOICE_LINE_HISTORY, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, lineRec := range records {
		err := json.Unmarshal([]byte(lineRec), &line)
		if err != nil {
			myLogger.Debugf("ERROR parsing lineHistory :", records, err)
			return shim.Error("ERROR parsing LineHistory")
		}
		if line.BciID == bciId && line.InvoiceNumber == invoiceNo {
			lineHistory = append(lineHistory, line)
		}
	}

	return shim.Success(util.MarshalToBytes(lineHistory))
}

func GetPOByBuyerVendorName(stub shim.ChaincodeStubInterface, buyerID string, companyCode string) ([]po.PO, error) {

	myLogger.Debugf("input for filters================", buyerID, companyCode)
	record, fetchErr := db.TableStruct{Stub: stub, TableName: util.TAB_PO_BY_BUYER, PrimaryKeys: []string{buyerID}, Data: ""}.Get()
	if fetchErr != nil {
		myLogger.Debugf("ERROR in parsing po :", fetchErr)
	}
	var pos []po.PO
	if record != "" {
		poNumbers := strings.Split(record, "|")
		for _, poNumber := range poNumbers {
			po, _ := po.GetPO(stub, []string{poNumber})
			myLogger.Debugf("currentPo==============", po)
			InvVendorDetails, err := vmd.GetVendor(stub, po.ERPSystem, po.VendorId, po.Client)
			myLogger.Debugf("InvVendorDetails============", InvVendorDetails, err)
			if err == "" {
				if po.CompanyCode == companyCode {
					//if util.ProbableMatch(po.CompanyCode(), companyCode) {
					myLogger.Debugf("Entered if condition")
					vendor_Name := InvVendorDetails.VendorName
					if strings.Contains(vendor_Name, "IBM") || strings.Contains(vendor_Name, "International Business Machines") {
						pos = append(pos, po)
					}

				}
				//}
			}

		}
	}
	return pos, nil
}

func GetPosByIBMAPVendorName(stub shim.ChaincodeStubInterface, companyCode string) []po.PO {
	var uniquePONumbers []string
	var pos []po.PO
	invoiceKeys := FetchInvoiceByStatus(stub, INV_STATUS_PENDING_AP)
	for _, invoiceKey := range invoiceKeys {
		primaryKeys := strings.Split(invoiceKey, "~")
		if len(primaryKeys) == 2 {
			invoice, _ := GetInvoice(stub, []string{primaryKeys[0], primaryKeys[1]})
			if !(util.StringArrayContains(uniquePONumbers, invoice.DcDocumentData.DcHeader.PoNumber)) {
				uniquePONumbers = append(uniquePONumbers, invoice.DcDocumentData.DcHeader.PoNumber)
			}
		}
	}

	for _, poNumber := range uniquePONumbers {
		po, _ := po.GetPO(stub, []string{poNumber})
		fmt.Printf("currentPo==============", po)
		InvVendorDetails, err := vmd.GetVendor(stub, po.ERPSystem, po.VendorId, po.Client)
		myLogger.Debugf("InvVendorDetails============", InvVendorDetails, err)
		if err == "" {
			if po.CompanyCode == companyCode {
				fmt.Printf("Entered if condition")
				vendor_Name := InvVendorDetails.VendorName
				if strings.Contains(vendor_Name, "IBM") || strings.Contains(vendor_Name, "International Business Machines") {
					pos = append(pos, po)
				}

			}

		}
	}
	return pos
}

func AddSAPProcessedInvoiceRecords(stub shim.ChaincodeStubInterface, invoiceRecArr string) pb.Response {
	var invoices []Invoice

	err := json.Unmarshal([]byte(invoiceRecArr), &invoices)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input invoice array:", err)
	}
	var contextObj Context
	InitCache(&contextObj)
	for _, invoice := range invoices {
		contextObj.Invoice = invoice
		AddInvoiceForDuplicateCheck(stub, &contextObj)
		AddSAPProcessedInvoice(stub, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, string(util.MarshalToBytes(invoice)))
	}
	return shim.Success(nil)
}

func AddSAPProcessedInvoice(stub shim.ChaincodeStubInterface, bciId string, scanID string, invoiceStr string) {
	db.TableStruct{Stub: stub, TableName: util.TAB_SAP_PROCESSED_INVOICE, PrimaryKeys: []string{scanID, bciId}, Data: invoiceStr}.Add()
}

func GetSAPProcessedInvoiceRecord(stub shim.ChaincodeStubInterface, bciId string, ScanId string) pb.Response {
	var invoiceDetails DetailedInvoice
	invoice, _ := GetSAPProcessedInvoice(stub, []string{ScanId, bciId})
	invoiceDetails.Invoice = invoice
	return shim.Success(util.MarshalToBytes(invoiceDetails))
}

func GetSAPProcessedInvoice(stub shim.ChaincodeStubInterface, keys []string) (Invoice, string) {
	invoiceRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_SAP_PROCESSED_INVOICE, PrimaryKeys: keys, Data: ""}.Get()
	var invoice Invoice
	err := json.Unmarshal([]byte(invoiceRecord), &invoice)
	if err != nil {
		myLogger.Debugf("ERROR parsing invoice :", invoiceRecord, err)
		return invoice, "ERROR parsing invoice"
	}
	return invoice, ""
}

func GetAllSAPProcessedInvoices(stub shim.ChaincodeStubInterface) pb.Response {
	invoiceRecMap, _ := db.TableStruct{Stub: stub, TableName: util.TAB_SAP_PROCESSED_INVOICE, PrimaryKeys: []string{}, Data: ""}.GetAll()
	var invoices []Invoice
	for _, invoiceRecord := range invoiceRecMap {
		var invoice Invoice
		err := json.Unmarshal([]byte(invoiceRecord), &invoice)
		if err != nil {
			myLogger.Debugf("ERROR parsing invoice :", invoiceRecord, err)
			return shim.Error("ERROR parsing invoice")
		}
		invoices = append(invoices, invoice)
	}
	return shim.Success(util.MarshalToBytes(invoices))
}
