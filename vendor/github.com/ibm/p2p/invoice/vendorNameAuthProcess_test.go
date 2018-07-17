package invoice

import (
	"encoding/json"
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

//Test case to test vendor does not existing in VMD
func Test_VNAPVendorIDNotInVMD(t *testing.T) {

	stub := initVNASChaincode(t)
	invoiceJSON := getInvoiceJSON(_testdata_VALID_INVOICE)

	//invoiceJSON, _ := json.Marshal(invoice)

	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("vendNameAuth"), invoiceJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.Logf("%+v", resp)
		t.FailNow()
	}
	invStatus := extractInvoiceStatus(resp.Payload)
	if invStatus.Status != CONTINUE || invStatus.InternalStatus != _lc_INTERNAL_STATUS_MOVE_TO_REMITTOID {
		t.FailNow()
	}
}
func Test_VNAPSinglePOVendorIDMatches(t *testing.T) {

	//Prepare the reference data

	vmdData := GenerateTestDataInMap([]byte(buildSpec(vmdDataSpecMin)))
	if vmdData == nil {
		t.FailNow()
		t.Logf("\n Could not generate VMD Data")
	}
	vendorID, _ := vmdData["vendorid"].(string)
	poData := GenerateTestDataInMap([]byte(buildSpec(poDataSpecMin)))
	if poData == nil {
		t.FailNow()
		t.Logf("\n Could not generate PO data")
	}
	poData["vendorid"] = vendorID
	poNumber, _ := poData["ponumber"].(string)
	clientID, _ := poData["client"].(string)
	erpSystem, _ := vmdData["erpsystem"].(string)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	//Set the vendor id
	invoice.DcDocumentData.DcHeader.VendorID = vendorID
	invoice.DcDocumentData.DcHeader.Client = clientID
	invoice.DcDocumentData.DcHeader.ErpSystem = erpSystem
	//Set the same PO number
	for index, dcLine := range invoice.DcDocumentData.DcLines {
		dcLine.PoNumber = poNumber
		invoice.DcDocumentData.DcLines[index] = dcLine
	}
	stub := initVNASChaincode(t)
	saveVMD(stub, vmdData, t)
	savePO(stub, poData, t)
	prettyJSON, _ := json.MarshalIndent(vmdData, " ", " ")
	t.Logf("Vendor injected \n %s", string(prettyJSON))
	prettyJSON, _ = json.MarshalIndent(poData, " ", " ")
	t.Logf("PO injected \n %s", string(prettyJSON))

	invoiceJSON, _ := json.Marshal(invoice)
	t.Logf("Invoice client %s vendorId %s ", invoice.DcDocumentData.DcHeader.Client, invoice.DcDocumentData.DcHeader.VendorID)
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("vendNameAuth"), invoiceJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.Logf("%+v", resp)
		t.FailNow()
	}
	invStatus := extractInvoiceStatus(resp.Payload)
	if invStatus.Status != CONTINUE || invStatus.InternalStatus != _lc_INTERNAL_STATUS_MOVE_TO_REMITTOID {
		t.FailNow()
	}
	prettyJSON, _ = json.MarshalIndent(invStatus, " ", "  ")
	t.Logf("Final Inventory Status %s", prettyJSON)
}
func Test_VNAPSinglePOVendIDDoNotMatchWithVMD(t *testing.T) {

	//Prepare the reference data

	vmdData := GenerateTestDataInMap([]byte(buildSpec(vmdDataSpecMin)))
	if vmdData == nil {
		t.FailNow()
		t.Logf("\n Could not generate VMD Data")
	}
	vendorID, _ := vmdData["vendorid"].(string)
	poData := GenerateTestDataInMap([]byte(buildSpec(poDataSpecMin)))
	if poData == nil {
		t.FailNow()
		t.Logf("\n Could not generate PO data")
	}

	poNumber, _ := poData["ponumber"].(string)
	clientID, _ := poData["client"].(string)
	erpSystem, _ := vmdData["erpsystem"].(string)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	//Set the vendor id
	invoice.DcDocumentData.DcHeader.ErpSystem = erpSystem
	invoice.DcDocumentData.DcHeader.VendorID = vendorID
	invoice.DcDocumentData.DcHeader.Client = clientID
	//Set the same PO number
	for index, dcLine := range invoice.DcDocumentData.DcLines {
		dcLine.PoNumber = poNumber
		invoice.DcDocumentData.DcLines[index] = dcLine
	}
	stub := initVNASChaincode(t)
	saveVMD(stub, vmdData, t)
	savePO(stub, poData, t)
	prettyJSON, _ := json.MarshalIndent(vmdData, " ", " ")
	t.Logf("Vendor injected \n %s", string(prettyJSON))

	prettyJSON, _ = json.MarshalIndent(poData, " ", " ")
	t.Logf("PO injected \n %s", string(prettyJSON))
	t.Logf("Invoice  client %s ", invoice.DcDocumentData.DcHeader.Client)
	t.Logf("Invoice  vendor id %s ", invoice.DcDocumentData.DcHeader.VendorID)

	invoiceJSON, _ := json.Marshal(invoice)
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("vendNameAuth"), invoiceJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.Logf("%+v", resp)
		t.FailNow()
	}
	invStatus := extractInvoiceStatus(resp.Payload)
	prettyJSON, _ = json.MarshalIndent(invStatus, " ", "  ")
	t.Logf("Final Inventory Status %s", prettyJSON)
	if invStatus.Status != REJECT || invStatus.InternalStatus != _lc_INTERNAL_STATUS_MOVE_TO_VENDNAMEAUTH {
		t.FailNow()
	}

}
func Test_VNAPSinglePONameMatchWithVMD(t *testing.T) {

	//Prepare the reference data

	vmdData := GenerateTestDataInMap([]byte(buildSpec(vmdDataSpecMin)))
	if vmdData == nil {
		t.FailNow()
		t.Logf("\n Could not generate VMD Data")
	}
	vendorID, _ := vmdData["vendorid"].(string)
	vmdDataForPO := GenerateTestDataInMap([]byte(buildSpec(vmdDataSpecMin)))
	if vmdDataForPO == nil {
		t.FailNow()
		t.Logf("\n Could not generate VMD Data")
	}

	poData := GenerateTestDataInMap([]byte(buildSpec(poDataSpecMin)))
	if poData == nil {
		t.FailNow()
		t.Logf("\n Could not generate PO data")
	}

	poNumber, _ := poData["ponumber"].(string)
	poData["vendorid"] = vmdDataForPO["vendorid"]
	erpSystem, _ := vmdData["erpsystem"].(string)
	clientID, _ := poData["client"].(string)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	//Set the vendor id
	invoice.DcDocumentData.DcHeader.VendorID = vendorID
	invoice.DcDocumentData.DcHeader.Client = clientID
	invoice.DcDocumentData.DcHeader.ErpSystem = erpSystem
	vmdData["vendorname"] = vmdDataForPO["vendorname"]
	//Set the same PO number
	for index, dcLine := range invoice.DcDocumentData.DcLines {
		dcLine.PoNumber = poNumber
		invoice.DcDocumentData.DcLines[index] = dcLine
	}
	stub := initVNASChaincode(t)
	saveVMD(stub, vmdData, t)
	saveVMD(stub, vmdDataForPO, t)
	savePO(stub, poData, t)
	prettyJSON, _ := json.MarshalIndent(vmdData, " ", " ")
	t.Logf("Vendor injected \n %s", string(prettyJSON))
	prettyJSON, _ = json.MarshalIndent(vmdDataForPO, " ", " ")
	t.Logf("Vendor injected for PO \n %s", string(prettyJSON))
	prettyJSON, _ = json.MarshalIndent(poData, " ", " ")
	t.Logf("PO injected \n %s", string(prettyJSON))

	invoiceJSON, _ := json.Marshal(invoice)
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("vendNameAuth"), invoiceJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.Logf("%+v", resp)
		t.FailNow()
	}
	invStatus := extractInvoiceStatus(resp.Payload)
	prettyJSON, _ = json.MarshalIndent(invStatus, " ", "  ")
	t.Logf("Final Inventory Status %s", prettyJSON)
	if invStatus.Status != CONTINUE || invStatus.InternalStatus != _lc_INTERNAL_STATUS_MOVE_TO_REMITTOID {
		t.FailNow()
	}

}
func Test_VNAPSinglePONameMatchWithSupplerName(t *testing.T) {

	//Prepare the reference data
	vmdData := GenerateTestDataInMap([]byte(buildSpec(vmdDataSpecMin)))
	if vmdData == nil {
		t.FailNow()
		t.Logf("\n Could not generate VMD Data")
	}
	vendorID, _ := vmdData["vendorid"].(string)
	erpSystem, _ := vmdData["erpsystem"].(string)
	vmdDataForPO := GenerateTestDataInMap([]byte(buildSpec(vmdDataSpecMin)))
	if vmdDataForPO == nil {
		t.FailNow()
		t.Logf("\n Could not generate VMD Data")
	}

	poData := GenerateTestDataInMap([]byte(buildSpec(poDataSpecMin)))
	if poData == nil {
		t.FailNow()
		t.Logf("\n Could not generate PO data")
	}

	poNumber, _ := poData["ponumber"].(string)
	poData["vendorid"] = vmdDataForPO["vendorid"]
	clientID, _ := poData["client"].(string)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	//Set the vendor id
	invoice.DcDocumentData.DcHeader.VendorID = vendorID
	invoice.DcDocumentData.DcHeader.Client = clientID
	invoice.DcDocumentData.DcHeader.ErpSystem = erpSystem
	vendName, _ := vmdDataForPO["vendorname"].(string)
	invoice.DcDocumentData.DcSwissHeader.SupplierName = vendName
	//Making sure that vmd Master has a diffrent name
	vmdData["vendorname"] = "XXX_YYYY_XXXXX"
	//Set the same PO number
	for index, dcLine := range invoice.DcDocumentData.DcLines {
		dcLine.PoNumber = poNumber
		invoice.DcDocumentData.DcLines[index] = dcLine
	}
	stub := initVNASChaincode(t)
	saveVMD(stub, vmdData, t)
	saveVMD(stub, vmdDataForPO, t)
	savePO(stub, poData, t)
	prettyJSON, _ := json.MarshalIndent(vmdData, " ", " ")
	t.Logf("Vendor injected \n %s", string(prettyJSON))
	prettyJSON, _ = json.MarshalIndent(vmdDataForPO, " ", " ")
	t.Logf("Vendor injected for PO \n %s", string(prettyJSON))
	prettyJSON, _ = json.MarshalIndent(poData, " ", " ")
	t.Logf("PO injected \n %s", string(prettyJSON))

	invoiceJSON, _ := json.Marshal(invoice)
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("vendNameAuth"), invoiceJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.Logf("%+v", resp)
		t.FailNow()
	}
	invStatus := extractInvoiceStatus(resp.Payload)
	prettyJSON, _ = json.MarshalIndent(invStatus, " ", "  ")
	t.Logf("Final Inventory Status %s", prettyJSON)
	if invStatus.Status != CONTINUE || invStatus.InternalStatus != _lc_INTERNAL_STATUS_MOVE_TO_REMITTOID {
		t.FailNow()
	}

}
func Test_VNAPSingleInvSuppierNameBusAsList(t *testing.T) {

	//Prepare the reference data

	vmdData := GenerateTestDataInMap([]byte(buildSpec(vmdDataSpecMin)))
	if vmdData == nil {
		t.FailNow()
		t.Logf("\n Could not generate VMD Data")
	}
	vendorID, _ := vmdData["vendorid"].(string)
	vmdDataForPO := GenerateTestDataInMap([]byte(buildSpec(vmdDataSpecMin)))
	if vmdDataForPO == nil {
		t.FailNow()
		t.Logf("\n Could not generate VMD Data")
	}

	poData := GenerateTestDataInMap([]byte(buildSpec(poDataSpecMin)))
	if poData == nil {
		t.FailNow()
		t.Logf("\n Could not generate PO data")
	}

	poNumber, _ := poData["ponumber"].(string)
	poData["vendorid"] = vmdDataForPO["vendorid"]
	clientID, _ := poData["client"].(string)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	//Set the vendor id
	invoice.DcDocumentData.DcHeader.VendorID = vendorID
	invoice.DcDocumentData.DcHeader.Client = clientID
	supplierName := "Renolds India"
	invoice.DcDocumentData.DcSwissHeader.SupplierName = supplierName
	//Making sure that vmd Master has a diffrent name
	vmdData["vendorname"] = "XXX_YYYY_XXXXX"
	//Set the same PO number
	for index, dcLine := range invoice.DcDocumentData.DcLines {
		dcLine.PoNumber = poNumber
		invoice.DcDocumentData.DcLines[index] = dcLine
	}
	stub := initVNASChaincode(t)
	saveVMD(stub, vmdData, t)
	saveVMD(stub, vmdDataForPO, t)
	savePO(stub, poData, t)
	saveBusAsList(stub, t, invoice.DcDocumentData.DcHeader.CompanyCode, invoice.DcDocumentData.DcHeader.InvType, invoice.DcDocumentData.DcHeader.DocSource, supplierName)
	prettyJSON, _ := json.MarshalIndent(vmdData, " ", " ")
	t.Logf("Vendor injected \n %s", string(prettyJSON))
	prettyJSON, _ = json.MarshalIndent(vmdDataForPO, " ", " ")
	t.Logf("Vendor injected for PO \n %s", string(prettyJSON))
	prettyJSON, _ = json.MarshalIndent(poData, " ", " ")
	t.Logf("PO injected \n %s", string(prettyJSON))

	invoiceJSON, _ := json.Marshal(invoice)
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("vendNameAuth"), invoiceJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.Logf("%+v", resp)
		t.FailNow()
	}
	invStatus := extractInvoiceStatus(resp.Payload)
	prettyJSON, _ = json.MarshalIndent(invStatus, " ", "  ")
	t.Logf("Final Inventory Status %s", prettyJSON)
	if invStatus.Status != CONTINUE || invStatus.InternalStatus != _lc_INTERNAL_STATUS_MOVE_TO_REMITTOID {
		t.FailNow()
	}

}
func Test_VNAPSingleInvSuppierNameAutoRejOnMismatch(t *testing.T) {

	//Prepare the reference data

	vmdData := GenerateTestDataInMap([]byte(buildSpec(vmdDataSpecMin)))
	if vmdData == nil {
		t.FailNow()
		t.Logf("\n Could not generate VMD Data")
	}
	vendorID, _ := vmdData["vendorid"].(string)
	erpsystem, _ := vmdData["erpsystem"].(string)
	vmdDataForPO := GenerateTestDataInMap([]byte(buildSpec(vmdDataSpecMin)))
	if vmdDataForPO == nil {
		t.FailNow()
		t.Logf("\n Could not generate VMD Data")
	}

	poData := GenerateTestDataInMap([]byte(buildSpec(poDataSpecMin)))
	if poData == nil {
		t.FailNow()
		t.Logf("\n Could not generate PO data")
	}

	poNumber, _ := poData["ponumber"].(string)
	poData["vendorid"] = vmdDataForPO["vendorid"]
	clientID, _ := poData["client"].(string)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	//Set the vendor id
	invoice.DcDocumentData.DcHeader.VendorID = vendorID
	invoice.DcDocumentData.DcHeader.Client = clientID
	invoice.DcDocumentData.DcHeader.ErpSystem = erpsystem
	supplierName := "Renolds India"
	invoice.DcDocumentData.DcSwissHeader.SupplierName = supplierName
	//Making sure that vmd Master has a diffrent name
	vmdData["vendorname"] = "XXX_YYYY_XXXXX"
	//Set the same PO number
	for index, dcLine := range invoice.DcDocumentData.DcLines {
		dcLine.PoNumber = poNumber
		invoice.DcDocumentData.DcLines[index] = dcLine
	}
	stub := initVNASChaincode(t)
	saveVMD(stub, vmdData, t)
	saveVMD(stub, vmdDataForPO, t)
	savePO(stub, poData, t)
	saveAutoRejct(stub, t, invoice.DcDocumentData.DcHeader.CompanyCode, invoice.DcDocumentData.DcHeader.InvType, invoice.DcDocumentData.DcHeader.DocSource)
	prettyJSON, _ := json.MarshalIndent(vmdData, " ", " ")
	t.Logf("Vendor injected \n %s", string(prettyJSON))
	prettyJSON, _ = json.MarshalIndent(vmdDataForPO, " ", " ")
	t.Logf("Vendor injected for PO \n %s", string(prettyJSON))
	prettyJSON, _ = json.MarshalIndent(poData, " ", " ")
	t.Logf("PO injected \n %s", string(prettyJSON))

	invoiceJSON, _ := json.Marshal(invoice)
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("vendNameAuth"), invoiceJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.Logf("%+v", resp)
		t.FailNow()
	}
	invStatus := extractInvoiceStatus(resp.Payload)
	prettyJSON, _ = json.MarshalIndent(invStatus, " ", "  ")
	t.Logf("Final Inventory Status %s", prettyJSON)
	if invStatus.Status != REJECT || invStatus.InternalStatus != _lc_INTERNAL_STATUS_MOVE_TO_VENDNAMEAUTH {
		t.FailNow()
	}

}
func Test_VNAPSingleInvIBMAPAction(t *testing.T) {

	//Prepare the reference data

	vmdData := GenerateTestDataInMap([]byte(buildSpec(vmdDataSpecMin)))
	if vmdData == nil {
		t.FailNow()
		t.Logf("\n Could not generate VMD Data")
	}
	vendorID, _ := vmdData["vendorid"].(string)
	erpsystem, _ := vmdData["erpsystem"].(string)

	vmdDataForPO := GenerateTestDataInMap([]byte(buildSpec(vmdDataSpecMin)))
	if vmdDataForPO == nil {
		t.FailNow()
		t.Logf("\n Could not generate VMD Data")
	}

	poData := GenerateTestDataInMap([]byte(buildSpec(poDataSpecMin)))
	if poData == nil {
		t.FailNow()
		t.Logf("\n Could not generate PO data")
	}

	poNumber, _ := poData["ponumber"].(string)
	poData["vendorid"] = vmdDataForPO["vendorid"]
	clientID, _ := poData["client"].(string)
	invoice, _ := buildInvoiceInput(_testdata_VALID_INVOICE)
	//Set the vendor id
	invoice.DcDocumentData.DcHeader.VendorID = vendorID
	invoice.DcDocumentData.DcHeader.Client = clientID
	invoice.DcDocumentData.DcHeader.ErpSystem = erpsystem
	supplierName := "Renolds India"
	invoice.DcDocumentData.DcSwissHeader.SupplierName = supplierName
	//Making sure that vmd Master has a diffrent name
	vmdData["vendorname"] = "XXX_YYYY_XXXXX"
	//Set the same PO number
	for index, dcLine := range invoice.DcDocumentData.DcLines {
		dcLine.PoNumber = poNumber
		invoice.DcDocumentData.DcLines[index] = dcLine
	}
	stub := initVNASChaincode(t)
	saveVMD(stub, vmdData, t)
	saveVMD(stub, vmdDataForPO, t)
	savePO(stub, poData, t)
	saveRoutesOnFailure(stub, t, invoice.DcDocumentData.DcHeader.CompanyCode, invoice.DcDocumentData.DcHeader.InvType, invoice.DcDocumentData.DcHeader.DocSource)
	prettyJSON, _ := json.MarshalIndent(vmdData, " ", " ")
	t.Logf("Vendor injected \n %s", string(prettyJSON))
	prettyJSON, _ = json.MarshalIndent(vmdDataForPO, " ", " ")
	t.Logf("Vendor injected for PO \n %s", string(prettyJSON))
	prettyJSON, _ = json.MarshalIndent(poData, " ", " ")
	t.Logf("PO injected \n %s", string(prettyJSON))

	invoiceJSON, _ := json.Marshal(invoice)
	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("vendNameAuth"), invoiceJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.Logf("%+v", resp)
		t.FailNow()
	}
	invStatus := extractInvoiceStatus(resp.Payload)
	prettyJSON, _ = json.MarshalIndent(invStatus, " ", "  ")
	t.Logf("Final Inventory Status %s", prettyJSON)
	if invStatus.Status != "X_ACTION" || invStatus.InternalStatus != "st-remitToId-1" {
		t.FailNow()
	}

}

