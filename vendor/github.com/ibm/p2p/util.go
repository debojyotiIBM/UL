/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package p2p

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ibm/db"
	"github.com/ibm/pme"
	logging "github.com/op/go-logging"
)

var myLogger = logging.MustGetLogger("Procure-To-Pay util")

const INV_STATUS_INIT = "NEW"
const INV_STATUS_PROCESSED = "PROCESSED"
const INV_STATUS_PROCESSING = "PROCESSING IN PROGRESS"
const INV_STATUS_REJECTED = "REJECTED"
const INV_STATUS_PENDING_AP = "AWAITING IBM AP ACTION"
const INV_STATUS_PENDING_BUYER = "AWAITING BUYER ACTION"
const INV_STATUS_PENDING_VMD = "AWAITING VMD ACTION"
const INV_STATUS_WAITING_PO_REFRESH = "WAITING DB REFRESH FOR PO"
const INV_STATUS_WAITING_VENDOR_REFRESH = "WAITING DB REFRESH FOR VENDOR"
const INV_STATUS_WAITING_EMAIL_REFRESH = "WAITING DB REFRESH FOR EMAIL"
const INV_STATUS_WAITING_INVOICE_FIX = "WAITING FOR INVOICE FIX"
const INV_STATUS_WAITING_BUYER_NOTIFICATION = "WAITING FOR BUYER TO BE NOTIFIED"
const INV_STATUS_WAITING_FOR_GRN = "WAITING FOR GRN"
const INV_STATUS_EASY_ROBO_GRN = "WAITING FOR EASY ROBO"
const INV_STATUS_TRIGGER_EMAIL = "TRIGGER EMAIL TO BUYER OR PLANNER"

type BCDate struct {
	t time.Time
}

func (ct *BCDate) Time() time.Time {
	return ct.t
}

func (ct *BCDate) String() string {
	return ct.t.Format(bcdLayout)
}

func (ct *BCDate) SetTime(input_time time.Time) {
	ct.t = input_time
	return
}

const bcdLayout = "20060102"

func (ct *BCDate) UnmarshalJSON(b []byte) (err error) {

	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.t = time.Time{}
		//fmt.Println("In Unmarshal JSON ", ct.t.String())
		return
	}
	ct.t, err = time.Parse(bcdLayout, s)
	if err != nil {
		fmt.Println("Parsing error, expects string in format : ", bcdLayout)
	}
	err = nil
	//fmt.Println("In Unmarshal JSON ", ct.t.String())
	return
}

func (ct BCDate) MarshalJSON() ([]byte, error) {
	//fmt.Println("In Marshal JSON ", ct.t.String())
	if ct.t.UnixNano() == nilTime {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", ct.t.Format(bcdLayout))), nil
}

func CreateDateObject(date string) BCDate {
	var ct BCDate
	s := strings.Trim(date, "\"")
	if s == "null" {
		ct.t = time.Time{}
		return ct
	}
	ct.t, _ = time.Parse(bcdLayout, s)
	return ct
}

var nilTime = (time.Time{}).UnixNano()

func (ct *BCDate) IsSet() bool {
	return ct.t.UnixNano() != nilTime
}

type BCTime struct {
	t time.Time
}

func (ct *BCTime) Time() time.Time {
	return ct.t
}

func (ct *BCTime) String() string {
	return ct.t.Format(bctLayout)
}

func (ct *BCTime) SetTime(input_time time.Time) {
	ct.t = input_time
	return
}

const bctLayout = "150405"

func (ct *BCTime) UnmarshalJSON(b []byte) (err error) {

	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.t = time.Time{}
		//fmt.Println("In Unmarshal JSON ", ct.t.String())
		return
	}
	ct.t, err = time.Parse(bctLayout, s)
	if err != nil {
		fmt.Println("Parsing error, expects string in format : ", bctLayout)
	}
	err = nil
	//fmt.Println("In Unmarshal JSON ", ct.t.String())
	return
}

