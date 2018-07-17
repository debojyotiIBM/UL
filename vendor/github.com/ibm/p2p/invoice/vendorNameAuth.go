package invoice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	po "github.com/ibm/p2p/po"
	vmd "github.com/ibm/p2p/vmd"
	logging "github.com/op/go-logging"
)

const _lc_INTERNAL_STATUS_MOVE_TO_REMITTOID = "st-remitToIid-1"
const _lc_INTERNAL_STATUS_MOVE_TO_VENDNAMEAUTH = "st-vendNameAuth-1"
const _lc_DYNA_CONFIG_VNA_AUTO_REJECT = "dyna_config_vna_auto_rej"
const _lc_DYNA_CONFIG_VNA_UI_ROUTE = "dyna_config_vna_ui_route"
const _lc_DYNA_CONFIG_VNA_BUS_AS_LIST = "dyna_config_vna_bus_as_list"

var _lc_vna_logger = logging.MustGetLogger("Vendor-Name-Authentication-Log")

//VendorNameAuthentication perform business process 3.2 as per the specification
func VendorNameAuthentication(stub shim.ChaincodeStubInterface, trxnContext *Context) (int, string, InvoiceStatus) {
	internalStatus := ""
	errorMessage := ""
	status := ""

	//returnStatusCd := 1
	var invStat InvoiceStatus
	invoice := trxnContext.Invoice
	//3.2.2
	vendorID := strings.TrimSpace(invoice.DcDocumentData.DcHeader.VendorID)
	//I am not checking the empty vendor id now
	//Part of 3.2.3 & 3.2.8
	vendorInInvoice, fetchMsg := vmd.GetVendor(stub, invoice.DcDocumentData.DcHeader.ErpSystem, vendorID, invoice.DcDocumentData.DcHeader.Client)
	if len(fetchMsg) > 0 {
		//3.2.3
		//Vendor does not exists
		internalStatus = _lc_INTERNAL_STATUS_MOVE_TO_REMITTOID
		status = CONTINUE
		addlInfo := AdditionalInfo{Value: "Vendor name does not exists in VMD . Exiting 3.2.3"}
		invStat, _ = SetInvoiceStatus(stub, trxnContext, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, status, "", "Vendor name does not exists in VMD", internalStatus, addlInfo)
		_lc_vna_logger.Debugf("Vendor name does not exists in VMD . Exiting 3.2.3")

		return 1, errorMessage, invStat
	}

	for _, ponumber := range collectPONumbers(invoice) {
		//3.2.4
		po, _ := po.GetPO(stub, []string{invoice.DcDocumentData.DcHeader.ErpSystem, ponumber, invoice.DcDocumentData.DcHeader.Client})
		//Assumption is PO does exist

		//3.2.5 on wards
		isOkToProceed, status, internalStatus, errorMessage := matchVendor(stub, po, vendorID, vendorInInvoice, invoice)
		if !isOkToProceed {
			//This is IBM_AP_ACTION or REJECT CASE
			invStat, _ = SetInvoiceStatus(stub, trxnContext, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, status, errorMessage, "", internalStatus, EMPTY_ADDITIONAL_INFO)
			_lc_vna_logger.Debugf("Vendor does not exist . Status setting to %s due to PO %s", status, ponumber)
			return 2, "", invStat
		}
	}
	addlInfo := AdditionalInfo{Value: "All invoice line items exited with CONTINUE status. Exiting from end"}
	//This is the continue case
	invStat, _ = SetInvoiceStatus(stub, trxnContext, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, CONTINUE, "", "", _lc_INTERNAL_STATUS_MOVE_TO_REMITTOID, addlInfo)
	_lc_vna_logger.Debugf("Vendor exists . Moving to next step")
	return 1, "", invStat
}

