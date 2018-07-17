package refTable

import (
	//"strconv"
	"encoding/json"
	"regexp"
	"sort"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	"github.com/ibm/p2p/vmd"
	//"github.com/op/go-logging"
	"strconv"
)

type LookupType struct {
	JsonObjectsArray_ string `json:"jsonObjects"`
}

func (lookuptype *LookupType) JsonObjectsArray() string {
	return lookuptype.JsonObjectsArray_
}

func (lookuptype *LookupType) SetJsonObjectsArray(JsonObjectsArray string) {
	lookuptype.JsonObjectsArray_ = JsonObjectsArray
}

type PaymentTerms struct {
	Client                         string
	Paymenttermcode                string
	Daylimit                       int
	Datetype                       string
	Calendarday                    int
	Additionalmonths               int
	Daysfrombaselinedatetopayment1 int
	Cashdiscpercentrate1           float64
	Daysfrombaselinedatetopayment2 int
	Cashdiscpercentrate2           float64
	Daysfrombaselinedatetopayment3 int
	Duedateforspecialcondition     int
	Monthsforspecialcondterm1      int
	Duedatespecialcond1            int
	Monthsforspecialcondterm2      int
	Duedatespecialcond2            int
	Monthsforspecialcondterm3      int
	Paymentblock                   string
	Numberstdtext                  string
	Accounttype                    string
	Loaddatetime                   util.BCDate
	Filename                       string
	Linenum                        string
	Modifydate                     util.BCDate
	Deletedate                     util.BCDate
	ERPSystem                      string
}

var paymentTerms PaymentTerms

// func (paymentterms *PaymentTerms) Client() string {
// 	return paymentterms.Client_
// }

// func (paymentterms *PaymentTerms) SetClient(Client string) {
// 	paymentterms.Client_ = Client
// }

// func (paymentterms *PaymentTerms) Paymenttermcode() string {
// 	return paymentterms.Paymenttermcode_
// }

// func (paymentterms *PaymentTerms) SetPaymenttermcode(Paymenttermcode string) {
// 	paymentterms.Paymenttermcode_ = Paymenttermcode
// }

// func (paymentterms *PaymentTerms) Daylimit() int {
// 	return paymentterms.Daylimit_
// }

// func (paymentterms *PaymentTerms) SetDaylimit(Daylimit int) {
// 	paymentterms.Daylimit_ = Daylimit
// }

// func (paymentterms *PaymentTerms) Datetype() string {
// 	return paymentterms.Datetype_
// }

// func (paymentterms *PaymentTerms) SetDatetype(Datetype string) {
// 	paymentterms.Datetype_ = Datetype
// }

// func (paymentterms *PaymentTerms) Calendarday() int {
// 	return paymentterms.Calendarday_
// }

// func (paymentterms *PaymentTerms) SetCalendarday(Calendarday int) {
// 	paymentterms.Calendarday_ = Calendarday
// }

// func (paymentterms *PaymentTerms) Additionalmonths() int {
// 	return paymentterms.Additionalmonths_
// }

// func (paymentterms *PaymentTerms) SetAdditionalmonths(Additionalmonths int) {
// 	paymentterms.Additionalmonths_ = Additionalmonths
// }

// func (paymentterms *PaymentTerms) Daysfrombaselinedatetopayment1() int {
// 	return paymentterms.Daysfrombaselinedatetopayment1_
// }

// func (paymentterms *PaymentTerms) SetDaysfrombaselinedatetopayment1(Daysfrombaselinedatetopayment1 int) {
// 	paymentterms.Daysfrombaselinedatetopayment1_ = Daysfrombaselinedatetopayment1
// }

// func (paymentterms *PaymentTerms) Cashdiscpercentrate1() float64 {
// 	return paymentterms.Cashdiscpercentrate1_
// }

// func (paymentterms *PaymentTerms) SetCashdiscpercentrate1(Cashdiscpercentrate1 float64) {
// 	paymentterms.Cashdiscpercentrate1_ = Cashdiscpercentrate1
// }

// func (paymentterms *PaymentTerms) Daysfrombaselinedatetopayment2() int {
// 	return paymentterms.Daysfrombaselinedatetopayment2_
// }