func (ct BCTime) MarshalJSON() ([]byte, error) {
	//fmt.Println("In Marshal JSON ", ct.t.String())
	if ct.t.UnixNano() == nilTime {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", ct.t.Format(bctLayout))), nil
}

func CreateTimeObject(date string) BCTime {
	var ct BCTime
	s := strings.Trim(date, "\"")
	if s == "null" {
		ct.t = time.Time{}
		return ct
	}
	ct.t, _ = time.Parse(bctLayout, s)
	return ct
}

func (ct *BCTime) IsSet() bool {
	return ct.t.UnixNano() != nilTime
}

type Args struct {
	Time BCDate
}

var data = `
    {"Time": "20140801"}
`

func main() {
	a := Args{}
	b := Args{}
	json.Unmarshal([]byte(data), &a)
	//fmt.Println(a.Time.Format(bcdLayout))
	data1, _ := json.Marshal(&a)
	fmt.Println(string(data1))
	json.Unmarshal(data1, &b)
	fmt.Println(b.Time.t.String())
}

func ProbableMatch(value1 string, value2 string) bool {
	myLogger.Debugf("value1 - ", value1)
	myLogger.Debugf("value2 - ", value2)
	if pme.Score(value1, value2) == 100 {
		return true
	}
	return false
}

func MarshalToBytes(value interface{}) []byte {
	bytes, marshallErr := json.Marshal(value)
	if marshallErr != nil {
		myLogger.Debugf("Error in marshalling : ", marshallErr, value)
		return bytes
	}
	return bytes
}

func UpdateReferenceData(stub shim.ChaincodeStubInterface, tableName string, keys []string, dataLink string) {
	record, _ := db.TableStruct{Stub: stub, TableName: tableName, PrimaryKeys: keys, Data: ""}.Get()
	myLogger.Debugf("Entered UpdateReferenceData funcc=================")
	myLogger.Debugf("Table Name from invoice================", tableName)
	myLogger.Debugf("TAB_INVOICE_BY_STATUS====================>", TAB_INVOICE_BY_STATUS)
	if record != "" {
		myLogger.Debugf("Record : ", record)
		myLogger.Debugf("dataLink : ", dataLink)

		dataFields := strings.Split(dataLink, "|")
		for _, dataField := range dataFields {
			if !strings.Contains(record, dataField) {
				record = record + "|" + dataField
			}
		}

		myLogger.Debugf("Before1 TAB_INVOICE_BY_STATUS If Condition===============")
		if tableName == TAB_INVOICE_BY_STATUS {
			// Remove previous status references for this ent ID
			myLogger.Debugf("Remove previous status references for this ent ID")
			myLogger.Debugf("data fileds=============", dataFields)
			UpdateStatusReference(stub, dataFields)
		}
	} else {
		record = dataLink
		dataFields := strings.Split(dataLink, "|")
		if tableName == TAB_INVOICE_BY_STATUS {
			// Remove previous status references for this ent ID
			myLogger.Debugf("Remove previous status references for this ent ID")
			myLogger.Debugf("data fileds=============", dataFields)
			UpdateStatusReference(stub, dataFields)
		}
	}
	if record != "" {
		db.TableStruct{Stub: stub, TableName: tableName, PrimaryKeys: keys, Data: record}.Add()
	}

}

func UpdateStatusReference(stub shim.ChaincodeStubInterface, dataFields []string) {

	myLogger.Debugf("UpdateStatusReference=======>", dataFields)
	invoiceStatusxRecord, _ := db.TableStruct{Stub: stub, TableName: TAB_INVOICE_BY_STATUS, PrimaryKeys: []string{}, Data: ""}.GetAll()

	for _, dataField := range dataFields {
		for invoiceStatus, record := range invoiceStatusxRecord {
			if strings.Contains(record, dataField) {
				record = removeReference(record, dataField)
				_, keys, _ := stub.SplitCompositeKey(invoiceStatus)
				db.TableStruct{Stub: stub, TableName: TAB_INVOICE_BY_STATUS, PrimaryKeys: keys, Data: record}.Add()
				break
			}
		}
	}
}

func removeReference(valueStr string, valueToRemove string) string {
	newValuesStr := ""
	dataFields := strings.Split(valueStr, "|")
	for _, dataField := range dataFields {
		if dataField != valueToRemove {
			if newValuesStr != "" {
				newValuesStr = newValuesStr + "|"
			}
			newValuesStr = newValuesStr + dataField
		}
	}
	return newValuesStr
}

func GetFloatFromString(val string) float64 {
	floatVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		myLogger.Debugf("Error converting to float64 : ", val)
		return 0.0
	}
	return floatVal
}