//SaveVNADynamicUIRouteConfig configurs the dynamic configuration of auto rejection
//parameter for Vendor Name Authentication flow
//Input JSON structure to add a configuration is
//	{
//		"compCode":"BCBCBC",
//		"invoiceType":"XYZ",
//		"inputSource":"upload",
//      "uiRoute":"AWAITING IBM AP ACTION"
//		"internalStatus": "st-vendNameAuth-1|st-remitToIid-1"
//	}
func SaveVNADynamicUIRouteConfig(stub shim.ChaincodeStubInterface) pb.Response {
	validInputSpec := `
	{
		"compCode":{ "type":"string","isMandatory":true },
		"invoiceType":{ "type":"string","isMandatory":true },
		"inputSource":{ "type":"string","isMandatory":true },
		"uiRoute":{ "type":"string","isMandatory":true },
		"internalStatus": { "type":"string","isMandatory":true }
	}
	`
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		_lc_vna_logger.Errorf("No argument provided to SaveVNADynamicUIRouteConfig ")
		return shim.Error(buildResponseMessage(false, "No argument provided to SaveVNADynamicUIRouteConfig"))
	}
	dataToSave := args[0]
	specification := buildSpec(validInputSpec)
	if !checkValidFiledsInMap([]byte(dataToSave), specification) {
		_lc_vna_logger.Errorf("Invalid configuration provided for SaveVNADynamicUIRouteConfig ")
		return shim.Error(buildResponseMessage(false, "Invalid configuration provided for SaveVNADynamicUIRouteConfig"))
	}

	return saveRecordInDynatable(stub, _lc_DYNA_CONFIG_VNA_UI_ROUTE, args[0])
}

//GetVNADynamicUIRouteConfig Returns the UI route config stored for a company code,
//invoice type and input source combination
//Input JSON structure to fetch in the infomration is
//	{
//		"compCode":"BCBCBC",
//		"invoiceType":"XYZ",
//		"inputSource":"upload"
//	}
func GetVNADynamicUIRouteConfig(stub shim.ChaincodeStubInterface) pb.Response {
	validInputSpec := `
	{
		"compCode":{ "type":"string","isMandatory":true },
		"invoiceType":{ "type":"string","isMandatory":true },
		"inputSource":{ "type":"string","isMandatory":true }
	}
	`
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		_lc_vna_logger.Errorf("No argument provided to GetVNADynamicUIRouteConfig ")
		return shim.Error(buildResponseMessage(false, "No argument provided to GetVNADynamicUIRouteConfig"))
	}
	queryParams := args[0]
	specification := buildSpec(validInputSpec)
	if !checkValidFiledsInMap([]byte(queryParams), specification) {
		_lc_vna_logger.Errorf("Invalid configuration provided for GetVNADynamicUIRouteConfig ")
		return shim.Error(buildResponseMessage(false, "Invalid configuration provided for GetVNADynamicUIRouteConfig"))
	}

	return retriveRecordInDynatable(stub, _lc_DYNA_CONFIG_VNA_UI_ROUTE, queryParams)
}

//SaveVNADynamicAutoRejConfig configurs the dynamic configuration of auto rejection
//parameter for Vendor Name Authentication flow
//Input JSON structure to add a configuration is
//	{
//		"compCode":"BCBCBC",
//		"invoiceType":"XYZ",
//		"inputSource":"upload",
//		"autoRejection": true| false //Please note that , this is boolean
//	}
func SaveVNADynamicAutoRejConfig(stub shim.ChaincodeStubInterface) pb.Response {
	validInputSpec := `
	{
		"compCode":{ "type":"string","isMandatory":true },
		"invoiceType":{ "type":"string","isMandatory":true },
		"inputSource":{ "type":"string","isMandatory":true },
		"autoRejection":{ "type":"bool","isMandatory":true }
	}
	`
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		_lc_vna_logger.Errorf("No argument provided to SaveVNADynamicAutoRejConfig ")
		return shim.Error(buildResponseMessage(false, "No argument provided to SaveVNADynamicAutoRejConfig"))
	}
	dataToSave := args[0]
	specification := buildSpec(validInputSpec)
	if !checkValidFiledsInMap([]byte(dataToSave), specification) {
		_lc_vna_logger.Errorf("Invalid configuration provided for SaveVNADynamicAutoRejConfig ")
		return shim.Error(buildResponseMessage(false, "Invalid configuration provided for SaveVNADynamicAutoRejConfig"))
	}

	return saveRecordInDynatable(stub, _lc_DYNA_CONFIG_VNA_AUTO_REJECT, args[0])
}

