/*
    Copyright IBM Corp. 2017 All Rights Reserved.
    Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
    @author : Pushpalatha M Hiremath
*/

package db

/*
    Generic database interface
*/

import(
  	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type TableStruct struct{
  Stub         shim.ChaincodeStubInterface
	TableName    string
  PrimaryKeys   []string
  Data         string
}

type Table interface {
	Add()(error)
	Update()(error)
  Get()(string, error)
  GetAll()(map[string]string, error)
  Delete()(error)
  GetHistory()(error)
}