// func (paymentterms *PaymentTerms) SetDaysfrombaselinedatetopayment2(Daysfrombaselinedatetopayment2 int) {
// 	paymentterms.Daysfrombaselinedatetopayment2_ = Daysfrombaselinedatetopayment2
// }

// func (paymentterms *PaymentTerms) Cashdiscpercentrate2() float64 {
// 	return paymentterms.Cashdiscpercentrate2_
// }

// func (paymentterms *PaymentTerms) SetCashdiscpercentrate2(Cashdiscpercentrate2 float64) {
// 	paymentterms.Cashdiscpercentrate2_ = Cashdiscpercentrate2
// }

// func (paymentterms *PaymentTerms) Daysfrombaselinedatetopayment3() int {
// 	return paymentterms.Daysfrombaselinedatetopayment3_
// }

// func (paymentterms *PaymentTerms) SetDaysfrombaselinedatetopayment3(Daysfrombaselinedatetopayment3 int) {
// 	paymentterms.Daysfrombaselinedatetopayment3_ = Daysfrombaselinedatetopayment3
// }

// func (paymentterms *PaymentTerms) Duedateforspecialcondition() int {
// 	return paymentterms.Duedateforspecialcondition_
// }

// func (paymentterms *PaymentTerms) SetDuedateforspecialcondition(Duedateforspecialcondition int) {
// 	paymentterms.Duedateforspecialcondition_ = Duedateforspecialcondition
// }

// func (paymentterms *PaymentTerms) Monthsforspecialcondterm1() int {
// 	return paymentterms.Monthsforspecialcondterm1_
// }

// func (paymentterms *PaymentTerms) SetMonthsforspecialcondterm1(Monthsforspecialcondterm1 int) {
// 	paymentterms.Monthsforspecialcondterm1_ = Monthsforspecialcondterm1
// }

// func (paymentterms *PaymentTerms) Duedatespecialcond1() int {
// 	return paymentterms.Duedatespecialcond1_
// }

// func (paymentterms *PaymentTerms) SetDuedatespecialcond1(Duedatespecialcond1 int) {
// 	paymentterms.Duedatespecialcond1_ = Duedatespecialcond1
// }

// func (paymentterms *PaymentTerms) Monthsforspecialcondterm2() int {
// 	return paymentterms.Monthsforspecialcondterm2_
// }

// func (paymentterms *PaymentTerms) SetMonthsforspecialcondterm2(Monthsforspecialcondterm2 int) {
// 	paymentterms.Monthsforspecialcondterm2_ = Monthsforspecialcondterm2
// }

// func (paymentterms *PaymentTerms) Duedatespecialcond2() int {
// 	return paymentterms.Duedatespecialcond2_
// }

// func (paymentterms *PaymentTerms) SetDuedatespecialcond2(Duedatespecialcond2 int) {
// 	paymentterms.Duedatespecialcond2_ = Duedatespecialcond2
// }

// func (paymentterms *PaymentTerms) Monthsforspecialcondterm3() int {
// 	return paymentterms.Monthsforspecialcondterm3_
// }

// func (paymentterms *PaymentTerms) SetMonthsforspecialcondterm3(Monthsforspecialcondterm3 int) {
// 	paymentterms.Monthsforspecialcondterm3_ = Monthsforspecialcondterm3
// }

// func (paymentterms *PaymentTerms) Paymentblock() string {
// 	return paymentterms.Paymentblock_
// }

// func (paymentterms *PaymentTerms) SetPaymentblock(Paymentblock string) {
// 	paymentterms.Paymentblock_ = Paymentblock
// }

// func (paymentterms *PaymentTerms) Numberstdtext() string {
// 	return paymentterms.Numberstdtext_
// }

// func (paymentterms *PaymentTerms) SetNumberstdtext(Numberstdtext string) {
// 	paymentterms.Numberstdtext_ = Numberstdtext
// }

// func (paymentterms *PaymentTerms) Accounttype() string {
// 	return paymentterms.Accounttype_
// }

// func (paymentterms *PaymentTerms) SetAccounttype(Accounttype string) {
// 	paymentterms.Accounttype_ = Accounttype
// }