func saveVMD(stub *shim.MockStub, vendorData interface{}, t *testing.T) {
	vendorArray := make([]interface{}, 0)
	vendorArray = append(vendorArray, vendorData)
	inputBytes, _ := json.Marshal(vendorArray)
	resp := stub.MockInvoke("378218-9381273-99900", [][]byte{[]byte("saveVMD"), inputBytes})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.Logf("%+v", resp)
		t.FailNow()
	}
}
func savePO(stub *shim.MockStub, poData interface{}, t *testing.T) {
	poArray := make([]interface{}, 0)
	poArray = append(poArray, poData)
	inputBytes, _ := json.Marshal(poArray)
	resp := stub.MockInvoke("378218-9381273-99901", [][]byte{[]byte("savePO"), inputBytes})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.Logf("%+v", resp)
		t.FailNow()
	}
}
func saveBusAsList(stub *shim.MockStub, t *testing.T, compCode, invType, inputSrc, supplerName string) {

	dynaConfig := make(map[string]interface{})
	dynaConfig["compCode"] = compCode
	dynaConfig["invoiceType"] = invType
	dynaConfig["inputSource"] = inputSrc
	dynaConfig[supplerName] = []string{"WIPRO", "SOFTLAYER"}
	dynaConfig["WIPRO"] = []string{"CATS Ltd", "ZYA Inc", supplerName}

	configJSON, _ := json.Marshal(dynaConfig)

	resp := stub.MockInvoke("378218-9381273-93883", [][]byte{[]byte("saveBASL"), configJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	}

}
func saveAutoRejct(stub *shim.MockStub, t *testing.T, compCode, invType, inputSrc string) {

	dynaConfig := make(map[string]interface{})
	dynaConfig["compCode"] = compCode
	dynaConfig["invoiceType"] = invType
	dynaConfig["inputSource"] = inputSrc
	dynaConfig["autoRejection"] = true
	configJSON, _ := json.Marshal(dynaConfig)

	resp := stub.MockInvoke("378218-9381273-93783", [][]byte{[]byte("saveAutoRej"), configJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	}

}
func saveRoutesOnFailure(stub *shim.MockStub, t *testing.T, compCode, invType, inputSrc string) {

	dynaConfig := make(map[string]interface{})
	dynaConfig["compCode"] = compCode
	dynaConfig["invoiceType"] = invType
	dynaConfig["inputSource"] = inputSrc
	dynaConfig["uiRoute"] = "X_ACTION"
	dynaConfig["internalStatus"] = "st-remitToId-1"
	configJSON, _ := json.Marshal(dynaConfig)

	resp := stub.MockInvoke("378218-9381273-93783", [][]byte{[]byte("saveRoute"), configJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	}

}

const poDataSpecMin = `
[{
	"fname":"erpsystem",
	"type":"pickList",
	
	"pickFromList":["P1P"]
},
{
	"fname":"client",
	"type":"pickList",
	"pickFromList":["010"]
	
},
{
	"fname":"vendorid",
	"type":"an-string",
	"size":8,
	"prefix":"VNA-"
},
{
	"fname":"ponumber",
	"type":"an-string",
	"size":10,
	"prefix":"VNA-"
}
]
`
const vmdDataSpecMin = `
[{
	"fname":"erpsystem",
	"type":"pickList",
	
	"pickFromList":["P1P"]
},
{
	"fname":"client",
	"type":"pickList",
	"pickFromList":["010"]
	
},
{
	"fname":"vendorid",
	"type":"an-string",
	"size":8,
	"prefix":"VNA-"
},
{
	"fname":"vendorname",
	"type":"pickList",
	"pickFromList":["WIPRO","IBM","SANJOY_EXPOTERS"]
	
}
]
`
