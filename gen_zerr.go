//go:build ignore

package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/rs/zerolog"
)

var excludedMethods []string = []string{
	// not propagated
	"AnErr",
	"Err",
	"Discard",
	"Enabled",
	"GetCtx",
	// handled manually
	"Msg",
	"MsgFunc",
	"Msgf",
	"Send",
}

func main() {
	logger := zerolog.New(os.Stdout)
	e := logger.Error()

	r := reflect.TypeOf(e)
	fmt.Println(r.NumMethod())

	mes := []*Method{}
	for i := range r.NumMethod() {
		m := r.Method(i)
		if slices.Contains(excludedMethods, m.Name) {
			continue
		}

		me, _ := parseMethod(m)
		// fmt.Printf("%+v\n", me)
		mes = append(mes, me)
	}

	f, err := os.Create("./internal/errors/zerr.go")
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()

	f.WriteString("package errors\n")

	for _, me := range mes {
		fmt.Println(me.RequiredImports)
		ins := []string{}
		for i, pn := range me.ParameterNames {
			ins = append(ins, fmt.Sprintf("%s %s", pn, me.ParameterTypes[i]))
		}
		f.WriteString(fmt.Sprintf(methodProto, me.Name, strings.Join(ins, ", "), strings.Join(me.ParameterNames, ", ")))
	}

}

type Method struct {
	Name            string
	ParameterNames  []string
	ParameterTypes  []string
	RequiredImports []string
}

var methodProto = `
func (err *Error) %[1]s(%[2]s) *Error {
    err.e.%[1]s(%[3]s)
	return err
}
`

var moduleProto = `
package errors

import (
%s
)

%s
`

func parseMethod(m reflect.Method) (*Method, error) {
	if m.Type.NumOut() != 1 {
		// FIXME panic if return type of method is not *zerolog.Event
		//  || m.Type.Out(0) != nil
		return nil, errors.New("oops")
	}

	ni := m.Type.NumIn() - 1

	var pns []string
	switch ni {
	case 0:
	case 1:
		pns = append(pns, "value")
	case 2:
		pns = append(pns, "key", "value")
	default:
		pns = append(pns, "key")
		for i := range ni - 1 {
			pns = append(pns, fmt.Sprintf("value%d", i+1))
		}
	}

	pts := []string{}
	ris := []string{}
	for i := range ni {
		in := m.Type.In(i + 1)
		pts = append(pts, in.Name())
		if ri := in.PkgPath(); ri != "" {
			ris = append(ris, ri)
		}
	}

	// b := strings.Builder{}
	// b.WriteString(fmt.Sprintf("func (err *zerr.Error) %s() *zerr.Error {}", m.Name))
	// return b.String()
	// m.Func.FieldByIndex([]int{0})

	// return fmt.Sprint(m.Name, "  ", m.Type)

	return &Method{Name: m.Name, ParameterNames: pns, ParameterTypes: pts, RequiredImports: slices.Compact(ris)}, nil
}

func Foo(string, int) {

}
