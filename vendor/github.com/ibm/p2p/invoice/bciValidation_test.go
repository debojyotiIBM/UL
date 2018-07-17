package invoice

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	util "github.com/ibm/p2p"
)

type BCITestChainCode struct {
}

func (sc *BCITestChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Success(nil)
}

//Invoke is the entry point for any transaction
func (sc *BCITestChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	var inputInvoice Invoice
	_, args := stub.GetFunctionAndParameters()
	err := json.Unmarshal([]byte(args[0]), &inputInvoice)
	if err != nil {
		shim.Error("Invalid invoice json")
	}
	trxnContext := Context{}
	trxnContext.Invoice = inputInvoice
	statCode, errMsg, invStatus := BCIValidation(stub, &trxnContext)
	returnMessage := make(map[string]interface{})
	returnMessage["statCode"] = statCode
	returnMessage["errMsg"] = errMsg
	returnMessage["invStatus"] = invStatus
	returnMessage["modifiedInvoice"] = trxnContext.Invoice
	returnPayload, _ := json.MarshalIndent(returnMessage, "", " ")
	return shim.Success(returnPayload)
}
func initChaincode(t *testing.T) *shim.MockStub {
	scc := new(BCITestChainCode)
	stub := shim.NewMockStub("xa2", scc)
	initRes := stub.MockInit("920202-342222-112134", [][]byte{[]byte("init")})
	if initRes.Status != shim.OK {
		t.Logf("Initalization failed")
		t.FailNow()
	}
	return stub
}

//Scenario for all ok
func Test_BCI_TestChainCode_CheckMandatoryFields(t *testing.T) {
	stub := initChaincode(t)
	invoiceDetails := getInvoiceJSON(_testdata_VALID_INVOICE)
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("bciValidaton"), invoiceDetails})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {

		invStatus := extractInvoiceStatus(resp.Payload)
		if invStatus.Status != CONTINUE {
			t.FailNow()
		}

	}

}
func Test_BCI_BlankVendorId(t *testing.T) {
	stub := initChaincode(t)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	invoice.DcDocumentData.DcHeader.VendorID = ""
	inputJSON, _ := json.MarshalIndent(invoice, "", "")
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("bciValidaton"), inputJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {

		invStatus := extractInvoiceStatus(resp.Payload)
		if invStatus.ReasonCode != "REMIT TO VENDOR ID NOT AVAILABLE" {
			t.FailNow()
		}
	}
}

func Test_BCI_BlankVendorName(t *testing.T) {
	stub := initChaincode(t)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	invoice.DcDocumentData.DcSwissHeader.SupplierName = ""
	inputJSON, _ := json.MarshalIndent(invoice, "", "")
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("bciValidaton"), inputJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {

		invStatus := extractInvoiceStatus(resp.Payload)
		if invStatus.ReasonCode != "REMIT TO VENDOR NAME NOT AVAILABLE" {
			t.FailNow()
		}
	}
}

