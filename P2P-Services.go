package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"

	util "github.com/ibm/p2p"
	"github.com/ibm/p2p/audit"
	"github.com/ibm/p2p/companyCode"
	"github.com/ibm/p2p/email"
	"github.com/ibm/p2p/grn"
	"github.com/ibm/p2p/invoice"
	"github.com/ibm/p2p/po"
	refTable "github.com/ibm/p2p/refTable"
	"github.com/ibm/p2p/report"
	"github.com/ibm/p2p/vmd"
	logging "github.com/op/go-logging"
)

var myLogger = logging.MustGetLogger("Procure-To-Pay")

type ServicesChaincode struct {
}

func (t *ServicesChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *ServicesChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	//invoice.InitiatizeProcessStepLookups()

	switch function {

	/* ======== Buyer Functions ===========*/
	case "ADD_BUYERS":
		// Called as part of blockchain Initial Load
		return email.AddEmailRecords(stub, args[0])
	case "GET_BUYER":
		email, _ := email.GetEmail(stub, args[0])
		return shim.Success(util.MarshalToBytes(email))
	case "GET_BUYERS":
		buyers := email.GetEmailRecords(stub)
		return shim.Success(util.MarshalToBytes(buyers))

		/* ======== PO Functions ===========*/
	case "ADD_POS":
		// Called as part of BCI Cron job Bulk Load
		return po.AddPoRecords(stub, args[0])
	case "ADD_POLINEITEMS":
		// Called as part of BCI Cron job Full Load
		return po.AddPoLineItemRecords(stub, args[0])
	case "GET_PO":
		PO, _ := po.GetPO(stub, []string{args[0], args[1], args[2]})
		return shim.Success(util.MarshalToBytes(PO))
	case "GET_PO_LINEITEM":
		POLine, _ := po.GetPOLineItem(stub, []string{args[0], args[1], args[2], args[3]})
		return shim.Success(util.MarshalToBytes(POLine))
	case "GET_ALL_POLINEITEMS_BY_PO":
		POLines := po.GetAllPOLineItemsByPO(stub, []string{args[0], args[1]})
		return shim.Success(util.MarshalToBytes(POLines))
	case "GET_POS_BY_BUYER":
		pos, _ := po.GetPosByBuyerID(stub, args[0])
		return shim.Success(util.MarshalToBytes(pos))
		/*	case "GET_POS_BY_IBMAP":
			pos := invoice.GetPosByIBMAP(stub)*/
		return shim.Success(util.MarshalToBytes(pos))
	case "GET_PO_LINEITEMS_BY_BUYER":
		pos, _ := po.GetPoLineitemsByBuyerID(stub, args[0])
		return shim.Success(util.MarshalToBytes(pos))
		/*	case "GET_POS_BY_IBMAP":
			pos := invoice.GetPosByIBMAP(stub)*/
		return shim.Success(util.MarshalToBytes(pos))
	case "GET_ALL_POS":
		pos := po.GetAllPOs(stub)
		return shim.Success(util.MarshalToBytes(pos))
	case "GET_ALL_POS_LINEITEMS":
		POLines := po.GetAllPOLineItemsByPO(stub, []string{})
		return shim.Success(util.MarshalToBytes(POLines))
	case "GET_VENDOR_PO":
		vendorsPO := po.GetPOByVendor(stub, args[0])
		return shim.Success(util.MarshalToBytes(vendorsPO))

		/* ======== Vendor Master Functions ===========*/
	case "ADD_VENDORS":
		// Called as part of BCI Cron job Bulk Load
		return vmd.AddVendorRecords(stub, args[0])
	case "ADD_VENDOR_BANK":
		// Called as part of BCI Cron job Full Load
		return vmd.AddVendorBankRecords(stub, args[0])
	case "ADD_VENDOR_COMPANY_CODE":
		// Called as part of BCI Cron job Full Load
		return vmd.AddVendorCompanyCodeRecords(stub, args[0])

	case "GET_VENDOR":
		return vmd.GetVendorRec(stub, args[0], args[1], args[2])
	case "GET_ALL_VENDORS":
		return vmd.GetAllVendors(stub)
	case "GET_ALL_VENDORS_BANK":
		return vmd.GetAllVendorsBank(stub)
	case "GET_ALL_VENDORS_COMPANYCODE":
		return vmd.GetAllVendorsCompanyCode(stub)

		/* ======== Company Code Functions ===========*/
	case "ADD_COMPANYCODES":
		return companyCode.AddCompanyRecords(stub, args[0])
	case "GET_COMPANY_CODE":
		companyCode, _ := companyCode.GetCompanyCode(stub, args[0], args[1])
		return shim.Success(util.MarshalToBytes(companyCode))
	case "ADD_COMPANYCODES_DAYS":
		return companyCode.AddCompanyCodeDays(stub, args[0])

	case "GET_COMPANYCODES_DAYS":
		return companyCode.GetAllCompanyCodeDays(stub)

	case "ADD_COMPANYCODES_CONTROLLER":
		return companyCode.AddControllers(stub, args[0])

	case "GET_COMPANYCODES_CONTROLLER":
		return companyCode.GetAllControllers(stub)

		/* ======== BCI Rejected Invoices Functions ===========*/
		/*	case "STORE_BCI_REJECTED_INVOICES":
			return invoice.AddRejectedInvoiceRecords(stub, args[0])*/

		/* ======== Invoice Functions ===========*/
	case "ADD_INVOICES":
		return invoice.AddInvoiceRecords(stub, args[0], true)

	case "UPDATE_INVOICES":
		return invoice.AddInvoiceRecords(stub, args[0], false)

	case "GET_INVOICE":
		return invoice.GetInvoiceRecord(stub, args[0], args[1])

	case "GET_ALL_INVOICES":
		return invoice.GetAllInvoices(stub)

	case "GET_DUE_DATE_DETAILS":
		return invoice.GetDueDateDetail(stub, args[0])

	case "GET_TURN_AROUND_DETAILS":
		return invoice.GetTurnAroundDetail(stub, args[0])

	case "GET_INVOICES_FOR_IBMAP":
		return invoice.GetInvoicesForIbmap(stub)

	case "GET_INVOICE_BY_STATUS":
		invoiceStauses := invoice.GetInvoiceByStatus(stub, args[0])
		return shim.Success(util.MarshalToBytes(invoiceStauses))

	case "GET_INVOICE_BY_SUPPLIER":
		return invoice.GetInvoicesByVendorID(stub, args[0])

	case "GET_INVOICE_BY_BUYER":
		return invoice.GetInvoicesByBuyerID(stub, args[0])

	case "GET_INVOICE_STATUS":
		invStat, _ := invoice.GetInvoiceStatus(stub, []string{args[0], args[1]})
		return shim.Success(util.MarshalToBytes(invStat))

	case "SUBMIT_INVOICE":
		return invoice.SubmitInvoice(stub, args[0])

	case "RESUBMIT_AFTER_DB_REFRESH":
		return invoice.ResubmitAfterDbRefresh(stub, args[0])

		//	case "GET_EVENT_LIST":
		//		return invoice.GetAllEvents(stub)

		//	case "PARTIAL_GRN":
		//		myLogger.Debugf("PArtial GRN============================")
		//		return invoice.PartialGrn(stub, args[0])

	case "GET_LINE_HISTORY":
		return invoice.GetInvoiceLineHistory(stub, args[0], args[1])

	case "ADD_SAP_PROCESSED_INVOICES":
		return invoice.AddSAPProcessedInvoiceRecords(stub, args[0])

	case "GET_SAP_PROCESSED_INVOICE":
		return invoice.GetSAPProcessedInvoiceRecord(stub, args[0], args[1])

	case "GET_ALL_SAP_PROCESSED_INVOICES":
		return invoice.GetAllSAPProcessedInvoices(stub)

		/*=============NPNP functions======*/

	case "ADD_NPNP":
		return refTable.AddNpnpRecords(stub, args[0])

	case "GET_NPNP":
		npnpval, _ := refTable.GetNpnp(stub, args[0], args[1], args[2], args[3])
		return shim.Success(util.MarshalToBytes(npnpval))

	case "GET_ALL_NPNP":
		npnps := refTable.GetAllNpnp(stub)
		return shim.Success(util.MarshalToBytes(npnps))

		/*==================Tax Code Functions==========================*/

	case "ADD_TAXCODE":
		return refTable.AddTaxRecords(stub, args[0])

	case "GET_TAX_CODE":
		taxcodeval, _ := refTable.GetTaxCode(stub, args[0], args[1], args[2], args[3], args[4])
		return shim.Success(util.MarshalToBytes(taxcodeval))

	case "GET_ALL_TAXCODE":
		taxcodes := refTable.GetAllTaxCode(stub)
		return shim.Success(util.MarshalToBytes(taxcodes))

	/*==================ExchangeRate Functions==========================*/

	case "ADD_EXCHANGE_RATE":
		return refTable.AddExchangeRecords(stub, args[0])

	case "GET_EXCHANGE_RATE":
		exchangerateval, _ := refTable.GetExchangeRate(stub, args[0], args[1], args[2], args[3], args[4], args[5])
		return shim.Success(util.MarshalToBytes(exchangerateval))

	case "GET_ALL_EXCHANGE_RATE":
		exchangerate := refTable.GetAllExchangeRate(stub)
		return shim.Success(util.MarshalToBytes(exchangerate))

	/*-------------------DynamicTables-----------------------------*/

	case "ADD_DYNAMICTABLES":
		return refTable.AddDynamicTablesRecords(stub, args[0])

	case "GET_DYNAMIC_TABLES":
		dynamictablesval, _ := refTable.GetDynamicTables(stub, args[0], args[1], args[2], args[3], args[4])
		return shim.Success(util.MarshalToBytes(dynamictablesval))

	case "GET_ALL_DYNAMIC_TABLES":
		dynamictables := refTable.GetAllDynamicTables(stub)
		return shim.Success(util.MarshalToBytes(dynamictables))

	/*-------------------Currency-----------------------------*/

	case "ADD_CURRENCY":
		return refTable.AddCurrencyRecords(stub, args[0])

	case "GET_CURRENCY":
		currencyval, _ := refTable.GetCurrency(stub, args[0], args[1], args[2], args[3], args[4])
		return shim.Success(util.MarshalToBytes(currencyval))

	case "GET_ALL_CURRENCY":
		currency := refTable.GetAllCurrency(stub)
		return shim.Success(util.MarshalToBytes(currency))

	/*-------------------CurrencyDescription-----------------------------*/

	case "ADD_CURRENCYDESCRIPTION":
		return refTable.AddCurrencyDescRecords(stub, args[0])

	case "GET_CURRENCYDESCRIPTION":
		currencyDescriptionval, _ := refTable.GetCurrencyDescription(stub, args[0], args[1], args[2], args[3], args[4])
		return shim.Success(util.MarshalToBytes(currencyDescriptionval))

	case "GET_ALL_CURRENCYDESCRIPTION":
		currencydesc := refTable.GetAllCurrencyDescription(stub)
		return shim.Success(util.MarshalToBytes(currencydesc))

	/*-------------------SAPCountry-----------------------------*/

	case "ADD_SAPCOUNTRY":
		return refTable.AddSAPCountryRecords(stub, args[0])

	case "GET_SAPCOUNTRY":
		sapcountryval, _ := refTable.GetSAPCountry(stub, args[0], args[1], args[2])
		return shim.Success(util.MarshalToBytes(sapcountryval))

	case "GET_ALL_SAPCOUNTRY":
		sapcountry := refTable.GetAllSAPCountry(stub)
		return shim.Success(util.MarshalToBytes(sapcountry))

	/*-------------------PaymentTermsDescription-----------------------------*/

	case "ADD_PAYMENTTERMSDESCRIPTION":
		return refTable.AddPaymentTermsDescRecords(stub, args[0])

	case "GET_PAYMENTTERMSDESCRIPTION":
		paymentTermsval, _ := refTable.GetPaymentTermsDescriptions(stub, args[0], args[1], args[2], args[3])
		return shim.Success(util.MarshalToBytes(paymentTermsval))

	case "GET_ALL_PAYMENTTERMSDESCRIPTION":
		paymenttermsdesc := refTable.GetAllPaymentTermsDescription(stub)
		return shim.Success(util.MarshalToBytes(paymenttermsdesc))

	/*-------------------DCGRN-----------------------------*/

	case "ADD_DCGRN":
		return refTable.AddDCGRNRecords(stub, args[0])

	case "GET_DCGRN":
		dcgrnval, _ := refTable.GetDCGRN(stub, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8], args[9], args[10])
		return shim.Success(util.MarshalToBytes(dcgrnval))

	case "GET_ALL_DCGRN":
		dcgrns := refTable.GetAllDCGRN(stub)
		return shim.Success(util.MarshalToBytes(dcgrns))

	/*-------------------PaymentTerms-----------------------------*/

	case "ADD_PAYMENTTERMS":
		return refTable.AddPaymentTermsRecords(stub, args[0])

	case "GET_PAYMENTTERMS":
		paymentTermsval, _ := refTable.GetPaymentTerms(stub, args[0], args[1], args[2], args[3])
		return shim.Success(util.MarshalToBytes(paymentTermsval))

	case "GET_ALL_PAYMENTTERMS":
		paymentTermsval := refTable.GetAllPaymentTerms(stub)
		return shim.Success(util.MarshalToBytes(paymentTermsval))

		/* ======== GRN Functions ===========*/
	case "GET_GRN":
		grns, _ := grn.GetGRN(stub, []string{args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8]})
		return shim.Success(util.MarshalToBytes(grns))

	case "ADD_GRNS":
		// Called as part of BCI Cron job Bulk Load
		return grn.AddGrnRecords(stub, args[0])

	case "GET_ALL_GRNS":
		return grn.GetAllGRNs(stub)

	case "GET_GRNS_BY_PO":
		grns := grn.GetGrnsByPO(stub, []string{args[0], args[1]})
		return shim.Success(util.MarshalToBytes(grns))

	case "GET_GRNS_BY_PO_AND_POLINE":
		grns := grn.GetGrnsByPOAndPOLine(stub, []string{args[0], args[1], args[2]})
		return shim.Success(util.MarshalToBytes(grns))
		/* ======== Audit Functions ===========*/
	case "ADD_AUDIT":
		audit.AddAuditLog(stub, []string{args[0], args[1]}, args[2])
		return shim.Success([]byte("Success"))

	case "GET_AUDIT":
		return audit.GetAudit(stub, []string{args[0], args[1]})

	case "GET_HISTORY":
		return audit.GetAuditLog(stub, args[0], args[1])

	case "GET_INVOICE_DETAILS_HISTORY":
		return audit.GetInvoiceDetailsLog(stub, args[0], args[1])

	case "GET_REPORT":
		return report.GetReport(stub, args[0], args[1], args[2])

	case "GET_CHAINCODE_VERSION":
		return shim.Success(util.MarshalToBytes("Sprint2-V-New-20-Jun-2016"))

	/*=============OTHERS==============*/

	case "CLEAR_WORLD_STATE":
		return util.ClearWorldState(stub, args)

	/*case "GET_POS_BY_BUYER_VENDORNAME":
		pos, _ := invoice.GetPOByBuyerVendorName(stub, args[0], args[1])
		return shim.Success(util.MarshalToBytes(pos))

	case "GET_PO_BY_IBMAP_VENDOR_NAME":
		vendorsPO := invoice.GetPosByIBMAPVendorName(stub, args[0])
		return shim.Success(util.MarshalToBytes(vendorsPO))*/

	/*	case "ADD_REF_TABLES":
		return reftable.AddLookup(stub, args[0], args[1])*/

	/*	case "GET_PAYMENT_TERMS":
		paymentTerms := reftable.FindPaymentTerms(stub, args[0], args[1])
		return shim.Success(util.MarshalToBytes(paymentTerms))*/
	/*
		case "GET_DUE_DATE":
			var invScanDate util.BCDate
			invScanDate.UnmarshalJSON([]byte(args[0]))

			var invDocDate util.BCDate
			invDocDate.UnmarshalJSON([]byte(args[1]))

			dueDate := reftable.CalculateDueDate(stub, invScanDate, invDocDate, args[2], args[3], args[4])
			return shim.Success(util.MarshalToBytes(dueDate))*/

	/*	case "DYNAMIC_GRN":
			return invoice.DynamicGRN(stub, args[0], args[1], args[2])

		case "DYNAMIC_GRN_UP_DATE":
			return invoice.DynamicGRNUpdate(stub, args[0], args[1], args[2])*/

	/*
		case "GET_ALL_EMAIL_INVOICES":
			return invoice.GetAllInvoiceEmailRecords(stub)
		case "GET_INVOICE_EMAIL":
			    return invoice.GetInvoiceEmailRecord(stub, args[0], args[1])

		case "GET_BUYER_REMINDER_EMAIL_INVOICES":
			return invoice.GetBuyerReminderEmailRecords(stub)*/

	/*	case "SINGLE_VS_MULTIPLE_PO":
					dayCount,_,_ := invoice.ChecksingleVsMultiplePo(stub,args[0])
						return shim.Success(util.MarshalToBytes(dayCount))

		case "COMPANY_CODE_SELECTION":
					dayCount,_,_ := invoice.SelectCompanyCode(stub,args[0])
					return shim.Success(util.MarshalToBytes(dayCount))*/
	default:
		return shim.Error("Received unknown function invocation")
	}

	return shim.Error("Received unknown function invocation")
}

// @Deprecated
// func (t *ServicesChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) (pb.Response) {
// 		return nil, nil
// }

func main() {
	err := shim.Start(new(ServicesChaincode))
	if err != nil {
		fmt.Printf("Error starting ServicesChaincode: %s", err)
	}
}
