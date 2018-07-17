package invoice

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	util "github.com/ibm/p2p"
)

//Definition of local constants
const _lc_INTERNAL_STATUS_MOVE_NEXTSTEP = "st-singlevsmultiple-1"
const _lc_INTERNAL_STATUS_VENDOR_ACTION_REQUIRED = "st-bciValidation-1"
const _lc_STATUS_REJECT_BACK_TO_VENDOR = REJECT
const _lc_STATUS_MOVE_NEXTSTEP = CONTINUE
const _lc_STATUS_IBM_AP = INV_STATUS_PENDING_AP

//Trucattion direction
const _lc_TRUCATION_LTOR = "LEFT_TO_RIGHT"
const _lc_TRUCATION_RTOL = "RIGHT_TO_LEFT"
const _lc_BCI_CONFIG_KEY = "BC_CONFIG_KEY"

//BCIValidationConfig represents configuration to control the invoice number/date valiation
type BCIValidationConfig struct {
	MaxInvoiceDigits           int    `json:"maxInvoiceNumberDigits"`
	TrucationDirection         string `json:"truncateDirection"`
	RemoveLeadingZero          bool   `json:"removeLeadingZero"`
	RemoveSpecialChars         bool   `json:"removeSpecialChars"`
	FutueInvDateAutoRejecttion bool   `json:"futureInvDateAutoRej"`
	FutureInvDaysAllowed       int    `json:"futureInvDaysAllowed"`
	MaxPastInvoiceDays         int    `json:"maxPastInvoiceDays"`
	CompanyCode                string `json:"compCode"`
	InvoiceType                string `json:"invoiceType"`
	InputSource                string `json:"inputSource"`
}

//BCIValidationConfigUpdate assumes the input to be in the following json format
/*
	{
		"compCode":"BCBCBC",
		"invoiceType":"XYZ",
		"inputSource":"upload",
		"maxInvoiceNumberDigits": 10,
		"truncateDirection":"LEFT_TO_RIGHT| RIGHT_TO_LEFT",
		"removeLeadingZero":true,
		"removeSpecialChars":true,
		"futureInvDateAutoRej": true,
		"futureInvDaysAllowed": 365
		"maxPastInvoiceDays": 365
	}
*/
func BCIValidationConfigUpdate(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	bciConfig := BCIValidationConfig{}
	//Checking the first argument
	configJSON := args[0]
	err := json.Unmarshal([]byte(configJSON), &bciConfig)
	if len(configJSON) == 0 || err != nil {
		return shim.Error(buildResponseMessage(false, "Empty or wrong configuration update given"))
	}
	if isOk, msg := validateBCIConfigData(bciConfig); !isOk {
		return shim.Error(buildResponseMessage(false, msg))
	}
	configJSONToStore, _ := json.Marshal(bciConfig)
	keyToStore := fmt.Sprintf("%s_%s_%s_%s", _lc_BCI_CONFIG_KEY, strings.TrimSpace(bciConfig.CompanyCode), strings.TrimSpace(bciConfig.InvoiceType), strings.TrimSpace(bciConfig.InputSource))
	wsUpdErr := stub.PutState(keyToStore, configJSONToStore)
	if wsUpdErr != nil {
		return shim.Error(buildResponseMessage(false, fmt.Sprintf("%v", wsUpdErr)))
	}
	return shim.Success([]byte(buildResponseMessage(true, "Configuration updated successfully")))

}

//validateBCIConfigData validates the BCI config data before placing them to the world state
func validateBCIConfigData(config BCIValidationConfig) (bool, string) {
	if config.MaxInvoiceDigits == 0 {
		return false, "Invalid number of invoice digits"
	}
	if config.TrucationDirection != _lc_TRUCATION_LTOR && config.TrucationDirection != _lc_TRUCATION_RTOL {
		return false, "Invalid truncateDirection "
	}
	if config.FutueInvDateAutoRejecttion && config.FutureInvDaysAllowed <= 0 {
		return false, "Invalid future invoice rejection days "
	}
	if config.MaxPastInvoiceDays <= 0 {
		return false, "Invalid past invoice rejection days "
	}
	if !isNonEmpty(config.CompanyCode) {
		return false, "Invalid company code "
	}
	if !isNonEmpty(config.InvoiceType) {
		return false, "Invalid invoice type "
	}
	if !isNonEmpty(config.InputSource) {
		return false, "Invalid input source "
	}
	return true, ""
}

