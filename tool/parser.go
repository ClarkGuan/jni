package tool

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
)

var identifier = regexp.MustCompile(`^\w+`)
var constr = []byte("const")
var methodPtr = regexp.MustCompile(`^\(JNICALL\s+\*(\w+)\)`)

func parseIdentifier(code []byte) (id *cType, next []byte, err error) {
	if code[0] == ',' || unicode.IsSpace(rune(code[0])) {
		next = code[1:]
		next = bytes.TrimLeftFunc(next, unicode.IsSpace)
	} else {
		next = code
	}

	// 判断是否是 varArgs
	if next[0] == '.' && next[1] == '.' && next[2] == '.' {
		id = &cType{isVarArgs: true}
		next = bytes.TrimLeftFunc(next[3:], unicode.IsSpace)
		return
	}

	loc := identifier.FindIndex(next)
	if len(loc) == 0 {
		err = fmt.Errorf("%q 找不到标识符1", code)
		return
	}

	id = new(cType)

	// 判断是否以 const 开头
	if bytes.Equal(next[loc[0]:loc[1]], constr) {
		id.isConst = true
		next = bytes.TrimLeftFunc(next[loc[1]:], unicode.IsSpace)

		if next[0] == '*' {
			err = fmt.Errorf("%q 处理不了 const * xxx 的情况", code)
			id = nil
			return
		}

		loc = identifier.FindIndex(next)
		if len(loc) == 0 {
			err = fmt.Errorf("%q 找不到标识符2", code)
			return
		}
	}

	id.typeName = string(next[loc[0]:loc[1]])
	next = bytes.TrimLeftFunc(next[loc[1]:], unicode.IsSpace)
	if next[0] == '*' {
		id.isPtr = true
		next = bytes.TrimLeftFunc(next[1:], unicode.IsSpace)
	}
	if next[0] == '*' {
		id = nil
		err = fmt.Errorf("%q 处理不了多重指针的情况", code)
	}

	return
}

func parseFuncPointerName(code []byte) (name string, next []byte, err error) {
	loc := methodPtr.FindSubmatchIndex(code)
	if len(loc) == 0 {
		err = fmt.Errorf("%q 找不到函数指针名称", code)
		return
	} else if len(loc) < 4 {
		err = fmt.Errorf("%q 找不到函数指针名称分组", code)
	} else {
		name = string(code[loc[2]:loc[3]])
	}
	next = bytes.TrimLeftFunc(code[loc[1]:], unicode.IsSpace)
	return
}

func parseFuncParamList(code []byte) (params []*param, next []byte, err error) {
	if code[0] == '(' {
		next = code[1:]
	} else {
		next = code
	}

	var t *cType
	var loc []int

	for next[0] != ')' {
		if t, next, err = parseIdentifier(next); err != nil {
			return
		}

		loc = identifier.FindIndex(next)
		// varargs 的情况
		if len(loc) == 0 {
			if t.isVarArgs && next[0] == ')' {
				params = append(params, &param{cType: *t})
			} else {
				err = fmt.Errorf("%q 解析时出现异常", next)
			}

			break
		}

		idName := string(next[loc[0]:loc[1]])
		// 避免形参名称和 Go 保留字冲突
		if idName == "string" {
			idName = "str"
		}
		params = append(params, &param{cType: *t, idName: idName})
		next = bytes.TrimLeftFunc(next[loc[1]:], unicode.IsSpace)
	}

	return
}

func parseSingleMethod(code []byte) (m *method, err error) {
	// 找返回值
	var ret *cType
	if ret, code, err = parseIdentifier(code); err != nil {
		return
	}

	// 找函数指针名称
	var name string
	if name, code, err = parseFuncPointerName(code); err != nil {
		return
	}

	// 找函数参数列表
	var params []*param
	if params, code, err = parseFuncParamList(code); err != nil {
		return
	}

	return &method{ret: ret, name: name, params: params}, nil
}

func parseJniMethodList(code string) (list []*method) {
	lines := strings.Split(code, ";")

	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])

		if len(lines[i]) > 0 {
			if m, err := parseSingleMethod([]byte(lines[i])); err != nil {
				fmt.Fprintf(os.Stderr, "// // 匹配 %q\n", lines[i])
				fmt.Fprintln(os.Stderr, "// //", err)
			} else {
				//fmt.Printf("%s\n", m)
				list = append(list, m)
			}
		}
	}

	return
}