func Test_BCI_BlankInvoiceNumber(t *testing.T) {
	stub := initChaincode(t)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	invoice.DcDocumentData.DcHeader.InvoiceNumber = ""
	inputJSON, _ := json.MarshalIndent(invoice, "", "")
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("bciValidaton"), inputJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {

		invStatus := extractInvoiceStatus(resp.Payload)
		if invStatus.ReasonCode != "INVOICE NUMBER" {
			t.FailNow()
		}
	}
}
func Test_BCI_BlankCurrency(t *testing.T) {
	stub := initChaincode(t)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	invoice.DcDocumentData.DcHeader.CurrencyCode = ""
	inputJSON, _ := json.MarshalIndent(invoice, "", "")
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("bciValidaton"), inputJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {

		invStatus := extractInvoiceStatus(resp.Payload)
		if invStatus.ReasonCode != "INVOICE CURRENCY  NOT AVAILABLE" {
			t.FailNow()
		}
	}
}
func Test_BCI_BlankCompanyCodeName(t *testing.T) {
	stub := initChaincode(t)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	invoice.DcDocumentData.DcSwissHeader.BuyerName = ""
	inputJSON, _ := json.MarshalIndent(invoice, "", "")
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("bciValidaton"), inputJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {

		invStatus := extractInvoiceStatus(resp.Payload)
		if invStatus.ReasonCode != "BILL TO ADDRESS / COMPANY CODE NAME NOT AVAILABLE" {
			t.FailNow()
		}
	}
}
func Test_BCI_BlankScanId(t *testing.T) {
	stub := initChaincode(t)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	invoice.DcDocumentData.DcHeader.ScanID = ""
	inputJSON, _ := json.MarshalIndent(invoice, "", "")
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("bciValidaton"), inputJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {

		invStatus := extractInvoiceStatus(resp.Payload)
		if invStatus.ReasonCode != "REPLICATE THE E-MAIL TEMPLATE CURRENTLY FOLLOWED IN DCIW" {
			t.FailNow()
		}
	} //BCI_009
}
func Test_BCI_InvalidLineItem(t *testing.T) {
	stub := initChaincode(t)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	invoice.DcDocumentData.DcLines[0].Description = ""
	inputJSON, _ := json.MarshalIndent(invoice, "", "")
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("bciValidaton"), inputJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {

		invStatus := extractInvoiceStatus(resp.Payload)
		t.Logf("%+v", invStatus)
		if !strings.Contains(invStatus.Status, REJECT) {
			t.FailNow()
		}
	}
}
func Test_BCI_InvalidInvoiceDate(t *testing.T) {
	stub := initChaincode(t)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	invoiceDate := util.BCDate{}
	oldDate, _ := time.Parse("2006-01-02", "2016-01-01")
	invoiceDate.SetTime(oldDate)
	invoice.DcDocumentData.DcHeader.DocDate = invoiceDate
	inputJSON, _ := json.MarshalIndent(invoice, "", "")
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("bciValidaton"), inputJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {

		invStatus := extractInvoiceStatus(resp.Payload)
		if invStatus.ReasonCode != "INVOICE DATE OR DELIVERY DATE IS OLDER THAN 1 YEAR" {
			t.FailNow()
		}
	}
}

//Extract an invoice from the returned result
func extractInvoiceStatus(payload []byte) InvoiceStatus {
	dataMap := make(map[string]interface{})
	json.Unmarshal(payload, &dataMap)
	invoiceJSON, _ := json.Marshal(dataMap["invStatus"])

	var invoiceStat InvoiceStatus
	json.Unmarshal(invoiceJSON, &invoiceStat)

	return invoiceStat
}