// func (paymentterms *PaymentTerms) Loaddatetime() *util.BCDate {
// 	return &paymentterms.Loaddatetime_
// }

// func (paymentterms *PaymentTerms) SetLoaddatetime(Loaddatetime util.BCDate) {
// 	paymentterms.Loaddatetime_ = Loaddatetime
// }

// func (paymentterms *PaymentTerms) Filename() string {
// 	return paymentterms.Filename_
// }

// func (paymentterms *PaymentTerms) SetFilename(Filename string) {
// 	paymentterms.Filename_ = Filename
// }

// func (paymentterms *PaymentTerms) Linenum() string {
// 	return paymentterms.Linenum_
// }

// func (paymentterms *PaymentTerms) SetLinenum(Linenum string) {
// 	paymentterms.Linenum_ = Linenum
// }

// func (paymentterms *PaymentTerms) Modifydate() *util.BCDate {
// 	return &paymentterms.Modifydate_
// }

// func (paymentterms *PaymentTerms) SetModifydate(Modifydate util.BCDate) {
// 	paymentterms.Modifydate_ = Modifydate
// }

// func (paymentterms *PaymentTerms) Deletedate() *util.BCDate {
// 	return &paymentterms.Deletedate_
// }

// func (paymentterms *PaymentTerms) SetDeletedate(Deletedate util.BCDate) {
// 	paymentterms.Deletedate_ = Deletedate
// }

//var myLogger = logging.MustGetLogger("Procure-To-Pay CompanyCode")

func AddPaymentTermsRecords(stub shim.ChaincodeStubInterface, payRecArr string) pb.Response {
	var paymentTerms []PaymentTerms
	err := json.Unmarshal([]byte(payRecArr), &paymentTerms)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input paymentTermsDescriptions array:", err)
	}
	for _, paymentTerms := range paymentTerms {
		db.TableStruct{Stub: stub, TableName: util.TAB_PAYMENTTERMS, PrimaryKeys: []string{paymentTerms.ERPSystem, paymentTerms.Client, paymentTerms.Paymenttermcode, strconv.Itoa(int(paymentTerms.Daylimit))}, Data: string(util.MarshalToBytes(paymentTerms))}.Add()
	}
	return shim.Success(nil)
}

/*
  Get PaymentTerms from blockchain
*/
func GetPaymentTerms(stub shim.ChaincodeStubInterface, erpsystem string, client string, paymenttermcode string, daylimit string) (PaymentTerms, string) {
	ccRecord, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PAYMENTTERMS, PrimaryKeys: []string{erpsystem, client, paymenttermcode, daylimit}, Data: ""}.Get()
	var paymentTerms PaymentTerms
	err := json.Unmarshal([]byte(ccRecord), &paymentTerms)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input paymentTerms:", err, ccRecord)
		return paymentTerms, "ERROR in parsing input paymentTerms"
	}
	return paymentTerms, ""
}

// GetALL method to get all data
func GetAllPaymentTerms(stub shim.ChaincodeStubInterface) []PaymentTerms {
	var allPaymentTerms []PaymentTerms
	PaymentTermsRec, _ := db.TableStruct{Stub: stub, TableName: util.TAB_PAYMENTTERMS, PrimaryKeys: []string{}, Data: ""}.GetAll()
	for _, grnRow := range PaymentTermsRec {
		var currentPaymentTerms PaymentTerms
		json.Unmarshal([]byte(grnRow), &currentPaymentTerms)
		allPaymentTerms = append(allPaymentTerms, currentPaymentTerms)
	}

	return allPaymentTerms
}

func AddLookup(stub shim.ChaincodeStubInterface, lookupKey string, jsonObject string) pb.Response {
	db.TableStruct{Stub: stub, TableName: util.TAB_LOOKUP, PrimaryKeys: []string{lookupKey}, Data: string(util.MarshalToBytes(jsonObject))}.Add()
	myLogger.Debugf("created lookup")
	return shim.Success(nil)
}

