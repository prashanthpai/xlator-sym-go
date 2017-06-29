package main

/*
#cgo LDFLAGS: -ldl

#include "options.h"
#include <dlfcn.h>
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

	handle := C.dlopen(C.CString(xlator), C.RTLD_LAZY)
	if handle == nil {
		return fmt.Errorf("dlopen(%s) failed; dlerror = %s",
			xlator, C.GoString((*C.char)(C.dlerror())))
	}
	defer func(h unsafe.Pointer, x string) {
		ret := int(C.dlclose(h))
		if ret != 0 {
			fmt.Printf("dlclose(%v) failed for xlator %s\n", h, x)
		}
	}(handle, xlator)

	p := C.dlsym(handle, C.CString("options"))
	if p == nil {
		// not an xlator
		return nil
	}

	tempSlice := (*[100]C.volume_option_t)(p)

	count := 0
	for _, o := range tempSlice {
		if o.key[0] == nil {
			break
		}
		count += 1
	}

	if count == 0 {
		// xlator has no options
		return nil
	}

	xlatorOptions := tempSlice[:count]
	fmt.Printf("\n%s\n", xlator)
	for _, option := range xlatorOptions {
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

	if f, err := os.Stat(xlatorDirs); err != nil || !f.Mode().IsDir() {
		fmt.Printf("Invalid xlator directory: %s\n", xlatorDirs)
		os.Exit(-1)
	}

	if err := filepath.Walk(xlatorDirs, findXlators); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
