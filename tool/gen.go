package tool

import (
	"bytes"
	"fmt"
)

func GenerateCode(pkg string) string {
	return generateCode(pkg, parseJniMethodList(headerCode))
}

func generateCode(pkg string, list []*method) string {
	buf := bytes.NewBuffer(nil)

	fmt.Fprintf(buf, "package %s\n", pkg)
	fmt.Fprint(buf, `
//
// #include <jni.h>
// #include <stdlib.h>
//
// static inline jint AttachCurrentThread(JavaVM *vm, JNIEnv **p_env) {
//     return (*vm)->AttachCurrentThread(vm, (void **) p_env, NULL);
// }
//
// static inline jint AttachCurrentThreadAsDaemon(JavaVM *vm, JNIEnv **p_env) {
//     return (*vm)->AttachCurrentThreadAsDaemon(vm, (void **) p_env, NULL);
// }
//
// static inline jint GetEnv(JavaVM *vm, JNIEnv **penv, jint version) {
//     return (*vm)->GetEnv(vm, (void **) penv, version);
// }
//
// static inline jint GetJavaVM(JNIEnv * env, JavaVM **vm) {
//     return (*env)->GetJavaVM(env, vm);
// }
//
// static inline int GetObjectRefType(JNIEnv * env, jobject obj) {
//     return (int) (*env)->GetObjectRefType(env, obj);
// }
//
`)

	for _, m := range list {
		if skip, err := generateFuncCode(m, buf); err == nil {
			if !skip {
				fmt.Fprintln(buf, "//")
			}
		}
	}

	fmt.Fprint(buf, `import "C"
import (
	"unicode/utf16"
	"unsafe"
)

const (
	JNI_VERSION_1_1 = 0x00010001
	JNI_VERSION_1_2 = 0x00010002
	JNI_VERSION_1_4 = 0x00010004
	JNI_VERSION_1_6 = 0x00010006

	JNI_FALSE = 0
	JNI_TRUE = 1

	JNI_OK = 0                 /* success */
	JNI_ERR = (-1)              /* unknown error */
	JNI_EDETACHED = (-2)              /* thread detached from the VM */
	JNI_EVERSION = (-3)              /* JNI version error */
	JNI_ENOMEM = (-4)              /* not enough memory */
	JNI_EEXIST = (-5)              /* VM already created */
	JNI_EINVAL = (-6)              /* invalid arguments */

	JNI_COMMIT = 1
	JNI_ABORT = 2
)

type RefType int

const (
	Invalid RefType = iota
	Local
	Global
	WeakGlobal
)

type Jobject = uintptr
type Jclass = uintptr
type Jthrowable = uintptr
type Jstring = uintptr
type Jarray = uintptr
type JbooleanArray = uintptr
type JbyteArray = uintptr
type JcharArray = uintptr
type JshortArray = uintptr
type JintArray = uintptr
type JlongArray = uintptr
type JfloatArray = uintptr
type JdoubleArray = uintptr
type JobjectArray = uintptr
type Jweak = uintptr
type Jvalue = uint64

type JmethodID = uintptr
type JfieldID = uintptr

func CMalloc(capacity int) unsafe.Pointer {
	return C.malloc(C.size_t(capacity))
}

func CFree(p unsafe.Pointer) {
	C.free(p)
}

func OfSlice(b []byte) unsafe.Pointer {
	return unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&b)))
}

type VM uintptr

func (vm VM) AttachCurrentThread() (Env, int) {
	var env *C.JNIEnv
	ret := int(C.AttachCurrentThread((*C.JavaVM)(unsafe.Pointer(vm)), &env))
	return Env(unsafe.Pointer(env)), ret
}

func (vm VM) AttachCurrentThreadAsDaemon() (Env, int) {
	var env *C.JNIEnv
	ret := int(C.AttachCurrentThreadAsDaemon((*C.JavaVM)(unsafe.Pointer(vm)), &env))
	return Env(unsafe.Pointer(env)), ret
}

func (vm VM) GetEnv(version int) (Env, int) {
	var env *C.JNIEnv
	ret := int(C.GetEnv((*C.JavaVM)(unsafe.Pointer(vm)), &env, C.jint(version)))
	return Env(unsafe.Pointer(env)), ret
}

type Env uintptr

func (env Env) GetJavaVM() (VM, int) {
	var vm *C.JavaVM
	ret := int(C.GetJavaVM((*C.JNIEnv)(unsafe.Pointer(env)), &vm))
	return VM(unsafe.Pointer(vm)), ret
}

func (env Env) GetObjectRefType(obj Jobject) RefType {
	return RefType(C.GetObjectRefType((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(obj)))
}

func (env Env) NewString(s string) Jstring {
	codes := utf16.Encode([]rune(s))
	size := len(codes)
	if size <= 0 {
		return 0
	} else {
		return Jstring(C.NewString((*C.JNIEnv)(unsafe.Pointer(env)), (*C.jchar)(unsafe.Pointer(&codes[0])), C.jsize(size)))
	}
}

func (env Env) GetStringUTF(ptr Jstring) []byte {
	jstr := C.jstring(ptr)
	size := C.GetStringUTFLength((*C.JNIEnv)(unsafe.Pointer(env)), jstr)
	ret := make([]byte, int(size))
	C.GetStringUTFRegion((*C.JNIEnv)(unsafe.Pointer(env)), jstr, C.jsize(0), C.GetStringLength((*C.JNIEnv)(unsafe.Pointer(env)), jstr), cmem(ret))
	return ret
}

func (env Env) NewDirectByteBuffer(address unsafe.Pointer, capacity int) Jobject {
	return Jobject(C.NewDirectByteBuffer((*C.JNIEnv)(unsafe.Pointer(env)), address, C.jlong(capacity)))
}

func (env Env) GetDirectBufferAddress(buf Jobject) unsafe.Pointer {
	return C.GetDirectBufferAddress((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(buf))
}

func (env Env) GetDirectBufferCapacity(buf Jobject) int {
	return int(C.GetDirectBufferCapacity((*C.JNIEnv)(unsafe.Pointer(env)), C.jobject(buf)))
}

func (env Env) GetBooleanArrayElement(array JbooleanArray, index int) bool {
	var ret C.jboolean
	C.GetBooleanArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jbooleanArray(array), C.jsize(index), C.jsize(1), &ret)
	return ret != C.JNI_FALSE
}

func (env Env) GetByteArrayElement(array JbyteArray, index int) byte {
	var ret C.jbyte
	C.GetByteArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jbyteArray(array), C.jsize(index), C.jsize(1), &ret)
	return byte(ret)
}

func (env Env) GetCharArrayElement(array JcharArray, index int) uint16 {
	var ret C.jchar
	C.GetCharArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jcharArray(array), C.jsize(index), C.jsize(1), &ret)
	return uint16(ret)
}

func (env Env) GetShortArrayElement(array JshortArray, index int) int16 {
	var ret C.jshort
	C.GetShortArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jshortArray(array), C.jsize(index), C.jsize(1), &ret)
	return int16(ret)
}

func (env Env) GetIntArrayElement(array JintArray, index int) int {
	var ret C.jint
	C.GetIntArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jintArray(array), C.jsize(index), C.jsize(1), &ret)
	return int(ret)
}

func (env Env) GetLongArrayElement(array JlongArray, index int) int64 {
	var ret C.jlong
	C.GetLongArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jlongArray(array), C.jsize(index), C.jsize(1), &ret)
	return int64(ret)
}

func (env Env) GetFloatArrayElement(array JfloatArray, index int) float32 {
	var ret C.jfloat
	C.GetFloatArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jfloatArray(array), C.jsize(index), C.jsize(1), &ret)
	return float32(ret)
}

func (env Env) GetDoubleArrayElement(array JdoubleArray, index int) float64 {
	var ret C.jdouble
	C.GetDoubleArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jdoubleArray(array), C.jsize(index), C.jsize(1), &ret)
	return float64(ret)
}

func (env Env) SetBooleanArrayElement(array JbooleanArray, index int, v bool) {
	cv := cbool(v)
	C.SetBooleanArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jbooleanArray(array), C.jsize(index), C.jsize(1), &cv)
}

func (env Env) SetByteArrayElement(array JbyteArray, index int, v byte) {
	cv := C.jbyte(v)
	C.SetByteArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jbyteArray(array), C.jsize(index), C.jsize(1), &cv)
}

func (env Env) SetCharArrayElement(array JcharArray, index int, v uint16) {
	cv := C.jchar(v)
	C.SetCharArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jcharArray(array), C.jsize(index), C.jsize(1), &cv)
}

func (env Env) SetShortArrayElement(array JshortArray, index int, v int16) {
	cv := C.jshort(v)
	C.SetShortArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jshortArray(array), C.jsize(index), C.jsize(1), &cv)
}

func (env Env) SetIntArrayElement(array JintArray, index int, v int) {
	cv := C.jint(v)
	C.SetIntArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jintArray(array), C.jsize(index), C.jsize(1), &cv)
}

func (env Env) SetLongArrayElement(array JlongArray, index int, v int64) {
	cv := C.jlong(v)
	C.SetLongArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jlongArray(array), C.jsize(index), C.jsize(1), &cv)
}

func (env Env) SetFloatArrayElement(array JfloatArray, index int, v float32) {
	cv := C.jfloat(v)
	C.SetFloatArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jfloatArray(array), C.jsize(index), C.jsize(1), &cv)
}

func (env Env) SetDoubleArrayElement(array JdoubleArray, index int, v float64) {
	cv := C.jdouble(v)
	C.SetDoubleArrayRegion((*C.JNIEnv)(unsafe.Pointer(env)), C.jdoubleArray(array), C.jsize(index), C.jsize(1), &cv)
}

`)

	for _, m := range list {
		if skip, err := generateGoFuncCode(m, buf); err == nil {
			if !skip {
				fmt.Fprintln(buf)
			}
		}
	}

	fmt.Fprint(buf, `func DoubleValue(f float64) Jvalue {
	return *(*Jvalue)(unsafe.Pointer(&f))
}

func FloatValue(f float32) Jvalue {
	return Jvalue(*(*uint32)(unsafe.Pointer(&f)))
}

func Int8Value(i int8) Jvalue {
	return Jvalue(*(*uint8)(unsafe.Pointer(&i)))
}

func Int16Value(i int16) Jvalue {
	return Jvalue(*(*uint16)(unsafe.Pointer(&i)))
}

func Int32Value(i int32) Jvalue {
	return Jvalue(*(*uint32)(unsafe.Pointer(&i)))
}

func IntValue(i int) Jvalue {
	return Jvalue(*(*uint)(unsafe.Pointer(&i)))
}

func BooleanValue(b bool) Jvalue {
	return Jvalue(cbool(b))
}

func Bool(b uint8) bool {
	return b != 0
}

func cmem(b []byte) *C.char {
	return (*C.char)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&b))))
}

func cbool(b bool) C.jboolean {
	if b {
		return C.JNI_TRUE
	} else {
		return C.JNI_FALSE
	}
}

func cvals(v []Jvalue) *C.jvalue {
	if len(v) == 0 {
		return nil
	}
	return (*C.jvalue)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&v))))
}

func cBooleanArray(a []bool) *C.jboolean {
	return (*C.jboolean)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&a))))
}

func cByteArray(a []byte) *C.jbyte {
	return (*C.jbyte)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&a))))
}

func cShortArray(a []int16) *C.jshort {
	return (*C.jshort)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&a))))
}

func cCharArray(a []uint16) *C.jchar {
	return (*C.jchar)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&a))))
}

func cIntArray(a []int32) *C.jint {
	return (*C.jint)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&a))))
}

func cLongArray(a []int64) *C.jlong {
	return (*C.jlong)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&a))))
}

func cFloatArray(a []float32) *C.jfloat {
	return (*C.jfloat)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&a))))
}

func cDoubleArray(a []float64) *C.jdouble {
	return (*C.jdouble)(unsafe.Pointer(*(*uintptr)(unsafe.Pointer(&a))))
}
`)

	return buf.String()
}