//GetVNADynamicAutoRejConfig Returns the auto rejection config stored for a company code,
//invoice type and input source combination
//Input JSON structure to fetch in the infomration is
//	{
//		"compCode":"BCBCBC",
//		"invoiceType":"XYZ",
//		"inputSource":"upload"
//	}
func GetVNADynamicAutoRejConfig(stub shim.ChaincodeStubInterface) pb.Response {
	validInputSpec := `
	{
		"compCode":{ "type":"string","isMandatory":true },
		"invoiceType":{ "type":"string","isMandatory":true },
		"inputSource":{ "type":"string","isMandatory":true }
	}
	`
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		_lc_vna_logger.Errorf("No argument provided to GetVNADynamicAutoRejConfig ")
		return shim.Error(buildResponseMessage(false, "No argument provided to GetVNADynamicAutoRejConfig"))
	}
	queryParams := args[0]
	specification := buildSpec(validInputSpec)
	if !checkValidFiledsInMap([]byte(queryParams), specification) {
		_lc_vna_logger.Errorf("Invalid configuration provided for GetVNADynamicAutoRejConfig ")
		return shim.Error(buildResponseMessage(false, "Invalid configuration provided for GetVNADynamicAutoRejConfig"))
	}

	return retriveRecordInDynatable(stub, _lc_DYNA_CONFIG_VNA_AUTO_REJECT, queryParams)
}

//SaveVNADynamicBusAsListConfig configurs the dynamic configuration of auto rejection
//parameter for Vendor Name Authentication flow
//Input JSON structure to add a configuration is
//	{
//		"compCode":"BCBCBC",
//		"invoiceType":"XYZ",
//		"inputSource":"upload",
//		"IBM": [" Wipro","Wipro International"],//Names are case insentive and spaces are ignored.
// 		"ABC Corp": ["XYZ Corp","Liuagong"," zercob Inc"] //Names are case insentive and spaces are ignored.
//	}
func SaveVNADynamicBusAsListConfig(stub shim.ChaincodeStubInterface) pb.Response {
	validInputSpec := `
	{
		"compCode":{ "type":"string","isMandatory":true },
		"invoiceType":{ "type":"string","isMandatory":true },
		"inputSource":{ "type":"string","isMandatory":true }
	}
	`
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		_lc_vna_logger.Errorf("No argument provided to SaveVNADynamicBusAsListConfig ")
		return shim.Error(buildResponseMessage(false, "No argument provided to SaveVNADynamicBusAsListConfig"))
	}
	dataToSave := args[0]
	specification := buildSpec(validInputSpec)
	if !checkValidFiledsInMap([]byte(dataToSave), specification) {
		_lc_vna_logger.Errorf("Invalid configuration provided for SaveVNADynamicBusAsListConfig ")
		return shim.Error(buildResponseMessage(false, "Invalid configuration provided for SaveVNADynamicBusAsListConfig"))
	}

	return saveRecordInDynatable(stub, _lc_DYNA_CONFIG_VNA_BUS_AS_LIST, args[0])
}

//GetVNADynamicBusAsListConfig Returns the auto business as list stored for a company code,
//invoice type and input source combination
//Input JSON structure to fetch in the infomration is
//	{
//		"compCode":"BCBCBC",
//		"invoiceType":"XYZ",
//		"inputSource":"upload"
//	}
func GetVNADynamicBusAsListConfig(stub shim.ChaincodeStubInterface) pb.Response {
	validInputSpec := `
	{
		"compCode":{ "type":"string","isMandatory":true },
		"invoiceType":{ "type":"string","isMandatory":true },
		"inputSource":{ "type":"string","isMandatory":true }
	}
	`
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		_lc_vna_logger.Errorf("No argument provided to GetVNADynamicBusAsListConfig ")
		return shim.Error(buildResponseMessage(false, "No argument provided to GetVNADynamicBusAsListConfig"))
	}
	queryParams := args[0]
	specification := buildSpec(validInputSpec)
	if !checkValidFiledsInMap([]byte(queryParams), specification) {
		_lc_vna_logger.Errorf("Invalid configuration provided for GetVNADynamicBusAsListConfig ")
		return shim.Error(buildResponseMessage(false, "Invalid configuration provided for GetVNADynamicBusAsListConfig"))
	}

	return retriveRecordInDynatable(stub, _lc_DYNA_CONFIG_VNA_BUS_AS_LIST, queryParams)
}