const _testdata_VALID_INVOICE = `
{
	"BciId": "[B@52ef6577230770560",
	"Version": "0",
	"DCDocumentData": {
	  "DC_Header": {
		"Inv_Type": "INVPO",
		"Doc_Date": "20180412",
		"Scan_Date": "20180414",
		"Company_Code": "2687",
		"Plant_Code": "0000",
		"Invoice_Num": "MO36968248",
		"Total_Amount": 2646.24,
		"Tax_Amount": 0,
		"Currency_Code": "USD",
		"Due_Date": "19000101",
		"Vendor_ID": "0050005687",
		"Vendor_VAT_Code": "N/A",
		"Buyer_VAT_Code": "13-2915928",
		"Scan_ID": "20180414_OB01997482",
		"Image_File": "AAA000150126568.tif",
		"Discount_Date": "19000101",
		"Discount_Percent": 0,
		"Invoice_Exception": "0",
		"Import_Date": "20180414",
		"Import_Status": "READY TO IMPORT",
		"Bank_Account": "1000113402498",
		"Bank_Code": "061000104",
		"Bank_OGM": "",
		"Bank_IBAN": "",
		"Bank_SWIFT": "SNTRUS3A",
		"Payment_Term": "",
		"Description": "",
		"IRPF_Value": 0,
		"Other_Information": "",
		"Nr_Of_Images": 1,
		"Exch_Rate": 0,
		"Indexer_Id": "OB10",
		"ERP_Ref_Number": "AAA000150126568",
		"VAT_Local_Amount": 0,
		"Discount_Amount": 0,
		"POR_Reference": "",
		"TaxReportingCountry": "US",
		"VATRegion": "",
		"AttentionOf": " ",
		"OB10_Link": "https://wwy.ob10.com/Archive?rn=WtmrvHa4bvWij5ROkwXH1mQ79FuPWBbKzeLhYS8J&ob10=AAA408568889&trxno=AAA000150126568",
		"Payment_Reference": "",
		"ShipFrom": null,
		"ShipTo": null,
		"BillFrom": null,
		"BillTo": null,
		"PO_Number": "DO12366735",
		"ErpSystem": "NA",
		"DocSource": "OB10",
		"DocType": "INVPO",
		"Client_ID":"NA"
	  },
	  "DC_Swiss_Header": {
		"Date_Of_Supply": "20180412",
		"Supplier_Name": "MOTION INDUSTRIES INC",
		"Supplier_Address1": "PO BOX 1477",
		"Supplier_Address2": "Birmingham",
		"Supplier_Address3": "AL",
		"Supplier_Address4": "35201",
		"Supplier_Address5": "United States",
		"Supplier_Address6": "",
		"Supplier_Address7": "",
		"Supplier_Address8": "",
		"Supplier_Address9": "",
		"Supplier_Address10": "",
		"Supplier_Fiscal_Name": "",
		"Supplier_Fiscal_Address1": "",
		"Supplier_Fiscal_Address2": "",
		"Supplier_Fiscal_Address3": "",
		"Supplier_Fiscal_Address4": "",
		"Supplier_Fiscal_Address5": "",
		"Supplier_Fiscal_Address6": "",
		"Supplier_Fiscal_Address7": "",
		"Supplier_Fiscal_Address8": "",
		"Supplier_Fiscal_Address9": "",
		"Supplier_Fiscal_Address10": "",
		"Buyer_Name": "Unilever United States, Inc.",
		"Buyer_Address1": "700 Sylvan Avenue",
		"Buyer_Address2": "Englewood Cliffs,",
		"Buyer_Address3": "NJ",
		"Buyer_Address4": "07632",
		"Buyer_Address5": "United States",
		"Buyer_Address6": "",
		"Buyer_Address7": "",
		"Buyer_Address8": "",
		"Buyer_Address9": "",
		"Buyer_Address10": "",
		"Total_Net": 2646.24
	  },
	  "DC_Swiss_Line": {
		"Description_Of_Supply": "AAA000041830"
	  },
	  "DC_Lines": [
		{
		  "Scan_ID": "20180414_OB01997482",
		  "Invoice_Line": 10,
		  "Amount": 9,
		  "Tax_Percent": 0,
		  "Tax_Amount": 0,
		  "Tax_Code": "I0",
		  "PO_Number": "DO12366735",
		  "PO_Line_Number": null,
		  "PO_Release": 0,
		  "Delivery_Note": "",
		  "Unit_Price": 0.9,
		  "Quantity": 10,
		  "MAT_NUMBER": "95664",
		  "CostCenter": "",
		  "InternalOrder": "",
		  "GLAccount": "",
		  "Description": "95664"
		},
		{
		  "Scan_ID": "20180414_OB01997482",
		  "Invoice_Line": 20,
		  "Amount": 10.2,
		  "Tax_Percent": 0,
		  "Tax_Amount": 0,
		  "Tax_Code": "I0",
		  "PO_Number": "DO12366735",
		  "PO_Line_Number": null,
		  "PO_Release": 0,
		  "Delivery_Note": "",
		  "Unit_Price": 1.02,
		  "Quantity": 10,
		  "MAT_NUMBER": "95667",
		  "CostCenter": "",
		  "InternalOrder": "",
		  "GLAccount": "",
		  "Description": "95667"
		}
	  ],
	  "xmlns": "http://tempuri.org/DCDocumentData.xsd"
	},
	"Status": "Success"
  }

`