//
// 处理参数，去掉参数中的 const
// 处理返回值，去掉返回值中的 const
func generateFuncCode(m *method, buf *bytes.Buffer) (skip bool, err error) {
	if m.isVarArgs() || m.isVaList() {
		return false, fmt.Errorf("%s 处理不了不定参数的情况", m)
	}

	if containsInSkipList(m.name, skipList) {
		// 忽略
		return true, nil
	}

	fmt.Fprintf(buf, "// static inline %s %s(%s) {\n",
		m.ret.toC().TypeDesc(), m.name, m.toC().paramList())
	ret := ""
	if m.hasRetVal() {
		ret = "return "
	}

	expr := fmt.Sprintf("(*%s)->%s(%s)",
		m.params[0].idName,
		m.name,
		m.toC().callList())

	expr = m.toC().beforeReturn(expr)

	fmt.Fprintf(buf, "//     %s%s;\n",
		ret,
		expr)
	fmt.Fprint(buf, "// }\n")

	return false, nil
}

func generateGoFuncCode(m *method, buf *bytes.Buffer) (bool, error) {
	if m.isVarArgs() || m.isVaList() {
		return false, fmt.Errorf("%s 处理不了不定参数的情况", m)
	}

	if containsInSkipList(m.name, skipList) || containsInSkipList(m.name, goSkipList) {
		// 忽略
		return true, nil
	}

	fmt.Fprintf(buf, "func (%s %s) %s(%s)%s {\n",
		m.params[0].idName, m.params[0].cType.toGo().TypeDesc(),
		m.name, m.toGo().paramList(), m.goRetVal())

	m.toGo().prepareReturn(buf)

	ret := ""
	if m.hasRetVal() {
		ret = "return "
	}

	expr := fmt.Sprintf("C.%s(%s)",
		m.name,
		m.toGo().callList())
	expr = m.toGo().beforeReturn(expr)

	fmt.Fprintf(buf, "\t%s%s\n",
		ret,
		expr)
	fmt.Fprint(buf, "}\n")

	return false, nil
}
