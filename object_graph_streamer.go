package ObjectGraphStreamer

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	hashLib "hash"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"
)

const JSISOStringFormat = "2006-01-02T15:04:05.999Z07:00"

type ValType interface {
	ToString() *string
	AsValue() interface{}
}

type JsonValType struct {
	ValType
	Val interface{}
}

// type Dict T // map[string]interface{}

func (j JsonValType) ToString() *string {
	out, err := json.Marshal(j.Val)
	if err != nil {
		panic(err)
	}
	str := string(out)
	return &str
}

func (j JsonValType) AsValue() interface{} {
	return j.Val
}

type PlainValType struct {
	ValType
	val *string
}

func (p PlainValType) ToString() *string {
	return p.val
}

func (p PlainValType) AsValue() interface{} {
	return p.val
}

type OutState string

const (
	NONE         = ""
	VALUE        = "Value"
	ATTRIBUTE    = "Attr"
	ARRAY_START  = "["
	ARRAY_END    = "]"
	OBJECT_START = "{"
	OBJECT_END   = "}"
)

func (o OutState) String() string {
	switch o {
	case ATTRIBUTE:
		return ATTRIBUTE
	case VALUE:
		return VALUE
	case ARRAY_START:
		return ARRAY_START
	case ARRAY_END:
		return ARRAY_END
	case OBJECT_START:
		return OBJECT_START
	case OBJECT_END:
		return OBJECT_END
	}
	panic(fmt.Sprintf("Should not reached:%s", string(o)))
}

type SVal struct {
	Attribute string
	Val       ValType
	OutState  OutState
	Paths     []string
}

type SvalFn func(prob SVal)
type ObjectProccesorFn func(prob *[]string) *[]string
type ArrayProccesorFn func(prob *[]interface{}) *[]interface{}
type ValFactoryFn func(prob interface{}) ValType

type ObjectGraphStreamerProps struct {
	Paths           []string
	ObjectProcessor ObjectProccesorFn
	ArrayProcessor  ArrayProccesorFn
	ValFactory      ValFactoryFn
}

func (ogsp *ObjectGraphStreamerProps) assignPaths(paths []string) *ObjectGraphStreamerProps {
	return &ObjectGraphStreamerProps{
		Paths:           paths,
		ObjectProcessor: ogsp.ObjectProcessor,
		ArrayProcessor:  ogsp.ArrayProcessor,
		ValFactory:      ogsp.ValFactory,
	}
}

func defaultObjectGraphStreamerProps(e []ObjectGraphStreamerProps) *ObjectGraphStreamerProps {
	var ogsp ObjectGraphStreamerProps
	if len(e) == 0 {
		ogsp = ObjectGraphStreamerProps{}
	} else {
		ogsp = e[0]
	}
	if ogsp.Paths == nil {
		ogsp.Paths = []string{}
	}
	if ogsp.ObjectProcessor == nil {
		ogsp.ObjectProcessor = func(a *[]string) *[]string { sort.Strings(*a); return a }
	}
	if ogsp.ArrayProcessor == nil {
		ogsp.ArrayProcessor = func(a *[]interface{}) *[]interface{} { return a }
	}
	if ogsp.ValFactory == nil {
		ogsp.ValFactory = func(e interface{}) ValType { return JsonValType{Val: e} }
	}
	return &ogsp
}

func ObjectGraphStreamer(e interface{},
	out SvalFn,
	partialOGSP ...ObjectGraphStreamerProps) {
	ogsp := defaultObjectGraphStreamerProps(partialOGSP)

	_, isTime := e.(time.Time)
	k := reflect.Invalid
	if e != nil {
		k = reflect.TypeOf(e).Kind()
	}
	valOf := reflect.ValueOf(e)
	if k == reflect.Slice {
		arrayPath := append(ogsp.Paths[:], "[")
		out(SVal{Paths: arrayPath, OutState: ARRAY_START})
		for i := 0; i < valOf.Len(); i++ {
			ObjectGraphStreamer(valOf.Index(i).Interface(), out, *ogsp.assignPaths(
				append(ogsp.Paths[:], fmt.Sprintf("%d", i))))
		}
		out(SVal{OutState: ARRAY_END, Paths: append(ogsp.Paths[:], "]")})
		return
	} else if k == reflect.Struct && !isTime {
		objectPath := append(ogsp.Paths[:], "{")
		out(SVal{OutState: OBJECT_START, Paths: objectPath})
		keys := make([]string, 0, valOf.NumField())
		m := make(map[string]interface{})
		for i := 0; i < valOf.NumField(); i++ {
			fl := valOf.Type().Field(i)
			if !fl.IsExported() {
				panic(fmt.Sprintf("Field '%v' is not exported!", fl.Name))
			}
			fieldName := fl.Name
			t, hasTag := fl.Tag.Lookup("json")
			if hasTag {
				fieldName = t
			}
			m[fieldName] = valOf.Field(i).Interface()
			keys = append(keys, fieldName)
		}
		for _, key := range *ogsp.ObjectProcessor(&keys) {
			_path := append(objectPath[:], key)
			out(SVal{Attribute: key, OutState: ATTRIBUTE, Paths: _path})
			ObjectGraphStreamer(m[key], out, *ogsp.assignPaths(_path))
		}
		out(SVal{OutState: OBJECT_END, Paths: append(ogsp.Paths[:], "}")})
		return
	}
	if k == reflect.Map && !isTime {
		mapPaths := append(ogsp.Paths[:], "{")
		out(SVal{OutState: OBJECT_START, Paths: mapPaths})
		mappe := e.(map[string]interface{})
		keys := make([]string, len(mappe))
		idx := 0
		for key := range mappe {
			keys[idx] = key
			idx++
		}
		for _, key := range *ogsp.ObjectProcessor(&keys) {
			_path := append(mapPaths[:], key)
			out(SVal{Attribute: key, OutState: ATTRIBUTE, Paths: _path})
			ObjectGraphStreamer(mappe[key], out, *ogsp.assignPaths(_path))
		}
		out(SVal{OutState: OBJECT_END, Paths: append(ogsp.Paths[:], "}")})
		return
	}
	// else {
	//	fmt.Println("Reflect:", k)
	//}

	out(SVal{Val: JsonValType{Val: e}, OutState: VALUE, Paths: ogsp.Paths})
}

