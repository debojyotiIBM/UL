package invoice

import (
	"bytes"
	"encoding/json"
	"html/template"
	"testing"

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
	statCode, errMsg, invStatus := LineAggregration(stub, &trxnContext)
	returnMessage := make(map[string]interface{})
	returnMessage["statCode"] = statCode
	returnMessage["errMsg"] = errMsg
	returnMessage["invStatus"] = invStatus
	returnMessage["modifiedInvoice"] = trxnContext.Invoice
	returnPayload, _ := json.MarshalIndent(returnMessage, "", " ")
	return shim.Success(returnPayload)
}

//Scenario for no aggregration
func Test_LineAggregation_WithMaterialNumber(t *testing.T) {

	scc := new(MockChainCode)
	stub := shim.NewMockStub("la1", scc)
	initRes := stub.MockInit("920202-342222-112134", [][]byte{[]byte("init")})
	if initRes.Status != shim.OK {
		t.Logf("Initalization failed")
		t.FailNow()
	}

	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	invoice.DcDocumentData.DcHeader.TaxAmount = 0.0
	dcLines := make([]DCLine, 3)
	//============================= Single PO ==========================
	dcLines[0] = buildDCLineItems(1, "AAAA-00001", 10, "XXX-1", "", 50, 2)
	dcLines[1] = buildDCLineItems(2, "AAAA-00001", 20, "XXX-1", "", 50, 2)
	dcLines[2] = buildDCLineItems(3, "AAAA-00001", 80, "XXX-1", "", 50, 2)
	invoice.DcDocumentData.DcLines = dcLines
	invoiceDetails, _ := json.Marshal(invoice)
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("la"), invoiceDetails})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {
		//t.Logf("%s", resp.Payload)
		modifiedInvoice := extractInvoice(resp.Payload)
		t.Logf("Output invoice line count %d", len(modifiedInvoice.DcDocumentData.DcLines))
		if len(modifiedInvoice.DcDocumentData.DcLines) != 1 {
			t.Fail()
		}
		displayDCLiles(modifiedInvoice.DcDocumentData.DcLines, t)
	}

	//============================= Single PO With Additional Items ==========================
	dcLines = make([]DCLine, 4)
	dcLines[0] = buildDCLineItems(1, "AAAA-00001", 10, "XXX-1", "", 50, 2)
	dcLines[1] = buildDCLineItems(2, "AAAA-00001", 20, "XXX-1", "", 50, 2)
	dcLines[2] = buildDCLineItems(3, "AAAA-00001", 80, "XXX-1", "", 50, 2)
	dcLines[3] = buildDCLineItems(4, "AC", 0, "", "", 50, 0)

	invoice.DcDocumentData.DcLines = dcLines
	invoiceDetails, _ = json.Marshal(invoice)
	resp = stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("la"), invoiceDetails})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {
		//t.Logf("%s", resp.Payload)
		modifiedInvoice := extractInvoice(resp.Payload)
		t.Logf("Output invoice line count %d", len(modifiedInvoice.DcDocumentData.DcLines))
		if len(modifiedInvoice.DcDocumentData.DcLines) != 2 {
			t.Fail()
		}
		displayDCLiles(modifiedInvoice.DcDocumentData.DcLines, t)

	}
	//============================= Multiple PO With Additional Items ==========================
	dcLines = make([]DCLine, 5)
	dcLines[0] = buildDCLineItems(1, "AAAA-00002", 10, "XXX-1", "", 50, 2)
	dcLines[1] = buildDCLineItems(2, "AAAA-00001", 20, "XXX-1", "", 50, 2)
	dcLines[2] = buildDCLineItems(3, "AAAA-00002", 80, "XXX-1", "", 50, 2)
	dcLines[3] = buildDCLineItems(4, "AAAA-00001", 80, "XXX-1", "", 50, 2)
	dcLines[4] = buildDCLineItems(5, "AC", 0, "", "", 50, 0)

	invoice.DcDocumentData.DcLines = dcLines
	invoiceDetails, _ = json.Marshal(invoice)
	resp = stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("la"), invoiceDetails})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {
		//t.Logf("%s", resp.Payload)
		modifiedInvoice := extractInvoice(resp.Payload)
		t.Logf("Output invoice line count %d", len(modifiedInvoice.DcDocumentData.DcLines))
		if len(modifiedInvoice.DcDocumentData.DcLines) != 3 {
			t.Fail()
		}
		displayDCLiles(modifiedInvoice.DcDocumentData.DcLines, t)

	}
	//============================= Multiple PO With Additional Items ==========================
	dcLines = make([]DCLine, 5)
	dcLines[0] = buildDCLineItems(1, "AAAA-00002", 10, "XXX-1", "", 50, 2)
	dcLines[1] = buildDCLineItems(2, "AAAA-00001", 20, "XXX-1", "", 50, 2)
	dcLines[2] = buildDCLineItems(3, "AAAA-00002", 80, "XXX-1", "", 50, 2)
	dcLines[3] = buildDCLineItems(4, "AAAA-00001", 80, "XXX-1", "", 50, 2)
	dcLines[4] = buildDCLineItems(5, "AC", 0, "", "", 50, 0)

	invoice.DcDocumentData.DcLines = dcLines
	invoiceDetails, _ = json.Marshal(invoice)
	resp = stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("la"), invoiceDetails})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {
		//t.Logf("%s", resp.Payload)
		modifiedInvoice := extractInvoice(resp.Payload)
		t.Logf("Output invoice line count %d", len(modifiedInvoice.DcDocumentData.DcLines))
		if len(modifiedInvoice.DcDocumentData.DcLines) != 3 {
			t.Fail()
		}
		displayDCLiles(modifiedInvoice.DcDocumentData.DcLines, t)

	}
	//============================= Multiple PO With Additional Items ==========================
	dcLines = make([]DCLine, 6)
	dcLines[0] = buildDCLineItems(1, "AAAA-00002", 10, "XXX-1", "", 50, 2)
	dcLines[1] = buildDCLineItems(2, "AAAA-00001", 20, "XXX-1", "", 50, 2)
	dcLines[2] = buildDCLineItems(3, "AAAA-00002", 80, "XXX-1", "", 50, 2)
	dcLines[3] = buildDCLineItems(4, "AAAA-00001", 80, "XXX-1", "", 50, 3)
	dcLines[4] = buildDCLineItems(5, "AAAA-00001", 80, "XXX-1", "", 50, 2)
	dcLines[5] = buildDCLineItems(6, "AC", 0, "", "", 50, 0)

	invoice.DcDocumentData.DcLines = dcLines
	invoiceDetails, _ = json.Marshal(invoice)
	resp = stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("la"), invoiceDetails})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {
		//t.Logf("%s", resp.Payload)
		modifiedInvoice := extractInvoice(resp.Payload)
		t.Logf("Output invoice line count %d", len(modifiedInvoice.DcDocumentData.DcLines))
		if len(modifiedInvoice.DcDocumentData.DcLines) != 4 {
			t.Fail()
		}

		displayDCLiles(modifiedInvoice.DcDocumentData.DcLines, t)

	}
}
func Test_LineAggregation_WithNoMaterialNumber(t *testing.T) {

	scc := new(MockChainCode)
	stub := shim.NewMockStub("la1", scc)
	initRes := stub.MockInit("920202-342222-112134", [][]byte{[]byte("init")})
	if initRes.Status != shim.OK {
		t.Logf("Initalization failed")
		t.FailNow()
	}

	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	invoice.DcDocumentData.DcHeader.TaxAmount = 0.0
	dcLines := make([]DCLine, 3)
	//============================= Single PO ==========================
	dcLines[0] = buildDCLineItems(1, "AAAA-00001", 10, "", "Pen", 50, 2)
	dcLines[1] = buildDCLineItems(2, "AAAA-00001", 20, "", "Pen", 50, 2)
	dcLines[2] = buildDCLineItems(3, "AAAA-00001", 80, "", "Pen", 50, 2)
	invoice.DcDocumentData.DcLines = dcLines
	invoiceDetails, _ := json.Marshal(invoice)
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("la"), invoiceDetails})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {
		//t.Logf("%s", resp.Payload)
		modifiedInvoice := extractInvoice(resp.Payload)
		t.Logf("Output invoice line count %d", len(modifiedInvoice.DcDocumentData.DcLines))
		if len(modifiedInvoice.DcDocumentData.DcLines) != 1 {
			t.Fail()
		}
		displayDCLiles(modifiedInvoice.DcDocumentData.DcLines, t)
	}

	//============================= Single PO With Additional Items ==========================
	dcLines = make([]DCLine, 4)
	dcLines[0] = buildDCLineItems(1, "AAAA-00001", 10, "", "Pen", 50, 2)
	dcLines[1] = buildDCLineItems(2, "AAAA-00001", 20, "", "Pen", 50, 2)
	dcLines[2] = buildDCLineItems(3, "AAAA-00001", 80, "", "Pen", 50, 2)
	dcLines[3] = buildDCLineItems(4, "AC", 0, "", "", 50, 0)

	invoice.DcDocumentData.DcLines = dcLines
	invoiceDetails, _ = json.Marshal(invoice)
	resp = stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("la"), invoiceDetails})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {
		//t.Logf("%s", resp.Payload)
		modifiedInvoice := extractInvoice(resp.Payload)
		t.Logf("Output invoice line count %d", len(modifiedInvoice.DcDocumentData.DcLines))
		if len(modifiedInvoice.DcDocumentData.DcLines) != 2 {
			t.Fail()
		}
		displayDCLiles(modifiedInvoice.DcDocumentData.DcLines, t)

	}
	//============================= Multiple PO With Additional Items ==========================
	dcLines = make([]DCLine, 5)
	dcLines[0] = buildDCLineItems(1, "AAAA-00002", 10, "", "Pen", 50, 2)
	dcLines[1] = buildDCLineItems(2, "AAAA-00001", 20, "", "Pen", 50, 2)
	dcLines[2] = buildDCLineItems(3, "AAAA-00002", 80, "", "Pen", 50, 2)
	dcLines[3] = buildDCLineItems(4, "AAAA-00001", 80, "", "Pen", 50, 2)
	dcLines[4] = buildDCLineItems(5, "AC", 0, "", "", 50, 0)

	invoice.DcDocumentData.DcLines = dcLines
	invoiceDetails, _ = json.Marshal(invoice)
	resp = stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("la"), invoiceDetails})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {
		//t.Logf("%s", resp.Payload)
		modifiedInvoice := extractInvoice(resp.Payload)
		t.Logf("Output invoice line count %d", len(modifiedInvoice.DcDocumentData.DcLines))
		if len(modifiedInvoice.DcDocumentData.DcLines) != 3 {
			t.Fail()
		}
		displayDCLiles(modifiedInvoice.DcDocumentData.DcLines, t)

	}
	//============================= Multiple PO With Additional Items ==========================
	dcLines = make([]DCLine, 5)
	dcLines[0] = buildDCLineItems(1, "AAAA-00002", 10, "", "Pen", 50, 2)
	dcLines[1] = buildDCLineItems(2, "AAAA-00001", 20, "", "Pen", 50, 2)
	dcLines[2] = buildDCLineItems(3, "AAAA-00002", 80, "", "Pen", 50, 2)
	dcLines[3] = buildDCLineItems(4, "AAAA-00001", 80, "", "Pen", 50, 2)
	dcLines[4] = buildDCLineItems(5, "AC", 0, "", "", 50, 0)

	invoice.DcDocumentData.DcLines = dcLines
	invoiceDetails, _ = json.Marshal(invoice)
	resp = stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("la"), invoiceDetails})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {
		//t.Logf("%s", resp.Payload)
		modifiedInvoice := extractInvoice(resp.Payload)
		t.Logf("Output invoice line count %d", len(modifiedInvoice.DcDocumentData.DcLines))
		if len(modifiedInvoice.DcDocumentData.DcLines) != 3 {
			t.Fail()
		}
		displayDCLiles(modifiedInvoice.DcDocumentData.DcLines, t)

	}
	//============================= Multiple PO With Additional Items ==========================
	dcLines = make([]DCLine, 6)
	dcLines[0] = buildDCLineItems(1, "AAAA-00002", 10, "", "Pen", 50, 2)
	dcLines[1] = buildDCLineItems(2, "AAAA-00001", 20, "", "Pen", 50, 2)
	dcLines[2] = buildDCLineItems(3, "AAAA-00002", 80, "", "Pen", 50, 2)
	dcLines[3] = buildDCLineItems(4, "AAAA-00001", 80, "", "Pen", 50, 3)
	dcLines[4] = buildDCLineItems(5, "AAAA-00001", 80, "", "Pen", 50, 2)
	dcLines[5] = buildDCLineItems(6, "AC", 0, "", "", 50, 0)

	invoice.DcDocumentData.DcLines = dcLines
	invoiceDetails, _ = json.Marshal(invoice)
	resp = stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("la"), invoiceDetails})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {
		//t.Logf("%s", resp.Payload)
		modifiedInvoice := extractInvoice(resp.Payload)
		t.Logf("Output invoice line count %d", len(modifiedInvoice.DcDocumentData.DcLines))
		if len(modifiedInvoice.DcDocumentData.DcLines) != 4 {
			t.Fail()
		}

		displayDCLiles(modifiedInvoice.DcDocumentData.DcLines, t)

	}
}

