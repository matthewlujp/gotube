package gotube

import (
	"reflect"
	"testing"
)

var (
	invalidJs = []byte("var a = 1 + 1;\nvar hello = function(){ return \"hello\"}")
)

func TestExtractDecipherFuncName(t *testing.T) {
	js, errJsRead := readCompressedFile(mockScriptPath)
	if errJsRead != nil {
		t.Fatal(errJsRead)
	}

	expected := "FK"
	if fName, err := extractDecipherFuncName(js); err != nil {
		t.Error("failed to extract decipher function name")
	} else if !reflect.DeepEqual(fName, expected) {
		t.Errorf("wrong decipher function name extracted, got %s, expected %s", fName, expected)
	}
}

func TestExtractDecipherProcedure(t *testing.T) {
	js, errJsRead := readCompressedFile(mockScriptPath)
	if errJsRead != nil {
		t.Fatal(errJsRead)
	}

	expected := []string{"EK.Ck(a,11)", "EK.ml(a,1)", "EK.Ck(a,24)", "EK.aJ(a,50)"}
	if procedure, err := extractDecipherProcedure("FK", js); err != nil {
		t.Error("failed to extract decipher procedure")
	} else if !reflect.DeepEqual(procedure, expected) {
		t.Errorf("wrong decipher procedure extracted, got %v, expected %v", procedure, expected)
	}
}

func TestExtractConverters(t *testing.T) {
	js, errJsRead := readCompressedFile(mockScriptPath)
	if errJsRead != nil {
		t.Fatal(errJsRead)
	}

	expected := map[string]string{
		"Ck": "function(a){a.reverse()}",
		"aJ": "function(a,b){var c=a[0];a[0]=a[b%a.length];a[b%a.length]=c}",
		"ml": "function(a,b){a.splice(0,b)}",
	}
	if strConverters, err := extractConverters("EK", js); err != nil {
		t.Errorf("failed to extract raw converters, %s")
	} else if !reflect.DeepEqual(strConverters, expected) {
		t.Errorf("wrong raw converters extracted, got %s, expected %s", strConverters, expected)
	}
}

func TestBuildConverterMap(t *testing.T) {
	strConverterMap := map[string]string{
		"Ck": "function(a){a.reverse()}",
		"aJ": "function(a,b){var c=a[0];a[0]=a[b%a.length];a[b%a.length]=c}",
		"ml": "function(a,b){a.splice(0,b)}",
	}
	converterMap := buildConverterMap(strConverterMap)
	expected := map[string]*converter{
		"Ck": &converter{convertType: Reverse},
		"aJ": &converter{convertType: Swap},
		"ml": &converter{convertType: Splice},
	}
	for k, c := range expected {
		if v, ok := converterMap[k]; !ok {
			t.Errorf("%s is not included in the built map %v", k, converterMap)
		} else if c.convertType != v.convertType {
			t.Errorf("wrong convert type for %s, got %d, expected %d", k, v.convertType, c.convertType)
		}
	}
	if !reflect.DeepEqual(converterMap, expected) {
		t.Errorf("wrong converter map is built, got %v, expected %v", converterMap, expected)
	}
}

func TestBuildDecipherer(t *testing.T) {
	// js that does not contain decipher code
	d, err := newDecipherer(invalidJs)
	if err == nil {
		t.Error("error should be raised when reading invalid js code")
	}

	// correct js
	js, errJsRead := readCompressedFile(mockScriptPath)
	if errJsRead != nil {
		t.Fatal(errJsRead)
	}
	d, err = newDecipherer(js)
	if err != nil {
		t.Errorf("failed to build decipherer, %s", err)
	}

	expectedConverters := []*converter{
		&converter{params: []int{11}, convertType: Reverse},
		&converter{params: []int{1}, convertType: Splice},
		&converter{params: []int{24}, convertType: Reverse},
		&converter{params: []int{50}, convertType: Swap},
	}

	for i, c := range expectedConverters {
		if c.convertType != d.converters[i].convertType {
			t.Errorf("%d-th converter wrong, got type %d, expected type %d", i, d.converters[i].convertType, c.convertType)
		} else if !reflect.DeepEqual(c.params, d.converters[i].params) {
			t.Errorf("%d-th converter wrong, got params %v, expected params %v", i, d.converters[i].params, c.params)
		}
	}
}

func TestSplice(t *testing.T) {
	arr := []byte("abcdefg")
	c := converter{
		params:      []int{3},
		convertType: Splice,
	}
	expected := []byte("defg")
	if converted, err := c.convert(arr); err != nil {
		t.Errorf("splice failed, %s", err)
	} else if !reflect.DeepEqual(converted, expected) {
		t.Errorf("splice wrong, got %s, expected %s", converted, expected)
	}
}

func TestSwap(t *testing.T) {
	arr := []byte("abcdefg")
	c := converter{
		params:      []int{3},
		convertType: Swap,
	}
	expected := []byte("dbcaefg")
	if converted, err := c.convert(arr); err != nil {
		t.Errorf("swap failed, %s", err)
	} else if !reflect.DeepEqual(converted, expected) {
		t.Errorf("swap wrong, got %s, expected %s", converted, expected)
	}
}

func TestReverse(t *testing.T) {
	arr := []byte("abcdefg")
	c := converter{
		convertType: Reverse,
	}
	expected := []byte("gfedcba")
	if converted, err := c.convert(arr); err != nil {
		t.Errorf("reverse failed, %s", err)
	} else if !reflect.DeepEqual(converted, expected) {
		t.Errorf("reverse wrong, got %s, expected %s", converted, expected)
	}
}

func TestDecipher(t *testing.T) {
	content, errJsRead := readCompressedFile(mockScriptPath)
	if errJsRead != nil {
		t.Fatal(errJsRead)
	}
	d, err := newDecipherer(content)
	if err != nil {
		t.Errorf("failed to build decipherer, %s", err)
	}
	cipheredSignature := "55B2F7214E041D71337613EA784BC0797F6064EE.6497710E7951322E10ACCD773A9DA4B7A492B6E77"
	decipheredSignature := "95B2F7214E041D71337613EA784BC0797F6064EE.6497710E7551322E10ACCD773A9DA4B7A492B6E7"
	if deciphered, err := d.Decipher(cipheredSignature); err != nil {
		t.Errorf("decipher failed, %s", err)
	} else if deciphered != decipheredSignature {
		t.Errorf("decipher wrong, got %s, expected %s", deciphered, decipheredSignature)
	}
}