//isVendNameAuthAutoRej return the config. If not found then returns the false
func isVendNameAuthAutoRej(stub shim.ChaincodeStubInterface, invoice Invoice) bool {
	keyToSearch := buildKeyForDyamicTable(_lc_DYNA_CONFIG_VNA_AUTO_REJECT, invoice.DcDocumentData.DcHeader.CompanyCode, invoice.DcDocumentData.DcHeader.InvType, invoice.DcDocumentData.DcHeader.DocSource)
	configJSON, err := stub.GetState(keyToSearch)
	if err == nil && len(configJSON) > 0 {
		config := make(map[string]bool)
		//I am not checking the config parsing as data to be stored would be a validated one
		json.Unmarshal(configJSON, &config)
		if value, isOk := config["autoRejection"]; isOk {
			return value
		}
	}
	return false
}

//getUIActionConfigForVendNameAuth retrives the dyna table config for the invoice if exists
//Returns INV_STATUS_PENDING_AP, _lc_INTERNAL_STATUS_MOVE_TO_VENDNAMEAUTH if not found
func getUIActionConfigForVendNameAuth(stub shim.ChaincodeStubInterface, invoice Invoice) (string, string) {

	keyToSearch := buildKeyForDyamicTable(_lc_DYNA_CONFIG_VNA_UI_ROUTE, invoice.DcDocumentData.DcHeader.CompanyCode, invoice.DcDocumentData.DcHeader.InvType, invoice.DcDocumentData.DcHeader.DocSource)
	configJSON, err := stub.GetState(keyToSearch)
	if err == nil && len(configJSON) > 0 {
		config := make(map[string]string)
		//I am not checking the config parsing as data to be stored would be a validated one
		json.Unmarshal(configJSON, &config)
		intStatus, isOk := config["internalStatus"]
		status, isRouteExists := config["uiRoute"]
		if isRouteExists && isOk {
			return status, intStatus
		}
	}

	return INV_STATUS_PENDING_AP, _lc_INTERNAL_STATUS_MOVE_TO_VENDNAMEAUTH
}

//Match the invoice vendor name with dyna table based business as list
func matchValidBusinessAsList(stub shim.ChaincodeStubInterface, invoice Invoice) bool {
	isMatchFound := false
	keyToSearch := buildKeyForDyamicTable(_lc_DYNA_CONFIG_VNA_BUS_AS_LIST, invoice.DcDocumentData.DcHeader.CompanyCode, invoice.DcDocumentData.DcHeader.InvType, invoice.DcDocumentData.DcHeader.DocSource)
	_lc_vna_logger.Debugf("BUS LIST KEY |%s| ", keyToSearch)
	configJSON, err := stub.GetState(keyToSearch)
	_lc_vna_logger.Debugf("BUS LIST Data %s ", string(configJSON))

	if err == nil && len(configJSON) > 0 {
		config := make(map[string]interface{})
		//I am not checking the config parsing as data to be stored would be a validated one
		json.Unmarshal(configJSON, &config)
		for key, value := range config {
			isMatchFound = matchVendorName(invoice.DcDocumentData.DcSwissHeader.SupplierName, key)
			if isMatchFound {
				break
			}
			//Check in the value
			switch value.(type) {
			case string:
				valueToMatch, _ := value.(string)
				isMatchFound = matchVendorName(valueToMatch, invoice.DcDocumentData.DcSwissHeader.SupplierName)
			case []string:
				listOfVendorNames, _ := value.([]string)
				isMatchFound = findInList(listOfVendorNames, invoice.DcDocumentData.DcSwissHeader.SupplierName)
			}
			if isMatchFound {
				break
			}
		}

	}
	return isMatchFound
}

