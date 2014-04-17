package encoding

import (
	"encoding/json"
	"encoding/xml"
	"testing"
)

type exampleObjectForNullInt64 struct {
	NullValue NullInt64
}

func TestXMLMarshallingNullInt(t *testing.T) {

	e1 := exampleObjectForNullInt64{}

	expectedResult := "<exampleObjectForNullInt64><NullValue xsi:nil=\"true\"></NullValue></exampleObjectForNullInt64>"
	marshalAndCheckResult(e1, expectedResult, t)

	expectedResult = "<exampleObjectForNullInt64><NullValue>50</NullValue></exampleObjectForNullInt64>"
	e1.NullValue.Int64Value = 50
	e1.NullValue.IsValid = true
	marshalAndCheckResult(e1, expectedResult, t)

	expectedResult = "<exampleObjectForNullInt64><NullValue xsi:nil=\"true\"></NullValue></exampleObjectForNullInt64>"
	e1.NullValue.Int64Value = 100
	e1.NullValue.IsValid = false
	marshalAndCheckResult(e1, expectedResult, t)
}

func TestXMLUnmarshalNullInt(t *testing.T) {
	marshalledXML := "<exampleObjectForNullInt64><NullValue xsi:nil=\"true\"></NullValue></exampleObjectForNullInt64>"
	e1 := exampleObjectForNullInt64{}
	if err := xml.Unmarshal([]byte(marshalledXML), &e1); err != nil {
		t.Fatalf("Unable to unmarshal into xml object: %+v", err)
	}

	if e1.NullValue.IsValid {
		t.Fatal("Expected the value to be null but it wasnt")
	}

	marshalledXML = "<exampleObjectForNullInt64><NullValue>100</NullValue></exampleObjectForNullInt64>"
	if err := xml.Unmarshal([]byte(marshalledXML), &e1); err != nil {
		t.Fatalf("Unable to unmarshal into xml object: %+v", err)
	}

	if e1.NullValue.Int64Value != 100 {
		t.Fatalf("Value expected to be %d instead it was %d", 100, e1.NullValue.Int64Value)
	}

	if !e1.NullValue.IsValid {
		t.Fatal("Expected the value to be null but it wasnt")
	}

}

func marshalAndCheckResult(e1 exampleObjectForNullInt64, expectedResult string, t *testing.T) {
	xmlData, err := xml.Marshal(&e1)
	if err != nil {
		t.Fatalf("Error when trying to marshall nultInt: %+v", err)
	}

	output := string(xmlData)
	if output != expectedResult {
		t.Fatalf("Expected marshalling of object to result in %s but instead it was %s", expectedResult, output)
	}
}

func TestJSONMarshallingNullInt(t *testing.T) {
	e1 := exampleObjectForNullInt64{}
	expectedResult := "{\"NullValue\":null}"
	marshalJsonAndCheckResult(e1, expectedResult, t)

	e1 = exampleObjectForNullInt64{
		NullValue: NullInt64{
			IsValid:    true,
			Int64Value: 123,
		},
	}

	expectedResult = "{\"NullValue\":123}"
	marshalJsonAndCheckResult(e1, expectedResult, t)

	e1.NullValue.IsValid = false
	expectedResult = "{\"NullValue\":null}"
	marshalJsonAndCheckResult(e1, expectedResult, t)

	e1.NullValue.IsValid = true
	e1.NullValue.Int64Value = 1
	expectedResult = "{\"NullValue\":1}"
	marshalJsonAndCheckResult(e1, expectedResult, t)
}

func TestJsonUnmarshalNullInt(t *testing.T) {
	marshalledJson := "{\"NullValue\":null}"
	e1 := exampleObjectForNullInt64{}

	if err := json.Unmarshal([]byte(marshalledJson), &e1); err != nil {
		t.Fatalf("Unable to unmarshal json: %+v", err)
	}

	if e1.NullValue.IsValid {
		t.Fatal("Null value should indicate that the value is null but it doesnt")
	}

	marshalledJson = "{\"NullValue\":10}"
	if err := json.Unmarshal([]byte(marshalledJson), &e1); err != nil {
		t.Fatalf("Unable to unmarshal json: %+v", err)
	}

	if e1.NullValue.Int64Value != 10 {
		t.Fatalf("Value should be 10 instead it is %d", e1.NullValue.Int64Value)
	}

	if !e1.NullValue.IsValid {
		t.Fatal("Should not indicate that its null but it does")
	}

	marshalledJson = "{\"NullValue\":1}"
	if err := json.Unmarshal([]byte(marshalledJson), &e1); err != nil {
		t.Fatalf("Unable to unmarshal json: %+v", err)
	}

	if e1.NullValue.Int64Value != 1 {
		t.Fatalf("Value should be 1 instead it is %d", e1.NullValue.Int64Value)
	}

	if !e1.NullValue.IsValid {
		t.Fatal("Should not indicate that its null but it does")
	}

}

func marshalJsonAndCheckResult(e1 exampleObjectForNullInt64, expectedResult string, t *testing.T) {
	jsonData, err := json.Marshal(&e1)
	if err != nil {
		t.Fatalf("Unable to marshal json: %+v", err.Error())
	}

	output := string(jsonData)
	if output != expectedResult {
		t.Fatalf("Expected marshalling of json object to result in %s but instead it was %s", expectedResult, output)
	}
}