//buildResponseMessage creates a json for returning from the smart contract
func buildResponseMessage(isSuccess bool, msg string) string {
	return fmt.Sprintf("{\"isSucess\":\"%v\" , \"message\":\"%s\"}", isSuccess, msg)
}

//BCIValidation performs the business of BCI validation on the input invoice
func BCIValidation(stub shim.ChaincodeStubInterface, trxnContext *Context) (int, string, InvoiceStatus) {
	internalStatus := ""
	errorMessage := ""
	status := ""
	reasonCode := ""

	var invStat InvoiceStatus
	invoice := trxnContext.Invoice
	if !isNonEmpty(invoice.DcDocumentData.DcHeader.VendorID) {

		internalStatus = _lc_INTERNAL_STATUS_VENDOR_ACTION_REQUIRED
		status = _lc_STATUS_REJECT_BACK_TO_VENDOR
		errorMessage = "REMIT TO VENDOR ID NOT AVAILABLE"
		reasonCode = "BCI_001"
	} else if !isNonEmpty(invoice.DcDocumentData.DcSwissHeader.SupplierName) {
		internalStatus = _lc_INTERNAL_STATUS_VENDOR_ACTION_REQUIRED
		status = _lc_STATUS_REJECT_BACK_TO_VENDOR
		errorMessage = "REMIT TO VENDOR NAME NOT AVAILABLE"
		reasonCode = "BCI_002"
	} else if isOk, inernalStat, errMsg := validateInvoiceNumber(stub, &invoice); !isOk {
		internalStatus = inernalStat
		status = _lc_STATUS_REJECT_BACK_TO_VENDOR
		errorMessage = errMsg
		reasonCode = "BCI_003"
	} else if !isNonEmpty(invoice.DcDocumentData.DcHeader.CurrencyCode) {
		internalStatus = _lc_INTERNAL_STATUS_VENDOR_ACTION_REQUIRED
		status = _lc_STATUS_REJECT_BACK_TO_VENDOR
		errorMessage = "INVOICE CURRENCY  NOT AVAILABLE"
		reasonCode = "BCI_004"
	} else if isOk, inernalStat, rCode, errMsg := validateInvoiceDate(invoice.DcDocumentData.DcHeader.DocDate); !isOk {
		internalStatus = inernalStat
		status = _lc_STATUS_IBM_AP
		errorMessage = errMsg
		reasonCode = rCode
	} else if !isNonEmpty(invoice.DcDocumentData.DcSwissHeader.BuyerName) {
		internalStatus = _lc_INTERNAL_STATUS_VENDOR_ACTION_REQUIRED
		status = _lc_STATUS_REJECT_BACK_TO_VENDOR
		errorMessage = "BILL TO ADDRESS / COMPANY CODE NAME NOT AVAILABLE"
		reasonCode = "BCI_008"
	} else if !isNonEmpty(invoice.DcDocumentData.DcHeader.ScanID) {
		internalStatus = _lc_INTERNAL_STATUS_VENDOR_ACTION_REQUIRED
		status = _lc_STATUS_REJECT_BACK_TO_VENDOR
		errorMessage = "REPLICATE THE E-MAIL TEMPLATE CURRENTLY FOLLOWED IN DCIW"
		reasonCode = "BCI_009"
	} else {
		for _, lineItem := range invoice.DcDocumentData.DcLines {
			if isLineOk, inernalStat, errMsg := validateLineID(lineItem); !isLineOk {
				internalStatus = inernalStat
				status = _lc_STATUS_REJECT_BACK_TO_VENDOR
				errorMessage = errMsg
				reasonCode = "BCI_006"
				break
			}
		}
	}
	if len(reasonCode) > 0 {
		invStat, _ = SetInvoiceStatus(stub, trxnContext, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, status, errorMessage, "", internalStatus, EMPTY_ADDITIONAL_INFO)
		/*statCode := 0
		if status == _lc_STATUS_IBM_AP {
			statCode = 2
		}*/
		return 2, "", invStat
	}
	invStat, _ = SetInvoiceStatus(stub, trxnContext, invoice.BCIID, invoice.DcDocumentData.DcHeader.ScanID, _lc_STATUS_MOVE_NEXTSTEP, "", "", _lc_INTERNAL_STATUS_MOVE_NEXTSTEP, EMPTY_ADDITIONAL_INFO)
	return 1, "", invStat

}

