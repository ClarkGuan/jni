package tool

import (
	"bytes"
	"fmt"
	"strings"
)

// 表示 C 语言的类型
type cType struct {
	isPtr     bool
	isConst   bool
	typeName  string
	isVarArgs bool
}

func (c *cType) String() string {
	return fmt.Sprintf("{isPtr: %t, isConst: %t, typeName: %s, isVarArgs: %t}",
		c.isPtr, c.isConst, c.typeName, c.isVarArgs)
}

func (c *cType) isVoid() bool {
	return !c.isPtr && c.typeName == "void"
}

func (c *cType) toC() *cTypeOutput {
	return (*cTypeOutput)(c)
}

func (c *cType) toGo() *cTypeGoOutput {
	return (*cTypeGoOutput)(c)
}

type cTypeOutput cType

func (output *cTypeOutput) TypeDesc() string {
	s := output.typeName
	if output.isPtr {
		s += " *"
	}
	return s
}

type cTypeGoOutput cType

func (output *cTypeGoOutput) TypeDesc() string {
	if output.isPtr {
		switch output.typeName {
		case "JNIEnv":
			return "Env"

		case "JavaVM":
			return "VM"

		case "void":
			return "unsafe.Pointer"

		case "char":
			return "string"
		}
	} else {
		switch output.typeName {
		case "void":
			return ""

		case "jboolean":
			return "bool"

		case "jbyte":
			return "byte"

		case "jshort":
			return "int16"

		case "jchar":
			return "uint16"

		case "jsize", "jint":
			return "int"

		case "jlong":
			return "int64"

		case "jfloat":
			return "float32"

		case "jdouble":
			return "float64"

		case "jobject", "jclass", "jthrowable", "jstring",
			"jarray", "jbooleanArray", "jbyteArray", "jcharArray",
			"jshortArray", "jintArray", "jlongArray", "jfloatArray",
			"jdoubleArray", "jobjectArray", "jweak":
			return "J" + output.typeName[1:]

		case "jmethodID", "jfieldID":
			return "J" + output.typeName[1:]
		}
	}

	s := "C." + output.typeName
	if output.isPtr {
		s = "*" + s
	}
	return s
}

// 表示 C 语言的参数
type param struct {
	cType
	idName string
}

func (p *param) toC() *paramOutput {
	return (*paramOutput)(p)
}

func (p *param) toGo() *paramGoOutput {
	return (*paramGoOutput)(p)
}

type paramOutput param

func (param *paramOutput) paramDesc() string {
	if param.isPtr && param.typeName == "jboolean" && param.idName == "isCopy" {
		return ""
	}

	return param.cType.toC().TypeDesc() + " " + param.idName
}

func (param *paramOutput) callDesc() string {
	if param.isPtr && param.typeName == "jboolean" && param.idName == "isCopy" {
		return "NULL"
	}

	return param.idName
}

type paramGoOutput param

func (param *paramGoOutput) paramDesc() string {
	if param.isPtr && param.typeName == "jboolean" && param.idName == "isCopy" {
		return ""
	}

	return param.idName + " " + param.cType.toGo().TypeDesc()
}

func (param *paramGoOutput) paramListDesc() string {
	if param.isPtr {
		switch param.typeName {
		case "jboolean":
			return param.idName + " []bool"

		case "jbyte":
			return param.idName + " []byte"

		case "jshort":
			return param.idName + " []int16"

		case "jchar":
			return param.idName + " []uint16"

		case "jint":
			return param.idName + " []int32"

		case "jlong":
			return param.idName + " []int64"

		case "jfloat":
			return param.idName + " []float32"

		case "jdouble":
			return param.idName + " []float64"
		}
	}

	return param.idName + " " + param.cType.toGo().TypeDesc()
}

func (param *paramGoOutput) paramListWrapper() string {
	if param.isPtr {
		switch param.typeName {
		case "jboolean":
			return "cBooleanArray"

		case "jbyte":
			return "cByteArray"

		case "jshort":
			return "cShortArray"

		case "jchar":
			return "cCharArray"

		case "jint":
			return "cIntArray"

		case "jlong":
			return "cLongArray"

		case "jfloat":
			return "cFloatArray"

		case "jdouble":
			return "cDoubleArray"
		}
	}
	return ""
}