type OutputFN func(str string)

type JsonProps struct {
	indent  int
	newline string
}

func NewJsonProps(nSpaces int, newLine string) *JsonProps {
	nl := "\n"
	if newLine != "" {
		nl = newLine
	}
	return &JsonProps{
		indent:  nSpaces,
		newline: nl,
	}
}

type JsonCollector struct {
	output    OutputFN
	indent    string
	commas    []string
	elements  []int
	props     *JsonProps
	nextLine  string
	attribute string
}

func NewJsonCollector(o OutputFN, p *JsonProps) *JsonCollector {
	props := p
	if props == nil {
		props = NewJsonProps(0, "")
	}

	nextLine := ""
	if props.indent > 0 {
		nextLine = props.newline
	}

	return &JsonCollector{
		output:    o,
		indent:    strings.Repeat(" ", props.indent),
		commas:    []string{""},
		elements:  []int{0},
		props:     props,
		nextLine:  nextLine,
		attribute: "",
	}
}

func (j *JsonCollector) Suffix() string {
	if j.elements[len(j.elements)-1] > 0 {
		commas := len(j.commas)
		if commas > 0 {
			commas -= 1
		}
		return fmt.Sprintf("%v%v", j.nextLine, strings.Repeat(j.indent, commas))
	}
	return ""
}

func (j *JsonCollector) Append(sVal SVal) {
	if sVal.OutState != NONE {
		switch sVal.OutState {
		case ARRAY_START:
			j.output(fmt.Sprintf("%v%v%v[", j.commas[len(j.commas)-1], j.Suffix(), j.attribute))
			j.attribute = ""
			j.commas[len(j.commas)-1] = ","
			j.commas = append(j.commas, "")
			j.elements = append(j.elements, 0)
		case ARRAY_END:
			j.commas = j.commas[:len(j.commas)-1]
			j.output(fmt.Sprintf("%v]", j.Suffix()))
			j.elements = j.elements[:len(j.elements)-1]
		case OBJECT_START:
			j.output(fmt.Sprintf("%v%v%v{", j.commas[len(j.commas)-1], j.Suffix(), j.attribute))
			j.attribute = ""
			j.commas[len(j.commas)-1] = ","
			j.commas = append(j.commas, "")
			j.elements = append(j.elements, 0)
		case OBJECT_END:
			j.commas = j.commas[:len(j.commas)-1]
			j.output(fmt.Sprintf("%v}", j.Suffix()))
			j.elements = j.elements[:len(j.elements)-1]
		}
	}

	if sVal.Val != nil {
		j.elements[len(j.elements)-1]++
		j.output(fmt.Sprintf("%v%v%v%v", j.commas[len(j.commas)-1], j.Suffix(), j.attribute, *sVal.Val.ToString()))
		j.attribute = ""
		j.commas[len(j.commas)-1] = ","
	}

	if sVal.Attribute != "" {
		eidx := len(j.elements)
		if eidx != 0 {
			eidx -= 1
		}
		j.elements[eidx]++

		b, err := json.Marshal(sVal.Attribute)
		if err != nil {
			panic(err)
		}
		space := ""
		if len(j.indent) > 0 {
			space = " "
		}
		j.attribute = fmt.Sprintf("%v:%v", string(b), space)

	}
}

type HashCollector struct {
	hash hashLib.Hash
}

func NewHashCollector() *HashCollector {
	return &HashCollector{
		hash: sha256.New(),
	}
}

func (h *HashCollector) Digest() string {
	// b := []byte{}
	return base58.Encode(h.hash.Sum(nil))
}

func (h *HashCollector) Append(sval SVal) {
	if sval.OutState != NONE {
		return
	}

	// fmt.Println("SVAL", sval)
	if sval.Attribute != "" {
		// fmt.Println("ATTRIB", sval.attribute)
		h.hash.Write([]byte(sval.Attribute))
	}

	if sval.Val != nil {
		vl := sval.Val.AsValue()
		tval, isTime := vl.(time.Time)
		var t string
		if isTime {
			t = tval.Format(JSISOStringFormat)
		} else {
			t = fmt.Sprintf("%v", vl)
		}
		// fmt.Println("VAL", t)
		h.hash.Write([]byte(t))
	}
}

type JsonHash struct {
	JsonStr *string
	Hash    *string
}
