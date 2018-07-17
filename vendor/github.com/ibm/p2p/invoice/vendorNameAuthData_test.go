package invoice

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

const _lc_RAND_VALUESTRING = "0123456789"
const _lc_RANDOM_ALPHANUMVALUESTRING = "ABCDEFGHIJKLMNOPQRSTWXYZ0123456789"

type TestDataGenerationSpec struct {
	FieldName       string   `json:"fname"`
	FieldType       string   `json:"type"`
	FieldSize       int      `json:"size"`
	FieldMax        int      `json:"maxValue"`
	FieldMultiplier *int32   `json:"multipler"`
	FieldPattern    []int    `json:"pattern"`
	FieldPrefix     *string  `json:"prefix"`
	FieldSuffix     *string  `json:"suffix"`
	ValidList       []string `json:"pickFromList"`
	randomizer      *rand.Rand
}

func (tds *TestDataGenerationSpec) Init() {
	randSource := rand.NewSource(time.Now().UnixNano())
	tds.randomizer = rand.New(randSource)
}

func (tds TestDataGenerationSpec) GenerateData() interface{} {
	var returnValue interface{}
	switch tds.FieldType {
	case "an-string":
		value := ""
		if tds.FieldSize > 0 {
			value = tds.RandANString(tds.FieldSize)
		}
		if tds.FieldPrefix != nil {
			value = *tds.FieldPrefix + value
		}
		if tds.FieldSuffix != nil {

			value = value + *tds.FieldPrefix
		}
		returnValue = value
	case "n-string":
		value := ""
		if tds.FieldSize > 0 {
			value = tds.RandString(tds.FieldSize)
		}
		if tds.FieldPrefix != nil {
			value = *tds.FieldPrefix + value
		}
		if tds.FieldSuffix != nil {

			value = value + *tds.FieldPrefix
		}
		returnValue = value
	case "pickList":
		returnValue = tds.PickFromList(tds.ValidList)
	case "float32":
		baseValue := tds.randomizer.Float32()
		if tds.FieldMultiplier != nil {
			baseValue = baseValue * float32(*tds.FieldMultiplier)
		}
		returnValue = baseValue
	case "float64":
		baseValue := tds.randomizer.Float64()
		if tds.FieldMultiplier != nil {
			baseValue = baseValue * float64(*tds.FieldMultiplier)
		}
		returnValue = baseValue
	case "int":
		baseValue := tds.randomizer.Intn(tds.FieldMax)
		returnValue = baseValue
	case "uuid":
		returnValue = tds.GenerateRandomString(tds.FieldPattern)
	}
	return returnValue
}
func (tds TestDataGenerationSpec) GenerateRandomString(pattern []int) string {
	randomString := ""
	for _, len := range pattern {
		randomString = randomString + "-" + tds.RandString(len)
	}
	return strings.TrimPrefix(randomString, "-")
}
func (tds TestDataGenerationSpec) RandString(n int) string {

	b := make([]byte, n)
	for i := range b {
		b[i] = _lc_RAND_VALUESTRING[tds.randomizer.Intn(len(_lc_RAND_VALUESTRING))]
	}
	return string(b)
}
func (tds TestDataGenerationSpec) RandANString(n int) string {

	b := make([]byte, n)
	for i := range b {
		b[i] = _lc_RANDOM_ALPHANUMVALUESTRING[tds.randomizer.Intn(len(_lc_RANDOM_ALPHANUMVALUESTRING))]
	}
	return string(b)
}
func (tds TestDataGenerationSpec) PickFromList(list []string) string {
	dataSize := len(list)
	return list[tds.randomizer.Intn(dataSize)]
}
func GenerateTestData(specBytes []byte) interface{} {
	return GenerateTestDataInMap(specBytes)
}
func GenerateTestDataInMap(specBytes []byte) map[string]interface{} {

	fieldSpecs := make([]TestDataGenerationSpec, 0)
	err := json.Unmarshal(specBytes, &fieldSpecs)
	if err != nil {
		fmt.Printf("\nSpec unmarshal Error %v", err)
		return nil
	}
	dataMap := make(map[string]interface{})
	for _, fieldSpec := range fieldSpecs {
		fieldSpec.Init()
		dataMap[fieldSpec.FieldName] = fieldSpec.GenerateData()
	}
	return dataMap
}
func Test_VNARandomDataGenerator(t *testing.T) {
	vmdSpec := `
	[
	{
		"fname":"erpsystem",
		"type":"pickList",
		
		"pickFromList":["SAP-1"],
		"prefix":"VNA-"
	},
	{
		"fname":"client",
		"type":"pickList",
		"pickFromList":["1010"]
		
	},
	{
		"fname":"vendorid",
		"type":"an-string",
		"size":8,
		"prefix":"VNA-"
	},
	{
		"fname":"vendorname",
		"type":"pickList",
		"pickFromList":["WIPRO"]
		
	}
	]
	`
	testData := GenerateTestData([]byte(buildSpec(vmdSpec)))
	if testData == nil {
		t.FailNow()
	}
	prettyJSON, _ := json.MarshalIndent(testData, " ", " ")
	t.Logf("%s", string(prettyJSON))
}
