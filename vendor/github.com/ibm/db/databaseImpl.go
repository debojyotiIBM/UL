/*
   Copyright IBM Corp. 2017 All Rights Reserved.
   Licensed under the IBM India Pvt Ltd, Version 1.0 (the "License");
   @author : Pushpalatha M Hiremath
*/

package db

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/op/go-logging"
)

/*
   Go RocksDB interface
*/

var myLogger = logging.MustGetLogger("Procure-To-Pay Database Operations")

var STATE_DB_OPERATIONS_LOGGING bool = false

func (tab TableStruct) Add() error {
	tab.Data = strings.Replace(tab.Data, "\\", "", -1)
	dataBytes, err := json.Marshal(tab.Data)
	if err != nil {
		fmt.Println("Table data format error :", err)
	}
	compositeKey, _ := tab.Stub.CreateCompositeKey(tab.TableName, tab.PrimaryKeys)
	if STATE_DB_OPERATIONS_LOGGING {
		myLogger.Debugf("..................................................")
		myLogger.Debugf("Store Data : ")
		myLogger.Debugf("Key - %v", compositeKey)
		myLogger.Debugf("Value - %v", tab.Data)
		myLogger.Debugf("..................................................")
	}
	tab.Stub.PutState(compositeKey, dataBytes)
	return nil
}

func (tab TableStruct) AddBytes(dataBytes []byte) error {

	compositeKey, _ := tab.Stub.CreateCompositeKey(tab.TableName, tab.PrimaryKeys)
	if STATE_DB_OPERATIONS_LOGGING {
		myLogger.Debugf("..................................................")
		myLogger.Debugf("Store Data : ")
		myLogger.Debugf("Key - %v", compositeKey)
		myLogger.Debugf("Value : Bytes")
		myLogger.Debugf("..................................................")
	}
	tab.Stub.PutState(compositeKey, dataBytes)
	return nil
}

func (tab TableStruct) Update() error {
	return nil
}

func (tab TableStruct) Get() (string, error) {

	compositeKey, _ := tab.Stub.CreateCompositeKey(tab.TableName, tab.PrimaryKeys)
	data, _ := tab.Stub.GetState(compositeKey)
	dataVal := string(data)
	dataVal = strings.Replace(dataVal, "\\", "", -1)
	if len(dataVal) > 2 && string(dataVal[0]) == "\"" {
		dataVal = dataVal[1:(len(dataVal) - 1)]
	}
	if STATE_DB_OPERATIONS_LOGGING {
		myLogger.Debugf("..................................................")
		myLogger.Debugf("Get Data : ")
		myLogger.Debugf("Key - %v", compositeKey)
		myLogger.Debugf("Val - %v", dataVal)
		myLogger.Debugf("..................................................")
	}
	return dataVal, nil
}

func (tab TableStruct) GetBytes() ([]byte, error) {

	compositeKey, _ := tab.Stub.CreateCompositeKey(tab.TableName, tab.PrimaryKeys)
	data, _ := tab.Stub.GetState(compositeKey)
	if STATE_DB_OPERATIONS_LOGGING {
		myLogger.Debugf("..................................................")
		myLogger.Debugf("Get Data : ")
		myLogger.Debugf("Key - %v", compositeKey)
		myLogger.Debugf("Val : Bytes")
		myLogger.Debugf("..................................................")
	}
	return data, nil
}

func (tab TableStruct) GetAll() (map[string]string, error) {
	keysIter, _ := tab.Stub.GetStateByPartialCompositeKey(tab.TableName, tab.PrimaryKeys)
	defer keysIter.Close()
	var rowsMap map[string]string
	rowsMap = make(map[string]string)

	for keysIter.HasNext() {
		resp, iterErr := keysIter.Next()
		if iterErr != nil {
			return nil, errors.New(fmt.Sprintf("keys operation failed. Error accessing state: %s", iterErr))
		}

		dataVal := string(resp.Value)
		dataVal = strings.Replace(dataVal, "\\", "", -1)
		if len(dataVal) > 2 && string(dataVal[0]) == "\"" {
			dataVal = dataVal[1:(len(dataVal) - 1)]
		}
		// if STATE_DB_OPERATIONS_LOGGING {
		// 	objType, keys, _ := tab.Stub.SplitCompositeKey(string(resp.Key))
		// 	myLogger.Debugf("GetAll Key : %v :: %v", objType, keys)
		// }
		rowsMap[string(resp.Key)] = dataVal
	}
	if STATE_DB_OPERATIONS_LOGGING {
		myLogger.Debugf("..................................................")
		myLogger.Debugf("Get all Data: ")
		myLogger.Debugf("Key - %v, %v", tab.TableName, tab.PrimaryKeys)
		myLogger.Debugf("Val - %v", rowsMap)
		myLogger.Debugf("..................................................")
	}
	return rowsMap, nil
}

func (tab TableStruct) Delete() error {
	compositeKey, _ := tab.Stub.CreateCompositeKey(tab.TableName, tab.PrimaryKeys)
	tab.Stub.DelState(compositeKey)
	return nil
}

func (tab TableStruct) GetHistory() (string, error) {
	compositeKey, _ := tab.Stub.CreateCompositeKey(tab.TableName, tab.PrimaryKeys)
	keysIter, err := tab.Stub.GetHistoryForKey(compositeKey)
	if err != nil {
		myLogger.Debugf("query operation failed. Error accessing history: ", err)
		return "", errors.New(fmt.Sprintf("query operation failed. Error accessing state: %s", err))
	}
	defer keysIter.Close()

	var buffer bytes.Buffer
	for keysIter.HasNext() {
		resp, iterErr := keysIter.Next()
		if iterErr != nil {
			myLogger.Debugf("query operation failed. Error accessing history: ", err)
			return "", errors.New(fmt.Sprintf("query operation failed. Error accessing state: %s", err))
		}

		dataVal := string(resp.Value)
		dataVal = strings.Replace(dataVal, "\\", "", -1)
		if len(dataVal) > 2 && string(dataVal[0]) == "\"" {
			dataVal = dataVal[1:(len(dataVal) - 1)]
		}
		buffer.WriteString(dataVal)
	}
	if STATE_DB_OPERATIONS_LOGGING {
		myLogger.Debugf("..................................................")
		myLogger.Debugf("Get History - ")
		myLogger.Debugf("Key : %v", compositeKey)
		myLogger.Debugf("Value : %v", buffer.String())
		myLogger.Debugf("..................................................")
	}
	return buffer.String(), nil
}