func (param *paramGoOutput) callDesc() string {
	if param.isPtr {
		switch param.typeName {
		case "JNIEnv":
			return fmt.Sprintf("(*C.JNIEnv)(unsafe.Pointer(%s))", param.idName)

		case "JavaVM":
			return fmt.Sprintf("(*C.JavaVM)(unsafe.Pointer(%s))", param.idName)

		case "char":
			return fmt.Sprintf("cstr_%s", param.idName)

		case "jboolean":
			if param.idName == "isCopy" {
				return ""
			}
		}

	}

	if !param.isPtr {
		switch param.typeName {
		case "jboolean":
			return fmt.Sprintf("cbool(%s)", param.idName)

		case "jbyte", "jshort", "jchar", "jsize", "jint", "jlong", "jfloat", "jdouble":
			return fmt.Sprintf("C.%s(%s)", param.typeName, param.idName)

		case "jobject", "jclass", "jthrowable", "jstring",
			"jarray", "jbooleanArray", "jbyteArray", "jcharArray",
			"jshortArray", "jintArray", "jlongArray", "jfloatArray",
			"jdoubleArray", "jobjectArray", "jweak":
			return fmt.Sprintf("C.%s(%s)", param.typeName, param.idName)

		case "jmethodID", "jfieldID":
			return fmt.Sprintf("C.%s(unsafe.Pointer(%s))", param.typeName, param.idName)
		}
	}

	return param.idName
}

// 表示 C 语言的函数
type method struct {
	name   string
	ret    *cType
	params []*param
}

func (m *method) hasRetVal() bool {
	return !m.ret.isVoid()
}

func (m *method) goRetVal() string {
	if !m.hasRetVal() {
		return ""
	}

	return " " + m.ret.toGo().TypeDesc()
}

func (m *method) isVarArgs() bool {
	return len(m.params) > 0 && m.params[len(m.params)-1].isVarArgs
}

func (m *method) isVaList() bool {
	for _, p := range m.params {
		if p.typeName == "va_list" {
			return true
		}
	}

	return false
}

func (m *method) String() string {
	return fmt.Sprintf("{name: %s, ret: %s, params: %v}", m.name, m.ret, m.params)
}

func (m *method) toC() *methodOutput {
	return (*methodOutput)(m)
}

func (m *method) toGo() *methodGoOutput {
	return (*methodGoOutput)(m)
}

type methodOutput method

func (output *methodOutput) paramList() string {
	if len(output.params) == 0 {
		return "void"
	}

	buf := bytes.NewBuffer(nil)
	for i, p := range output.params {
		desc := p.toC().paramDesc()
		if desc == "" {
			continue
		}

		if i > 0 {
			fmt.Fprintf(buf, ", ")
		}

		fmt.Fprint(buf, desc)
	}

	return buf.String()
}

func (output *methodOutput) callList() string {
	if len(output.params) == 0 {
		return ""
	}

	buf := bytes.NewBuffer(nil)
	for i, p := range output.params {
		desc := p.toC().callDesc()
		if desc == "" {
			continue
		}

		if i > 0 {
			fmt.Fprintf(buf, ", ")
		}

		fmt.Fprint(buf, desc)
	}

	return buf.String()
}

func (output *methodOutput) beforeReturn(in string) string {
	// 去掉返回值中的 const
	if output.ret.isConst {
		return fmt.Sprintf("(%s) %s", output.ret.toC().TypeDesc(), in)
	}
	return in
}

type methodGoOutput method

func (output *methodGoOutput) isArrayRegion() bool {
	return strings.HasSuffix(output.name, "ArrayRegion") && (strings.HasPrefix(output.name, "Get") || strings.HasPrefix(output.name, "Set"))
}

func (output *methodGoOutput) isCallFunc() bool {
	lastParam := output.params[len(output.params)-1]
	return (strings.HasPrefix(output.name, "Call") || strings.HasPrefix(output.name, "New")) &&
		strings.HasSuffix(output.name, "A") &&
		lastParam.isPtr && lastParam.typeName == "jvalue"
}

