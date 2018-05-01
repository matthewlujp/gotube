package gotube

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	funcNameRegex = regexp.MustCompile(`signature",\s?([a-zA-Z0-9$]+)\(`)
	opRegex       = regexp.MustCompile(`\w+?\.(\w{2})\(\w,([0-9]+)\)`)
)

type decipherer interface {
	Decipher(string) (string, error)
}

type youtubeDecipherer struct {
	converters []*converter
}

func newDecipherer(js []byte) (*youtubeDecipherer, error) {
	funcName, errExtractFuncName := extractDecipherFuncName(js)
	if errExtractFuncName != nil {
		return nil, errExtractFuncName
	}
	procedure, errExtractProcedure := extractDecipherProcedure(funcName, js)
	if errExtractProcedure != nil {
		return nil, errExtractProcedure
	}
	convertersDictName := strings.Split(procedure[0], ".")[0] // extract "EK"
	strConverters, errExtractConverters := extractConverters(convertersDictName, js)
	if errExtractConverters != nil {
		return nil, errExtractConverters
	}
	fMap := buildConverterMap(strConverters)

	// build decipherer
	converters := make([]*converter, 0, len(procedure))
	for _, p := range procedure {
		res := opRegex.FindStringSubmatch(p)
		if res == nil {
			return nil, fmt.Errorf("failed to extract operation from %s", p)
		}
		param, err := strconv.Atoi(res[2])
		if err != nil {
			return nil, err
		}
		converters = append(
			converters,
			&converter{convertType: fMap[res[1]].convertType, params: []int{param}},
		)
	}

	d := youtubeDecipherer{converters: converters}
	return &d, nil
}

// extractDecipherFuncName returns something like "FK"
func extractDecipherFuncName(js []byte) (string, error) {
	res := funcNameRegex.FindSubmatch(js)
	if res == nil || len(res) < 2 {
		return "", errors.New("no function name extracted")
	}
	return string(res[1][:]), nil
}

// extractDecipherProcedures returns something like ["EK.Ck(a,11)", "EK.ml(a,1)", "EK.Ck(a,24)", "EK.aJ(a,50)"]
func extractDecipherProcedure(fName string, js []byte) ([]string, error) {
	decipherProcedureRegex, err := regexp.Compile(fName + `=function\(\w\){[a-z=\.\(\"\)]*;(.*);(?:.+)}`)
	if err != nil {
		return nil, err
	}
	decipherProcedure := decipherProcedureRegex.FindSubmatch(js)
	if decipherProcedure == nil || len(decipherProcedure) < 2 {
		return nil, errors.New("no procedure extraced")
	}
	return strings.Split(string(decipherProcedure[1][:]), ";"), nil
}

// extractConverters returns map of converters
func extractConverters(convertersDictName string, js []byte) (map[string]string, error) {
	convertersPattern := fmt.Sprintf(`var\s%s=\{([\s\S]+?)\};`, convertersDictName)
	r, err := regexp.Compile(convertersPattern)
	if err != nil {
		return nil, err
	}
	// "Ck:function(a){a.reverse()},\naJ:function(a,b){var c=a[0];a[0]=a[b%a.length];a[b%a.length]=c},\nml:function(a,b){a.splice(0,b)}"
	funcs := r.FindSubmatch(js)
	if funcs == nil || len(funcs) < 2 {
		return nil, errors.New("no functions extracted")
	}
	// ["Ck:function(a){a.reverse()}", "aJ:function(a,b){var c=a[0];a[0]=a[b%a.length];a[b%a.length]=c}", "ml:function(a,b){a.splice(0,b)}"]
	labelAndFuncs := strings.Split((strings.Replace(string(funcs[1][:]), "\n", " ", -1)), ", ")
	strFuncMap := make(map[string]string, len(labelAndFuncs))
	for _, lf := range labelAndFuncs {
		res := strings.Split(lf, ":")
		strFuncMap[res[0]] = res[1]
	}
	return strFuncMap, nil
}

func buildConverterMap(m map[string]string) map[string]*converter {
	fMap := make(map[string]*converter)
	for k, f := range m {
		if strings.Contains(f, "reverse") {
			fMap[k] = &converter{convertType: Reverse}
		} else if strings.Contains(f, "splice") {
			fMap[k] = &converter{convertType: Splice}
		} else if strings.Contains(f, "{var ") {
			fMap[k] = &converter{convertType: Swap}
		}
	}
	return fMap
}

func (d *youtubeDecipherer) Decipher(encryptedSignature string) (string, error) {
	byteSig := []byte(encryptedSignature)
	for _, c := range d.converters {
		s, err := c.convert(byteSig)
		if err != nil {
			return "", err
		}
		byteSig = byteSig[:len(s)]
		copy(byteSig, s)
	}
	return string(byteSig[:]), nil
}

type ConvertType int

const (
	Reverse ConvertType = iota
	Splice
	Swap
)

type converter struct {
	params      []int
	convertType ConvertType
}

func (c *converter) convert(v []byte) ([]byte, error) {
	switch c.convertType {
	case Reverse:
		return c.reverse(v)
	case Splice:
		return c.splice(v)
	case Swap:
		return c.swap(v)
	default:
		return nil, errors.New("undefined convert type")
	}
}

func (c *converter) reverse(v []byte) ([]byte, error) {
	ret := make([]byte, 0, len(v))
	for i := 0; i < len(v); i++ {
		ret = append(ret, v[len(v)-1-i])
	}
	return ret, nil
}

func (c *converter) splice(v []byte) ([]byte, error) {
	if len(c.params) < 1 {
		return nil, errors.New("no params empty")
	} else if c.params[0] >= len(v) {
		return nil, errors.New("param is larger than length of v")
	}
	ret := make([]byte, 0, len(v)-c.params[0])
	for _, b := range v[c.params[0]:] {
		ret = append(ret, b)
	}
	return ret, nil
}

func (c *converter) swap(v []byte) ([]byte, error) {
	if len(c.params) < 1 {
		return nil, errors.New("no params empty")
	} else if c.params[0] >= len(v) {
		return nil, errors.New("param is larger than length of v")
	}
	ret := make([]byte, 0, len(v))
	for i, b := range v {
		if i == 0 {
			ret = append(ret, v[c.params[0]%len(v)])
		} else if i == c.params[0]%len(v) {
			ret = append(ret, v[0])
		} else {
			ret = append(ret, b)
		}
	}
	return ret, nil
}
