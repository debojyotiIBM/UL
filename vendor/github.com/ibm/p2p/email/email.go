/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package email

import (
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/ibm/db"
	util "github.com/ibm/p2p"
	"github.com/op/go-logging"
)

type Email struct {
	Id_    string `json:"id"`
	Email_ string `json:"email"`
	EType_ string `json:"type"`
}

func (email *Email) Id() string {
	return email.Id_
}

func (email *Email) SetId(Id string) {
	email.Id_ = Id
}

func (email *Email) Email() string {
	return email.Email_
}

func (email *Email) SetEmail(Email string) {
	email.Email_ = Email
}

func (email *Email) EType() string {
	return email.EType_
}

func (email *Email) SetEType(EType string) {
	email.EType_ = EType
}

var myLogger = logging.MustGetLogger("Procure-To-Pay : EMAIL")

/*
	Adds Email details to blockchain
*/

func AddEmailRecords(stub shim.ChaincodeStubInterface, emailRecArr string) pb.Response {
	var emails []Email
	err := json.Unmarshal([]byte(emailRecArr), &emails)
	if err != nil {
		myLogger.Debugf("ERROR in parsing input email array:", err, emails)
	}
	for _, email := range emails {
		db.TableStruct{Stub: stub, TableName: util.TAB_EMAIL, PrimaryKeys: []string{email.Id()}, Data: string(util.MarshalToBytes(email))}.Add()
	}
	return shim.Success(nil)
}

/*
 Get Email details from blockchain
*/
func GetEmail(stub shim.ChaincodeStubInterface, id string) (Email, string) {
	var email Email
	emailRec, fetchErr := db.TableStruct{Stub: stub, TableName: util.TAB_EMAIL, PrimaryKeys: []string{id}, Data: ""}.Get()
	if fetchErr != nil {
		myLogger.Debugf("ERROR fetching email :", fetchErr)
		return email, "ERROR fetching email"
	}

	if emailRec == "" {
		return email, "No Email found in the database"
	}

	err := json.Unmarshal([]byte(emailRec), &email)
	if err != nil {
		myLogger.Debugf("ERROR parsing email :", err)
		return email, "ERROR parsing email"
	}
	return email, ""
}

/*
 Get all the Buyers and Planners Email details from blockchain
*/
func GetEmailRecords(stub shim.ChaincodeStubInterface) []Email {
	var buyers []Email
	emailRecMap, _ := db.TableStruct{Stub: stub, TableName: util.TAB_EMAIL, PrimaryKeys: []string{}, Data: ""}.GetAll()

	for _, emailRec := range emailRecMap {
		var email Email
		json.Unmarshal([]byte(emailRec), &email)
		buyers = append(buyers, email)
	}
	return buyers
}