//Find vendor name in the list
func findInList(list []string, vendorName string) bool {
	for _, item := range list {
		if matchVendorName(item, vendorName) {
			return true
		}
	}
	return false
}

//Returns true is every thing matches per specification This is a single level validator
func checkValidFiledsInMap(dataJSON []byte, spec string) bool {
	isValid := true
	dataMap := make(map[string]interface{})
	specMap := make(map[string]FieldDataSpec)
	err := json.Unmarshal(dataJSON, &dataMap)
	if err != nil {
		_lc_vna_logger.Errorf("Unable to parse data json", err)
		return false
	}
	err = json.Unmarshal([]byte(spec), &specMap)
	if err != nil {
		_lc_vna_logger.Errorf("Unable to specification data json", err)
		return false
	}
	for filedName, fieldSpec := range specMap {
		data, isExisting := dataMap[filedName]
		if !isExisting && fieldSpec.IsMandatory {
			isValid = false
			break
		}
		if !fieldSpec.ValidateType(data) {
			isValid = false
			break
		}
	}
	return isValid
}

//Try to match vendor name as per the specification.
func matchVendor(stub shim.ChaincodeStubInterface, po po.PO, vendorID string, vendorInInvoice vmd.Vendor, invoice Invoice) (bool, string, string, string) {
	//3.2.5
	if strings.TrimSpace(po.VendorId) == vendorID {
		//Move to remit to id
		return true, CONTINUE, _lc_INTERNAL_STATUS_MOVE_TO_REMITTOID, ""
	}
	//3.2.6 Use the po.VendorId to retrive vendor from VMD
	vendorFrmPO, errVendorFetch := vmd.GetVendor(stub, po.ERPSystem, po.VendorId, po.Client)
	if len(errVendorFetch) > 0 {
		//Vendor mentioned in PO does not exist in Vendor Master

		return false, REJECT, _lc_INTERNAL_STATUS_MOVE_TO_VENDNAMEAUTH, "VENDOR MENTIONED IN PO DOES NOT EXIST IN VENDOR MASTER"
	}

	if matchVendorName(vendorFrmPO.VendorName, vendorInInvoice.VendorName) {
		//3.2.8
		_lc_vna_logger.Debugf("Vendor name matched %s ", vendorInInvoice.VendorName)
		return true, CONTINUE, _lc_INTERNAL_STATUS_MOVE_TO_REMITTOID, ""
	}
	if matchVendorName(vendorFrmPO.VendorName, invoice.DcDocumentData.DcSwissHeader.SupplierName) {
		//3.2.10 && 3.2.11
		_lc_vna_logger.Debugf("Vendor name matched with invoice supplier name %s ", invoice.DcDocumentData.DcSwissHeader.SupplierName)
		return true, CONTINUE, _lc_INTERNAL_STATUS_MOVE_TO_REMITTOID, ""
	}
	if matchValidBusinessAsList(stub, invoice) {
		//3.2.14
		_lc_vna_logger.Debugf("Matched with business as list")
		return true, CONTINUE, _lc_INTERNAL_STATUS_MOVE_TO_REMITTOID, ""
	}
	if isVendNameAuthAutoRej(stub, invoice) {
		//3.2.15
		return false, REJECT, _lc_INTERNAL_STATUS_MOVE_TO_VENDNAMEAUTH, "INVALID INVOICING PARTY"
	}
	//3.2.16
	status, internalStatus := getUIActionConfigForVendNameAuth(stub, invoice)
	return false, status, internalStatus, "VENDOR NAME AUTHENTICATION FAILED"

}

//matchVendorName matches 2 vendor names all whitespaces removed and case insensitive
func matchVendorName(name1, name2 string) bool {
	normalizedName1 := strings.Replace(name1, " ", "", -1)
	normalizedName2 := strings.Replace(name2, " ", "", -1)
	isEqual := strings.EqualFold(normalizedName1, normalizedName2)
	return isEqual
}

