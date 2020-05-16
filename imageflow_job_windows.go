// +build windows

package imageflow

/*
#cgo LDFLAGS: -L./ -limageflow
#include "imageflow.h"
*/
import "C"
import (
	"unsafe"
)

// Job to perform a task in imageflow
type Job struct {
	inner *C.struct_imageflow_context
}

// AddInput add input to context
func (job *Job) AddInput(id uint, byt []byte) {
	C.imageflow_context_add_input_buffer(job.inner, C.int(id), (*C.uchar)(C.CBytes(byt)), C.ulonglong(len(byt)), C.imageflow_lifetime_lifetime_outlives_function_call)
}

// AddOutput add output to context
func (job *Job) AddOutput(id uint) {
	C.imageflow_context_add_output_buffer(job.inner, C.int(id))

}

// Message execute a command
func (job *Job) Message(message []byte) {
	C.imageflow_context_send_json(job.inner, C.CString("v1/execute"), (*C.uchar)(C.CBytes(message)), C.ulonglong(len(message)))
}

// New Create a context
func New() Job {
	v := C.imageflow_context_create(3, 0)
	return Job{inner: v}
}

// GetOutput from the context
func (job *Job) GetOutput(id uint) []byte {
	ptr := (*C.uchar)(C.malloc(C.size_t(unsafe.Sizeof(uintptr(0)))))
	l := 0
	le := (*C.ulonglong)(unsafe.Pointer(&l))
	C.imageflow_context_get_output_buffer_by_id(job.inner, C.int(id), (&ptr), le)
	return C.GoBytes((unsafe.Pointer)(ptr), C.int(l))
}
