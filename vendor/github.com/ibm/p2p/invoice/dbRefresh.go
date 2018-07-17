package invoice

import (
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	util "github.com/ibm/p2p"
)

//var myLogger = logging.MustGetLogger("Procure-To-Pay util")

func ResubmitAfterDbRefresh(stub shim.ChaincodeStubInterface, status string) pb.Response {

	var invToSubmit InvoicesToReSubmit
	var resubmitDetails InvoiceSubmitRequest
	myLogger.Debugf("Inside resubmit After DB refresh=========================", status)
	//invoiceArr = GetInvoiceByStatus(stub,status)//"WAITING DB REFRESH FOR PO"
	invoiceKeys := FetchInvoiceByStatus(stub, status)
	myLogger.Debugf("size of invoice arr=========================", len(invoiceKeys), "====================", invoiceKeys[0])

	for _, invoiceKey := range invoiceKeys {
		var resubmitArr []InvoiceSubmitRequest
		primaryKeys := strings.Split(invoiceKey, "~")
		myLogger.Debugf("Primary keys are =========================", invoiceKey, "==========", len(primaryKeys), "====VALUE==", primaryKeys[0])
		if len(primaryKeys) == 2 {
			myLogger.Debugf("Length of keys 2=========================")
			invStatArr, _ := GetInvoiceStatus(stub, []string{primaryKeys[0], primaryKeys[1]})
			myLogger.Debugf("size of invoice status arr=========================", len(invStatArr))

			invoiceStatus := invStatArr[len(invStatArr)-1]

			resubmitDetails.BciId = invoiceStatus.BciId
			//resubmitDetails.InvoiceNumber = invoice.DcDocumentData.DcHeader.InvoiceNumber
			resubmitDetails.ScanID = invoiceStatus.ScanID
			resubmitDetails.Status = invoiceStatus.Status
			resubmitDetails.ReasonCode = invoiceStatus.ReasonCode
			resubmitDetails.Comments = invoiceStatus.Comments
			resubmitDetails.BuyerId = invoiceStatus.BuyerId
			resubmitDetails.BuyerEmailId = invoiceStatus.BuyerEmailId
			resubmitDetails.UserID = invoiceStatus.UserId
			resubmitArr = append(resubmitArr, resubmitDetails)
		}
		invToSubmit.InvoicesToSubmit = resubmitArr
		//util.MarshalToBytes(invToSubmit)
		myLogger.Debugf("Invoice to submit list without marshall =========================", invToSubmit)
		myLogger.Debugf("Invoice to submit list After Marshall  =========================", util.MarshalToBytes(invToSubmit))
		myLogger.Debugf("Invoice to submit list After marshall and string conversion =========================", string(util.MarshalToBytes(invToSubmit)))

		subInvrequest := string(util.MarshalToBytes(resubmitDetails))

		response := SubmitInvoice(stub, subInvrequest)
		myLogger.Debugf("Response from Submit Invoice =========================", response)
	}
	return shim.Success(nil)
}
