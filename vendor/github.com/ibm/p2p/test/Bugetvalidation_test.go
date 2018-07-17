package invoice

import (
	"bytes"
	"encoding/json"
	"html/template"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type MockChainCode struct {
}

func (sc *MockChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Success(nil)
}

//Invoke is the entry point for any transaction
func (sc *MockChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	var inputInvoice Invoice
	_, args := stub.GetFunctionAndParameters()
	err := json.Unmarshal([]byte(args[0]), &inputInvoice)
	if err != nil {
		shim.Error("Invalid invoice json")
	}
	trxnContext := Context{}
	trxnContext.Invoice = inputInvoice
	statCode, errMsg, invStatus := Budgetvalidation(stub, &trxnContext)
	returnMessage := make(map[string]interface{})
	returnMessage["statCode"] = statCode
	returnMessage["errMsg"] = errMsg
	returnMessage["invStatus"] = invStatus
	returnMessage["modifiedInvoice"] = trxnContext.Invoice
	returnPayload, _ := json.MarshalIndent(returnMessage, "", " ")
	return shim.Success(returnPayload)
}

//Scenario for budgetvalidation
func Test_Budgetvalidation_sc1(t *testing.T) {

	scc := new(MockChainCode)
	stub := shim.NewMockStub("bv1", scc)
	initRes := stub.MockInit("920202-342222-112134", [][]byte{[]byte("init")})
	if initRes.Status != shim.OK {
		t.Logf("Initalization failed")
		t.FailNow()
	}
	invoiceDetails := getInvoiceJSON(_lc_INVOICE_LIST_UNIQUEONE)
	inputInvoice, _ := buildInvoiceInput(_lc_INVOICE_LIST_UNIQUEONE)
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("la"), invoiceDetails})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {
		//t.Logf("%s", resp.Payload)
		modifiedInvoice := extractInvoice(resp.Payload)
		t.Logf("Output invoice line count %d", len(modifiedInvoice.DcDocumentData.DcLines))
			if invStatus.ReasonCode == "ST0000" || invStatus.ReasonCode == "st-budgetValidation-2" {
						t.FailNow()
			}
	}

}
 

//Build an invoice
func buildInvoiceInput(templateName string) (Invoice, error) {
	var invoice Invoice
	data := getInvoiceJSON(templateName)
	err := json.Unmarshal(data, &invoice)
	return invoice, err
}

//Extract an invoice from the returned result
func extractInvoice(payload []byte) Invoice {
	dataMap := make(map[string]interface{})
	json.Unmarshal(payload, &dataMap)
	invoiceJSON, _ := json.Marshal(dataMap["modifiedInvoice"])

	var invoice Invoice
	json.Unmarshal(invoiceJSON, &invoice)

	return invoice
}

/*func Test_Input(t *testing.T) {
	invoice, err := buildInvoiceInput(_lc_INVOICE_LIST_UNIQUEONE)
	if err != nil {
		t.Logf("%v", err)
		t.FailNow()
	} else {
		t.Logf("%v", invoice)
	}
}*/

func getInvoiceJSON(templateName string) []byte {
	dummyMap := make(map[string]string)
	tmpl, _ := template.New("invoice").Parse(templateName)
	var invoiceBytes bytes.Buffer
	tmpl.Execute(&invoiceBytes, dummyMap)
	return invoiceBytes.Bytes()
}

const _lc_INVOICE_LIST_UNIQUEONE = `
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
		  "PO_Line_Number": 10,
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
  }`
  
const _lc_INVOICE_LIST_MULTI = `
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
		},
		{
			"Scan_ID": "20180414_OB01997482",
			"Invoice_Line": 30,
			"Amount": 18,
			"Tax_Percent": 0,
			"Tax_Amount": 0,
			"Tax_Code": "I0",
			"PO_Number": "DO12366735",
			"PO_Line_Number": null,
			"PO_Release": 0,
			"Delivery_Note": "",
			"Unit_Price": 0.9,
			"Quantity": 20,
			"MAT_NUMBER": "95664",
			"CostCenter": "",
			"InternalOrder": "",
			"GLAccount": "",
			"Description": "95664"
		  }
	  ],
	  "xmlns": "http://tempuri.org/DCDocumentData.xsd"
	},
	"Status": "Success"
  }`