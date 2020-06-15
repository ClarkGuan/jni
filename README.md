# jni
方便集成 JNI 代码的 Go 库。

完整支持所有 JNI 功能函数：https://docs.oracle.com/javase/8/docs/technotes/guides/jni/spec/functions.html

配合 https://github.com/ClarkGuan/gojni 使用，方便 Java 与 Go 代码通信。

### 安装

安装前准备：

- 安装 JDK，并设置 JAVA_HOME 环境变量
- 安装 includejni ( go get -u github.com/ClarkGuan/includejni )
- 运行下面命令

```
git clone https://github.com/ClarkGuan/jni
cd jni
./build.sh
```

### 使用举例

我们假设您已经安装 JDK、Go 以及 C 编译器，这里以 MacOS 平台为例：

* 目录结构

```
.
└── src
    └── java
        └── com
            └── demo
                └── Main.java
```

* 编写 Java 源代码

```java
package com.demo;

public class Main {
    static {
        System.loadLibrary("hello");
    }

    public static void main(String[] args) {
        nativeHello();
        System.out.println(stringFromJNI());
    }

    private static native void nativeHello();

    private static native String stringFromJNI();
}
```

* 运行 gojni 工具（ 安装方法见：https://github.com/ClarkGuan/gojni ）

```
gojni src/java/com/demo/Main.java
```

* 此时在目录中生成文件 libs.c 和 libs.go，我们修改 libs.go 文件内容如下：

```go
package main

//
// #include <stdlib.h>
// #include <stddef.h>
// #include <stdint.h>
import "C"
import (
	"fmt"

	"github.com/ClarkGuan/jni"
)

//export jniOnLoad
func jniOnLoad(vm uintptr) {
	fmt.Println("JNI_OnLoad")
}

//export jniOnUnload
func jniOnUnload(vm uintptr) {
	fmt.Println("JNI_OnUnload")
}

//export jni_com_demo_nativeHello1
func jni_com_demo_nativeHello1(env uintptr, clazz uintptr) {
	fmt.Println("native hello form golang")
}

//export jni_com_demo_stringFromJNI2
func jni_com_demo_stringFromJNI2(env uintptr, clazz uintptr) uintptr {
	return jni.Env(env).NewString("This is string from Golang code!!!")
}

```

* 运行 go build 生成 Mac 上运行的动态库文件：

```bash
go build -buildmode=c-shared -ldflags="-w -s" -v -x -o libhello.dylib
```

注意：需要将 JNI 的头文件引入，否则 C 编译器可能找不到。使用环境变量 `CGO_CFLAGS`:

```bash
CGO_CFLAGS="-I$JAVA_HOME/include -I$JAVA_HOME/include/darwin" go build -buildmode=c-shared -ldflags="-w -s" -v -x -o libhello.dylib
```

或者使用前面下载的 includejni 工具：

```bash
includejni go build -buildmode=c-shared -ldflags="-w -s" -v -x -o libhello.dylib
```

* 编译 Java 源码

```
javac src/java/com/demo/Main.java
```

* 将 libhello.dylib 放入 Java 虚拟机可以找到的位置（通过 Java 环境变量 java.library.path 指定）

* 运行 Java 程序

```
java -cp src com.demo.Main
```

* 运行结果

```
JNI_OnLoad
native hello form golang
This is string from Golang code!!!
```
