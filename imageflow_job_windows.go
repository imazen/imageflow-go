// +build windows

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

// Job to perform a task in imageflow
type Job struct {
	inner *C.struct_imageflow_context
	err   bool
}

// CheckError is used to check if the context has error or not
func (job Job) CheckError() bool {
	if job.err {
		return true
	}
	val := C.imageflow_context_has_error(job.inner)
	if bool(val) {
		C.imageflow_context_destroy(job.inner)
		job.inner = nil
		job.err = true
	}
	return bool(val)
}

// AddInput add input to context
func (job *Job) AddInput(id uint, byt []byte) error {
	if job.CheckError() {
		return errors.New("Error in context")
	}
	result := C.imageflow_context_add_input_buffer(job.inner, C.int(id), (*C.uchar)(C.CBytes(byt)), C.ulonglong(len(byt)), C.imageflow_lifetime_lifetime_outlives_function_call)
	if !bool(result) {
		return errors.New("Error adding input buffer")
	}
	return nil
}

// AddOutput add output to context
func (job *Job) AddOutput(id uint) error {
	result := C.imageflow_context_add_output_buffer(job.inner, C.int(id))

	if !bool(result) {
		return errors.New("Error in adding outputs")
	}

	return nil
}

// Message execute a command
func (job *Job) Message(message []byte) error {
	C.imageflow_context_send_json(job.inner, C.CString("v1/execute"), (*C.uchar)(C.CBytes(message)), C.ulonglong(len(message)))
	if job.CheckError() {
		return errors.New("Error in context")
	}
	return nil
}

// New Create a context
func New() Job {
	v := C.imageflow_context_create(3, 0)
	return Job{inner: v}
}

// GetOutput from the context
func (job *Job) GetOutput(id uint) ([]byte, error) {
	ptr := (*C.uchar)(C.malloc(C.size_t(unsafe.Sizeof(uintptr(0)))))
	l := 0
	le := (*C.ulonglong)(unsafe.Pointer(&l))
	result := C.imageflow_context_get_output_buffer_by_id(job.inner, C.int(id), (&ptr), le)

	if !bool(result) {
		return nil, errors.New("Error in adding outputs")
	}
	return C.GoBytes((unsafe.Pointer)(ptr), C.int(l)), nil
}

// ReadError from context
func (job *Job) ReadError() error {
	l := 0
	le := (*C.ulong)(unsafe.Pointer(&l))
	byt := make([]byte, 512)
	for !bool(C.imageflow_context_error_write_to_buffer(job.inner, (*C.char)(unsafe.Pointer(&byt[0])), C.ulonglong(len(byt)), le)) {
		byt = make([]byte, len(byt)*2)
	}
	return errors.New(string(byt[0 : l-1]))
}
