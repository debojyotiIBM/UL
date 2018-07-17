package invoice

import (
	"encoding/json"
	"testing"

	"github.com/ibm/p2p/po"

	"github.com/ibm/p2p/vmd"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type VNATestChainCode struct {
}

func (sc *VNATestChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Success(nil)
}

//Invoke is the entry point for any transaction
func (sc *VNATestChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	var resp pb.Response
	method, args := stub.GetFunctionAndParameters()
	switch method {
	case "saveVMD":
		resp = vmd.AddVendorRecords(stub, args[0])
	case "getVMD":
		resp = vmd.GetAllVendors(stub)
	case "savePO":
		resp = po.AddPoRecords(stub, args[0])
	case "getPO":
		poList := po.GetAllPOs(stub)
		returnPayload, _ := json.MarshalIndent(poList, "", " ")
		resp = shim.Success(returnPayload)
	case "saveRoute":
		resp = SaveVNADynamicUIRouteConfig(stub)
	case "getRoute":
		resp = GetVNADynamicUIRouteConfig(stub)
	case "saveAutoRej":
		resp = SaveVNADynamicAutoRejConfig(stub)
	case "getAutoRej":
		resp = GetVNADynamicAutoRejConfig(stub)
	case "saveBASL":
		resp = SaveVNADynamicBusAsListConfig(stub)
	case "getBASL":
		resp = GetVNADynamicBusAsListConfig(stub)
	case "vendNameAuth":
		var inputInvoice Invoice
		err := json.Unmarshal([]byte(args[0]), &inputInvoice)
		if err != nil {
			resp = shim.Error("Invalid invoice json")
			break
		}
		trxnContext := Context{}
		trxnContext.Invoice = inputInvoice
		statCode, errMsg, invStatus := VendorNameAuthentication(stub, &trxnContext)
		returnMessage := make(map[string]interface{})
		returnMessage["statCode"] = statCode
		returnMessage["errMsg"] = errMsg
		returnMessage["invStatus"] = invStatus
		returnMessage["modifiedInvoice"] = trxnContext.Invoice
		returnPayload, _ := json.MarshalIndent(returnMessage, "", " ")
		resp = shim.Success(returnPayload)
	}
	return resp
}
func initVNASChaincode(t *testing.T) *shim.MockStub {
	scc := new(VNATestChainCode)
	stub := shim.NewMockStub("xzs829", scc)
	initRes := stub.MockInit("920202-342222-112134", [][]byte{[]byte("init")})
	if initRes.Status != shim.OK {
		t.Logf("Initalization failed")
		t.FailNow()
	}
	return stub
}