//Retrives the unique po numbers from the invoice lines
func collectPONumbers(invoice Invoice) []string {

	uniquePoNumbers := make(map[string]bool)
	for _, lineItem := range invoice.DcDocumentData.DcLines {
		uniquePoNumbers[lineItem.PoNumber] = true
	}
	poNumberList := make([]string, 0)
	for ponumber := range uniquePoNumbers {
		poNumberList = append(poNumberList, ponumber)
	}
	return poNumberList

}

//buildSpec build a string from specification template
func buildSpec(tmplContent string) string {
	dummyMap := make(map[string]string)
	tmpl, _ := template.New("specs").Parse(tmplContent)
	var invoiceBytes bytes.Buffer
	tmpl.Execute(&invoiceBytes, dummyMap)
	return string(invoiceBytes.Bytes())
}

//buildKeyForDyamicTable builds the dynamic table key field from the inputs
func buildKeyForDyamicTable(prefix, compCode, invType, inputSrc string) string {
	key := fmt.Sprintf("%s_%s_%s_%s", prefix, strings.TrimSpace(compCode), strings.TrimSpace(invType), strings.TrimSpace(inputSrc))
	return key
}

//saveRecordInDynatable builds the dynamic table key field from the inputs
func saveRecordInDynatable(stub shim.ChaincodeStubInterface, prefix, configJSON string) pb.Response {
	//Input is already tested . So no chance of error
	config := make(map[string]interface{})
	json.Unmarshal([]byte(configJSON), &config)
	compCode, _ := config["compCode"].(string)
	invType, _ := config["invoiceType"].(string)
	inputSrc, _ := config["inputSource"].(string)
	key := buildKeyForDyamicTable(prefix, compCode, invType, inputSrc)
	_lc_vna_logger.Debugf("Save key %s|", key)
	saveBytes, err := json.Marshal(config)
	if err != nil {
		return shim.Error(buildResponseMessage(false, "Unable to marshal json before saving to dyna table "+key))
	}
	err = stub.PutState(key, saveBytes)
	if err != nil {
		return shim.Error(buildResponseMessage(false, "Error in saving to dyna table"+key))
	}
	return shim.Success([]byte(buildResponseMessage(true, "Dynatable change updated successfully for "+key)))
}

//retriveRecordInDynatable returns the configuration saved
func retriveRecordInDynatable(stub shim.ChaincodeStubInterface, prefix, configJSON string) pb.Response {
	//Input is already tested . So no chance of error
	config := make(map[string]interface{})
	json.Unmarshal([]byte(configJSON), &config)
	compCode, _ := config["compCode"].(string)
	invType, _ := config["invoiceType"].(string)
	inputSrc, _ := config["inputSource"].(string)
	key := buildKeyForDyamicTable(prefix, compCode, invType, inputSrc)

	savedConfig, err := stub.GetState(key)
	if err != nil {
		return shim.Error(buildResponseMessage(false, "Error in retriveal from dyna table "+key))
	}
	return shim.Success(savedConfig)
}

//FieldDataSpec defines the specification for a field
type FieldDataSpec struct {
	DataType    string `json:"type"`
	IsMandatory bool   `json:"isMandatory"`
}

//ValidateType validates the data as per specification
func (fds FieldDataSpec) ValidateType(data interface{}) bool {
	isDataOk := false
	switch fds.DataType {
	case "string":
		_, isDataOk = data.(string)
	case "int":
		_, isDataOk = data.(int)
	case "arrayInt":
		_, isDataOk = data.([]int)
	case "arraySting":
		_, isDataOk = data.([]string)
	case "arrayObjets":
		_, isDataOk = data.([]interface{})
	case "mapOfObjects":
		_, isDataOk = data.(map[string]interface{})
	case "mapOfString":
		_, isDataOk = data.(map[string]string)
	case "object":
		_, isDataOk = data.(interface{})
	case "bool":
		_, isDataOk = data.(bool)
	}
	return isDataOk
}