func GetLookup(stub shim.ChaincodeStubInterface, lookupKey string) string {
	jsonObjects, _ := db.TableStruct{Stub: stub, TableName: util.TAB_LOOKUP, PrimaryKeys: []string{lookupKey}, Data: ""}.Get()
	myLogger.Debugf("retrieved lookup")
	return jsonObjects
	//return FindPaymentTerms(stub, "PAYMENTTERMCODE","0.267","CLIENT")
}

func FindPaymentTerms(stub shim.ChaincodeStubInterface, col string, colValue string) []PaymentTerms {

	jsonBytes := GetLookup(stub, "PaymentTerms")

	re := regexp.MustCompile("^\"|\"$")
	jsonBytes = re.ReplaceAllString(jsonBytes, "")

	//    	myLogger.Debugf("before unmarshal" + jsonBytes)

	var payTermss []PaymentTerms

	err1 := json.Unmarshal([]byte(jsonBytes), &payTermss)
	if err1 != nil {
		myLogger.Debugf("Error parsing JSON: ", err1)
	}
	return findkey(payTermss, col, colValue)
}

func findkey(arr []PaymentTerms, col string, colValue string) []PaymentTerms {
	myLogger.Debugf("inside find")
	var payTerms []PaymentTerms
	for _, arrobj := range arr {
		if paymentTerms.Paymenttermcode == colValue {
			payTerms = append(payTerms, arrobj)
		}
	}
	return payTerms

}

func CalculateDueDate(stub shim.ChaincodeStubInterface, invScanDate util.BCDate, invDocDate util.BCDate, vendorId string, erpSystem string, client string, paytermCode string) util.BCDate {

	var adoptScanFlag string
	var docDateOrScanTime util.BCDate
	var docDateOrScanDay int
	var baseLineDate time.Time
	var dueDate util.BCDate
	InvVendorDetails, err1 := vmd.GetVendor(stub, erpSystem, vendorId, client)
	if err1 == "" {
		adoptScanFlag = InvVendorDetails.AdoptScanDate
	}
	if adoptScanFlag == "X" {
		//"Scan_Date": "2017-11-06T00:00:00.0000000-00:00"
		docDateOrScanDate := invScanDate
		docDateOrScanTime = docDateOrScanDate
		//docDateOrScanTime, _ = time.Parse(time.RFC3339, docDateOrScanDate)
		//docDateOrScanDay = docDateOrScanTime.Day()
		docDateOrScanDay = docDateOrScanDate.Time().Day()
	} else {
		//"Doc_Date": "30-10-2017",
		docDateOrScanDate := invDocDate
		//docDateOrScanTime, _ = time.Parse("02-01-2006", docDateOrScanDate)
		//docDateOrScanDay = docDateOrScanTime.Day()
		docDateOrScanTime = docDateOrScanDate
		docDateOrScanDay = docDateOrScanDate.Time().Day()
	}

	selectedPT := getFilterPaymentTerms(stub, docDateOrScanDay, paytermCode)
	myLogger.Debugf("Selected Payment Terms: ==================>", selectedPT)

	//Started BaseLine Date
	if adoptScanFlag == "X" {
		baseLineDate = docDateOrScanTime.Time()
	} else {
		if docDateOrScanDay > selectedPT.Calendarday {
			baseLineDate = docDateOrScanTime.Time().AddDate(0, selectedPT.Additionalmonths, 0)
		} else if docDateOrScanDay < selectedPT.Calendarday {
			baseLineDate = docDateOrScanTime.Time().AddDate(0, selectedPT.Additionalmonths, selectedPT.Calendarday-docDateOrScanDay)
		}
	}
	myLogger.Debugf("Calculated BaseLine Date: ==================>", baseLineDate)

	//Started Due Date Calculations
	if selectedPT.Monthsforspecialcondterm1 == 0 && selectedPT.Duedateforspecialcondition == 0 &&
		selectedPT.Monthsforspecialcondterm2 == 0 && selectedPT.Duedatespecialcond1 == 0 &&
		selectedPT.Monthsforspecialcondterm3 == 0 && selectedPT.Duedatespecialcond2 == 0 {

		if selectedPT.Daysfrombaselinedatetopayment3 != 0 {
			myLogger.Debugf("Calculated Total Discounted Date3: ==================>", selectedPT.Daysfrombaselinedatetopayment3)
			dueDate.SetTime(baseLineDate.AddDate(0, 0, selectedPT.Daysfrombaselinedatetopayment3))
		} else if selectedPT.Daysfrombaselinedatetopayment2 != 0 {
			myLogger.Debugf("Calculated Total Discounted Date2: ==================>", selectedPT.Daysfrombaselinedatetopayment2)
			dueDate.SetTime(baseLineDate.AddDate(0, 0, selectedPT.Daysfrombaselinedatetopayment2))
		} else if selectedPT.Daysfrombaselinedatetopayment1 != 0 {
			myLogger.Debugf("Calculated Total Discounted Date1: ==================>", selectedPT.Daysfrombaselinedatetopayment1)
			dueDate.SetTime(baseLineDate.AddDate(0, 0, selectedPT.Daysfrombaselinedatetopayment1))
		} else {
			dueDate.SetTime(baseLineDate)
		}
		myLogger.Debugf("Calculated DueDate condition1: ==================>", dueDate)

	} else {
		var calculatedDays int
		var calculatedMonths int
		if selectedPT.Monthsforspecialcondterm3 != 0 || selectedPT.Monthsforspecialcondterm2 != 0 {
			calculatedDays = selectedPT.Duedatespecialcond2
			calculatedMonths = selectedPT.Monthsforspecialcondterm3
		} else if selectedPT.Monthsforspecialcondterm2 != 0 || selectedPT.Duedatespecialcond1 != 0 {
			calculatedDays = selectedPT.Duedatespecialcond1
			calculatedMonths = selectedPT.Monthsforspecialcondterm2
		} else if selectedPT.Monthsforspecialcondterm1 != 0 || selectedPT.Duedateforspecialcondition != 0 {
			calculatedDays = selectedPT.Duedateforspecialcondition
			calculatedMonths = selectedPT.Monthsforspecialcondterm1
		}
		//TODO
		myLogger.Debugf("Calculated days: ==================>", calculatedDays)
		myLogger.Debugf("Calculated months: ==================>", calculatedMonths)
		if calculatedDays < baseLineDate.Day() && calculatedMonths == 0 {
			//Then we add one months
			dueDate.SetTime(baseLineDate.AddDate(0, 1, 0))
			myLogger.Debugf("Calculated DueDate condition2: ==================>", dueDate)
		} else {
			//*dueDate = new DateTime(dueDate.Year, calculatedMonths, calculatedDays)
			dueDate.SetTime(time.Date(baseLineDate.Year(), time.Month(calculatedMonths)+baseLineDate.Month(), calculatedDays, 0, 0, 0, 0, time.UTC))
			myLogger.Debugf("Calculated DueDate condition3: ==================>", dueDate)
		}
	}
	return dueDate
}