func (output *methodGoOutput) paramList() string {
	if len(output.params) == 1 {
		return ""
	}

	buf := bytes.NewBuffer(nil)

	var ps []*param
	if output.isArrayRegion() {
		ps = output.params[1 : len(output.params)-2]
	} else if output.isCallFunc() {
		ps = output.params[1 : len(output.params)-1]
	} else {
		ps = output.params[1:]
	}

	for i, p := range ps {
		desc := p.toGo().paramDesc()
		if desc == "" {
			continue
		}

		if i > 0 {
			fmt.Fprintf(buf, ", ")
		}

		fmt.Fprint(buf, desc)
	}

	if output.isArrayRegion() {
		fmt.Fprintf(buf, ", ")
		fmt.Fprint(buf, output.params[len(output.params)-1].toGo().paramListDesc())
	} else if output.isCallFunc() {
		fmt.Fprintf(buf, ", %s ...Jvalue", output.params[len(output.params)-1].idName)
	}

	return buf.String()
}

func (output *methodGoOutput) callList() string {
	if len(output.params) == 0 {
		return ""
	}

	buf := bytes.NewBuffer(nil)

	var ps []*param
	if output.isArrayRegion() {
		ps = output.params[:len(output.params)-2]
	} else if output.isCallFunc() {
		ps = output.params[:len(output.params)-1]
	} else {
		ps = output.params
	}

	for i, p := range ps {
		desc := p.toGo().callDesc()
		if desc == "" {
			continue
		}

		if i > 0 {
			fmt.Fprintf(buf, ", ")
		}

		fmt.Fprint(buf, desc)
	}

	if output.isArrayRegion() {
		fmt.Fprintf(buf, ", ")
		lastParam := output.params[len(output.params)-1]
		fmt.Fprintf(buf, "C.jsize(len(%s))", lastParam.idName)
		fmt.Fprintf(buf, ", ")
		fmt.Fprintf(buf, "%s(%s)", lastParam.toGo().paramListWrapper(), lastParam.idName)
	} else if output.isCallFunc() {
		fmt.Fprintf(buf, ", cvals(%s)", output.params[len(output.params)-1].idName)
	}

	return buf.String()
}

func (output *methodGoOutput) beforeReturn(in string) string {
	// char * -> string
	if output.ret.isPtr {
		switch output.ret.typeName {
		case "char":
			return fmt.Sprintf("C.GoString(%s)", in)
		}
	}

	if !output.ret.isPtr {
		switch output.ret.typeName {
		case "jboolean":
			return fmt.Sprintf("%s != C.JNI_FALSE", in)

		case "jbyte":
			return fmt.Sprintf("byte(%s)", in)

		case "jshort":
			return fmt.Sprintf("int16(%s)", in)

		case "jchar":
			return fmt.Sprintf("uint16(%s)", in)

		case "jsize", "jint":
			return fmt.Sprintf("int(%s)", in)

		case "jlong":
			return fmt.Sprintf("int64(%s)", in)

		case "jfloat":
			return fmt.Sprintf("float32(%s)", in)

		case "jdouble":
			return fmt.Sprintf("float64(%s)", in)

		case "jobject", "jclass", "jthrowable", "jstring",
			"jarray", "jbooleanArray", "jbyteArray", "jcharArray",
			"jshortArray", "jintArray", "jlongArray", "jfloatArray",
			"jdoubleArray", "jobjectArray", "jweak":
			return fmt.Sprintf("%s(%s)", "J"+output.ret.typeName[1:], in)

		case "jfieldID", "jmethodID":
			return fmt.Sprintf("%s(unsafe.Pointer(%s))", "J"+output.ret.typeName[1:], in)
		}
	}

	return in
}

func (output *methodGoOutput) prepareReturn(buf *bytes.Buffer) {
	for _, p := range output.params {
		if p.isPtr && p.typeName == "char" {
			fmt.Fprintf(buf, "\tcstr_%s := C.CString(%s)\n", p.idName, p.idName)
			fmt.Fprintf(buf, "\tdefer C.free(unsafe.Pointer(cstr_%s))\n", p.idName)
		}
	}
}