func Test_LineAggregation_HeaderTax(t *testing.T) {

	scc := new(MockChainCode)
	stub := shim.NewMockStub("la1", scc)
	initRes := stub.MockInit("920202-342222-112134", [][]byte{[]byte("init")})
	if initRes.Status != shim.OK {
		t.Logf("Initalization failed")
		t.FailNow()
	}

	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	invoice.DcDocumentData.DcHeader.TaxAmount = 200.00
	dcLines := make([]DCLine, 3)

	//============================= Multiple PO With Additional Items ==========================
	dcLines = make([]DCLine, 6)
	dcLines[0] = buildDCLineItems(1, "PO-1", 10, "", "Pen", 50, 2)
	dcLines[1] = buildDCLineItems(2, "PO-1", 20, "", "Pen", 50, 2)
	dcLines[2] = buildDCLineItems(3, "PO-1", 80, "", "Pen", 50, 3)
	dcLines[3] = buildDCLineItems(4, "P0-2", 80, "", "Pen", 50, 3)
	dcLines[4] = buildDCLineItems(5, "P0-2", 80, "", "Pen", 50, 2)
	dcLines[5] = buildDCLineItems(6, "AC", 0, "", "", 50, 0)

	invoice.DcDocumentData.DcLines = dcLines
	invoiceDetails, _ := json.Marshal(invoice)
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("la"), invoiceDetails})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	} else {
		//t.Logf("%s", resp.Payload)
		modifiedInvoice := extractInvoice(resp.Payload)
		t.Logf("Output invoice line count %d", len(modifiedInvoice.DcDocumentData.DcLines))
		if len(modifiedInvoice.DcDocumentData.DcLines) != 3 {
			t.Fail()
		}

		displayDCLiles(modifiedInvoice.DcDocumentData.DcLines, t)

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

func Test_Input(t *testing.T) {
	t.SkipNow()
	invoice, err := buildInvoiceInput("_no_mat_desc_up_match")

	if err != nil {
		t.Logf("%v", err)
		t.FailNow()
	} else {
		t.Logf("%v", invoice)
	}
}

func getInvoiceJSON(templateName string) []byte {
	dummyMap := make(map[string]string)
	tmpl, _ := template.New("invoice").Parse(templateName)
	var invoiceBytes bytes.Buffer
	tmpl.Execute(&invoiceBytes, dummyMap)
	return invoiceBytes.Bytes()
}
func buildDCLineItems(lineNumber int, poNumber string, qty int, matnumber, description string, unitPrice, taxPercent float64) DCLine {
	dcLineItemMap := make(map[string]interface{})
	dcLineItemMap["Scan_ID"] = "234987234732_423462364"
	dcLineItemMap["Invoice_Line"] = lineNumber
	dcLineItemMap["Amount"] = float64(qty) * unitPrice
	dcLineItemMap["Tax_Percent"] = taxPercent / float64(100)
	dcLineItemMap["Tax_Amount"] = float64(qty) * unitPrice * taxPercent / float64(100)
	dcLineItemMap["Tax_Code"] = "I0"
	dcLineItemMap["PO_Number"] = poNumber
	dcLineItemMap["PO_Release"] = 0
	dcLineItemMap["Delivery_Note"] = ""
	dcLineItemMap["Unit_Price"] = unitPrice
	dcLineItemMap["Quantity"] = qty
	dcLineItemMap["MAT_NUMBER"] = matnumber
	dcLineItemMap["CostCenter"] = ""
	dcLineItemMap["Description"] = description
	dcLineItemMap["GLAccount"] = ""
	dcLineItemMap["InternalOrder"] = ""
	var dcLine DCLine
	jsonByes, _ := json.Marshal(dcLineItemMap)
	json.Unmarshal(jsonByes, &dcLine)
	return dcLine
}
func displayDCLiles(lines []DCLine, t *testing.T) {
	for index, line := range lines {
		t.Logf("Line item %d PO %s Mat %s DESC %s UP %.2f TAXP %.2f TAX %.2f", (index + 1), line.PoNumber, line.MatNumber, line.Description, line.UnitPrice, line.TaxPercent, line.TaxAmount)
	}
}