//validateLineID validates a given line item in DCLine
func validateLineID(lineDetails DCLine) (bool, string, string) {
	isValid := true

	internalStatus := _lc_INTERNAL_STATUS_VENDOR_ACTION_REQUIRED
	errorMessage := ""
	if !isTaxLine(lineDetails) {
		if !isNonEmpty(lineDetails.Description) {
			isValid = false
			errorMessage = "LINE ITEM DESCRIPTION  NOT AVAILABLE"
		} else if lineDetails.Quantity <= 0.0 {
			isValid = false
			errorMessage = "QUANTITY  NOT AVAILABLE"
		} else if lineDetails.UnitPrice <= 0.0 {
			isValid = false
			errorMessage = "UNIT PRICE  NOT AVAILABLE"
		} else if lineDetails.Amount <= 0.0 {
			isValid = false
			errorMessage = "LINE AMOUNT (QTY  X PRICE)  NOT AVAILABLE"
		}
	}
	return isValid, internalStatus, errorMessage
}

//isTaxLine checks if a given DCLine is a tax line or not
func isTaxLine(lineDetals DCLine) bool {
	if !isNonEmpty(lineDetals.MatNumber) && lineDetals.TaxAmount > 0.0 && lineDetals.TaxPercent > 0.0 {
		return true
	}
	return false
}

//validateInvoiceNumber validates the invoice number based on configuration
func validateInvoiceNumber(stub shim.ChaincodeStubInterface, invoice *Invoice) (bool, string, string) {
	internalStatus := ""
	errorMessage := ""

	if !isNonEmpty(invoice.DcDocumentData.DcHeader.InvoiceNumber) {
		internalStatus = _lc_INTERNAL_STATUS_VENDOR_ACTION_REQUIRED
		errorMessage = "INVOICE NUMBER"
		return false, internalStatus, errorMessage
	}

	//Process the invoice number
	invoiceNumber := invoice.DcDocumentData.DcHeader.InvoiceNumber
	bciConfig := getBCIValidationConfig(stub, invoice)
	//if length is more then run trucation
	if len(invoiceNumber) > bciConfig.MaxInvoiceDigits {
		switch bciConfig.TrucationDirection {
		case _lc_TRUCATION_LTOR:
			invoiceNumber = invoiceNumber[0 : bciConfig.MaxInvoiceDigits-1]
		case _lc_TRUCATION_RTOL:
			invoiceNumber = invoiceNumber[len(invoiceNumber)-bciConfig.MaxInvoiceDigits : bciConfig.MaxInvoiceDigits]
		}
	}
	//remove leading zero
	if bciConfig.RemoveLeadingZero {
		invoiceNumber = strings.TrimLeft(invoiceNumber, "0")
	}
	//remove special charaters
	invoice.DcDocumentData.DcHeader.InvoiceNumber = invoiceNumber

	return true, internalStatus, errorMessage

}

//getBCIValidationConfig fetches BCI Config data
func getBCIValidationConfig(stub shim.ChaincodeStubInterface, invoice *Invoice) BCIValidationConfig {

	keyToStore := fmt.Sprintf("%s_%s_%s_%s", _lc_BCI_CONFIG_KEY, strings.TrimSpace(invoice.DcDocumentData.DcHeader.CompanyCode), strings.TrimSpace(invoice.DcDocumentData.DcHeader.InvType), strings.TrimSpace(invoice.DcDocumentData.DcHeader.DocSource))
	config, err := stub.GetState(keyToStore)
	if err != nil || len(config) == 0 {
		//Returning a default config
		return BCIValidationConfig{MaxInvoiceDigits: 10, FutueInvDateAutoRejecttion: false, TrucationDirection: _lc_TRUCATION_LTOR, RemoveLeadingZero: false}
	}
	var bciConfig BCIValidationConfig
	json.Unmarshal(config, &bciConfig)
	return bciConfig

}

//validateInvoiceDate  performs the validation on invoice date:
func validateInvoiceDate(date util.BCDate) (bool, string, string, string) {

	diff := time.Since(date.Time())
	days := diff.Hours() / 24
	if days > 366 {
		return false, _lc_INTERNAL_STATUS_VENDOR_ACTION_REQUIRED, "BCI_007", "INVOICE DATE OR DELIVERY DATE IS OLDER THAN 1 YEAR"
	}
	return true, _lc_INTERNAL_STATUS_VENDOR_ACTION_REQUIRED, "", ""

}

//isNonEmpty checks if a string is not empty or not
func isNonEmpty(input string) bool {
	if len(strings.TrimSpace(input)) > 0 {
		return true
	}
	return false
}
