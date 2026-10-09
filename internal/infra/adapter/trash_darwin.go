//go:build darwin && cgo

package adapter

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation
#import <Foundation/Foundation.h>
#include <stdlib.h>
#include <string.h>

static char *xenonMoveToTrash(const char *path) {
	@autoreleasepool {
		NSURL *url = [NSURL fileURLWithPath:[NSString stringWithUTF8String:path]];
		NSError *error = nil;
		if ([[NSFileManager defaultManager] trashItemAtURL:url resultingItemURL:nil error:&error]) {
			return NULL;
		}
		return strdup([[error localizedDescription] UTF8String]);
	}
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

func moveToTrash(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	failure := C.xenonMoveToTrash(cPath)
	if failure == nil {
		return nil
	}
	defer C.free(unsafe.Pointer(failure))
	return errors.New(C.GoString(failure))
}