func GetStringFromFloat(val float64) string {
	valStr := strconv.FormatFloat(val, 'f', 2, 64)
	return valStr
}

func StringArrayContains(array []string, element string) bool {
	for _, ele := range array {
		if ele == element {
			return true
		}
	}
	return false
}

func RemoveRecord(stub shim.ChaincodeStubInterface, tableName string, keys []string, dataLink string) {
	recs3 := db.TableStruct{Stub: stub, TableName: TAB_INVOICE_BY_BUYER, PrimaryKeys: keys, Data: ""}.Delete()
	myLogger.Debugf("After removal", recs3)

}

func GetUUID() string {
	unix32bits := uint32(time.Now().UTC().Unix())
	buff := make([]byte, 12)
	numRead, err := rand.Read(buff)
	if numRead != len(buff) || err != nil {
		panic(err)
		myLogger.Debugf("Error in generating UUID", err)
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x-%x\n", unix32bits, buff[0:2], buff[2:4], buff[4:6], buff[6:8], buff[8:])
	return uuid
}

func GetIntFromString(val string) int64 {
	valStr, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		myLogger.Debugf("Error converting to int : ", val)
		return 0
	}
	return valStr
}
func GetStringFromInt(val int64) string {
	valStr := strconv.FormatInt(val, 10)
	return valStr
}

func EqualsIgnoreCase(one string, two string) bool {
	if strings.ToLower(strings.TrimSpace(one)) == strings.ToLower(strings.TrimSpace(two)) {
		return true
	}
	return false
}

func StrippedLowerCase(one string) string {
	return strings.ToLower(strings.TrimSpace(one))
}

func ClearWorldState(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) == 0 {
		return shim.Success(MarshalToBytes("No tables given in invoke"))
	}
	myLogger.Debugf("Started clearing world state")
	var tabList []string
	if EqualsIgnoreCase(args[0], "all") {
		tabList = ALL_TABLES_LIST
		for i, tab := range args {
			if i == 0 {
				continue
			}
			tabList = append(tabList, tab)
		}
	} else {
		tabList = args
	}

	for _, tableObjectType := range tabList {

		keysIter, _ := stub.GetStateByPartialCompositeKey(tableObjectType, []string{})
		for keysIter.HasNext() {
			resp, iterErr := keysIter.Next()
			if iterErr != nil {
				myLogger.Debugf("Keys iteration failed. Error accessing state : %v\n" + string(resp.Key))
				return shim.Error("Keys iteration failed. Error accessing state : " + string(resp.Key))
			}
			err := stub.DelState(string(resp.Key))
			if err != nil {
				myLogger.Debugf("Error in DelState for key : %v\n" + string(resp.Key))
				return shim.Error("Error in DelState for key : " + string(resp.Key))
			}

		}
		keysIter.Close()
		myLogger.Debugf("Deleted table : \"%v\"", tableObjectType)
	}
	myLogger.Debugf("Clearing world state completed")
	return shim.Success(MarshalToBytes("Successfully cleared world state"))
}
