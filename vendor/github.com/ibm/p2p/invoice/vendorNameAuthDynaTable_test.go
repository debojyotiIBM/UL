package invoice

import (
	"encoding/json"
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func Test_VNASUIRoute(t *testing.T) {
	stub := initVNASChaincode(t)
	dynaConfig := make(map[string]interface{})
	dynaConfig["compCode"] = "2646"
	dynaConfig["invoiceType"] = "INVPO"
	dynaConfig["inputSource"] = ""
	dynaConfig["uiRoute"] = "AWAITING IBM AP ACTION"
	dynaConfig["internalStatus"] = "st-vendNameAuth-1"

	configJSON, _ := json.Marshal(dynaConfig)

	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("saveRoute"), configJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	}
	t.Logf("%s", string(resp.Payload))
	resp = stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("getRoute"), configJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	}
	t.Logf("%s", string(resp.Payload))

}
func Test_VNASAutoRej(t *testing.T) {
	stub := initVNASChaincode(t)
	dynaConfig := make(map[string]interface{})
	dynaConfig["compCode"] = "2646"
	dynaConfig["invoiceType"] = "INVPO"
	dynaConfig["inputSource"] = ""
	dynaConfig["autoRejection"] = true

	configJSON, _ := json.Marshal(dynaConfig)

	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("saveAutoRej"), configJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	}
	t.Logf("%s", string(resp.Payload))
	resp = stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("getAutoRej"), configJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	}
	t.Logf("%s", string(resp.Payload))

}

func Test_VNASBASL(t *testing.T) {

	stub := initVNASChaincode(t)
	dynaConfig := make(map[string]interface{})
	dynaConfig["compCode"] = "2646"
	dynaConfig["invoiceType"] = "INVPO"
	dynaConfig["inputSource"] = ""
	dynaConfig["IBM"] = []string{"WIPRO", "SOFTLAYER"}
	dynaConfig["WIPRO"] = []string{"CATS Ltd", "ZYA Inc"}

	configJSON, _ := json.Marshal(dynaConfig)

	resp := stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("saveBASL"), configJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	}
	t.Logf("%s", string(resp.Payload))
	resp = stub.MockInvoke("378218-9381273-93823", [][]byte{[]byte("getBASL"), configJSON})
	if resp.Status != shim.OK {
		t.Logf("Invoke  failed")
		t.FailNow()
	}
	t.Logf("%s", string(resp.Payload))

}
