package main

/*
#cgo LDFLAGS: -ldl

#include <stdlib.h>    // free()
#include <dlfcn.h>     // dlopen(), dlsym(), dlclose()
#include "options.h"   // volume_option_t
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

func loadXlatorOptions(xlator string) error {

	csXlator := C.CString(xlator)
	defer C.free(unsafe.Pointer(csXlator))

	handle := C.dlopen(csXlator, C.RTLD_LAZY|C.RTLD_LOCAL)
	if handle == nil {
		return fmt.Errorf("dlopen(%s) failed; dlerror = %s",
			xlator, C.GoString((*C.char)(C.dlerror())))
	}
	defer C.dlclose(handle)

	csSym := C.CString("options")
	defer C.free(unsafe.Pointer(csSym))

	p := C.dlsym(handle, csSym)
	if p == nil {
		return nil
	}

	xlatorOptions := (*[100]C.volume_option_t)(p)
	for i, option := range xlatorOptions {
		if option.key[0] == nil {
			break
		}
		if i == 0 {
			fmt.Printf("\n%s\n", xlator)
		}
		fmt.Printf("%s\n", C.GoString(option.key[0]))
	}

	return nil
}

func findXlators(path string, f os.FileInfo, err error) error {
	if strings.HasSuffix(path, ".so") {
		if err := loadXlatorOptions(path); err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func main() {

	if len(os.Args) != 2 {
		msg := fmt.Sprintf("Usage:\n %s <path-to-xlator-dir>\n"+
			"Example:\n %s /usr/local/lib/glusterfs/3.12dev/xlator",
			os.Args[0], os.Args[0])
		fmt.Println(msg)
		os.Exit(-1)
	}

	xlatorDirs := os.Args[1]
	if err := filepath.Walk(xlatorDirs, findXlators); err != nil {
		fmt.Println(err)
	}
}