func getFilterPaymentTerms(stub shim.ChaincodeStubInterface, docDateOrScanDay int, ptCode string) PaymentTerms {
	var filteredPaymentTerms []PaymentTerms
	var filteredPT PaymentTerms
	filteredPaymentTerms = FindPaymentTerms(stub, "PaymentTermCode", ptCode)
	myLogger.Debugf("PaymentTerms Found==================>", len(filteredPaymentTerms))
	if len(filteredPaymentTerms) == 1 {
		return filteredPaymentTerms[0]
	} else {
		// var dayLimit int
		var dayLimitList []int
		var PAYTERM_MAP map[int]PaymentTerms
		PAYTERM_MAP = make(map[int]PaymentTerms)
		for _, pt := range filteredPaymentTerms {
			PAYTERM_MAP[pt.Daylimit] = pt
			dayLimitList = append(dayLimitList, pt.Daylimit)
		}
		sort.Ints(dayLimitList)
		myLogger.Debugf("DayLimit Sorted Values==================>", dayLimitList)
		//4, 8 ,13
		for idx, dlimit := range dayLimitList {
			if docDateOrScanDay <= dlimit {
				filteredPT = PAYTERM_MAP[dlimit]
				break
			} else if idx == len(dayLimitList)-1 {
				filteredPT = PAYTERM_MAP[dlimit]
				break
			}
		}
	}
	return filteredPT
}
