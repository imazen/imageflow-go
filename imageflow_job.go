// +build !windows

package imageflow

/*
#cgo LDFLAGS: -L./ -limageflow
#include "imageflow.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

// job to perform a task in imageflow
type job struct {
	inner *C.struct_imageflow_context
	err   bool
}

// CheckError is used to check if the context has error or not
func (job job) CheckError() bool {
	if job.err {
		return true
	}
	if job.inner == nil {
		return true
	}
	val := C.imageflow_context_has_error(job.inner)
	if val == C.bool(true) {
		job.err = true
		return true
	}
	return bool(val)
}

// AddInput add input to context
func (job *job) AddInput(id uint, byt []byte) error {
	if job.CheckError() {
		return job.ReadError()
	}
	result := C.imageflow_context_add_input_buffer(job.inner, C.int(id), (*C.uchar)(C.CBytes(byt)), C.ulong(len(byt)), C.imageflow_lifetime_lifetime_outlives_function_call)
	if !bool(result) {
		return job.ReadError()
	}
	return nil
}

// AddOutput add output to context
func (job *job) AddOutput(id uint) error {
	if job.CheckError() {
		return job.ReadError()
	}
	result := C.imageflow_context_add_output_buffer(job.inner, C.int(id))

	if !bool(result) {
		return job.ReadError()
	}

	return nil
}

// Message execute a command
func (job *job) Message(message []byte) error {
	if job.CheckError() {
		return job.ReadError()
	}
	C.imageflow_context_send_json(job.inner, C.CString("v1/execute"), (*C.uchar)(C.CBytes(message)), C.ulong(len(message)))
	if job.CheckError() {
		return job.ReadError()
	}
	return nil
}

// New Create a context
func newJob() job {
	v := C.imageflow_context_create(3, 0)
	return job{inner: v}
}

// GetOutput from the context
func (job *job) GetOutput(id uint) ([]byte, error) {
	if job.CheckError() {
		return nil, job.ReadError()
	}

	ptr := (*C.uchar)(C.malloc(C.size_t(unsafe.Sizeof(uintptr(0)))))
	l := 0
	le := (*C.ulong)(unsafe.Pointer(&l))
	result := C.imageflow_context_get_output_buffer_by_id(job.inner, C.int(id), (&ptr), le)

	if !bool(result) {
		return nil, job.ReadError()
	}
	return C.GoBytes((unsafe.Pointer)(ptr), C.int(l)), nil
}

// ReadError from the context
func (job *job) ReadError() error {
	l := 0
	le := (*C.ulong)(unsafe.Pointer(&l))
	byt := make([]byte, 512)
	for !bool(C.imageflow_context_error_write_to_buffer(job.inner, (*C.char)(unsafe.Pointer(&byt[0])), C.ulong(len(byt)), le)) {
		byt = make([]byte, len(byt)*2)
	}
	return errors.New(string(byt[0 : l-1]))
}
