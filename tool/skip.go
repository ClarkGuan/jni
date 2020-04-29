package tool

var skipList = []string{
	// 类操作
	"DefineClass",

	// 字符串操作
	"NewStringUTF",
	"GetStringChars",
	"ReleaseStringChars",
	"GetStringCritical",
	"ReleaseStringCritical",
	"GetStringRegion",
	"GetStringUTFChars",
	"ReleaseStringUTFChars",

	// 数组操作
	"GetBooleanArrayElements",
	"GetByteArrayElements",
	"GetCharArrayElements",
	"GetShortArrayElements",
	"GetIntArrayElements",
	"GetLongArrayElements",
	"GetFloatArrayElements",
	"GetDoubleArrayElements",
	"ReleaseBooleanArrayElements",
	"ReleaseByteArrayElements",
	"ReleaseCharArrayElements",
	"ReleaseShortArrayElements",
	"ReleaseIntArrayElements",
	"ReleaseLongArrayElements",
	"ReleaseFloatArrayElements",
	"ReleaseDoubleArrayElements",

	// 注册
	"RegisterNatives",
	"UnregisterNatives",

	// 引用操作
	"GetObjectRefType",
}

var goSkipList = []string{
	// 字符串操作
	"NewString",
	"NewStringUTF",
	"GetStringUTFLength",
	"GetStringUTFRegion",

	// NIO
	"NewDirectByteBuffer",
	"GetDirectBufferAddress",
	"GetDirectBufferCapacity",
}

func containsInSkipList(s string, list []string) bool {
	for _, ss := range list {
		if s == ss {
			return true
		}
	}

	return false
}
