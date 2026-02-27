package imageflow

/*
#cgo linux LDFLAGS: -L./ -limageflow -lm -ldl -lpthread
#cgo darwin LDFLAGS: -L./ -limageflow
#cgo windows LDFLAGS: -L./ -limageflow
#include "imageflow.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// job to perform a task in imageflow
type job struct {
	inner *C.struct_imageflow_context
	err   bool
}

// CheckError is used to check if the context has error or not
func (job *job) CheckError() bool {
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
func (job *job) AddInput(id uint, b []byte) error {
	if job.CheckError() {
		return job.ReadError()
	}

	cb := C.CBytes(b)
	defer C.free(cb)

	result := C.imageflow_context_add_input_buffer(job.inner, C.int(id),
		(*C.uchar)(cb), C.size_t(len(b)), C.imageflow_lifetime_lifetime_outlives_function_call)

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

	cs := C.CString("v1/execute")
	defer C.free(unsafe.Pointer(cs))

	cb := C.CBytes(message)
	defer C.free(cb)

	C.imageflow_context_send_json(job.inner, cs, (*C.uchar)(cb), C.size_t(len(message)))
	if job.CheckError() {
		return job.ReadError()
	}
	return nil
}

// newJob creates a context after verifying ABI compatibility
func newJob() (*job, error) {
	if !bool(C.imageflow_abi_compatible(C.IMAGEFLOW_ABI_VER_MAJOR, C.IMAGEFLOW_ABI_VER_MINOR)) {
		return nil, fmt.Errorf("imageflow ABI mismatch: header wants %d.%d",
			C.IMAGEFLOW_ABI_VER_MAJOR, C.IMAGEFLOW_ABI_VER_MINOR)
	}
	v := C.imageflow_context_create(C.IMAGEFLOW_ABI_VER_MAJOR, C.IMAGEFLOW_ABI_VER_MINOR)
	if v == nil {
		return nil, errors.New("imageflow_context_create returned nil")
	}
	return &job{inner: v}, nil
}

// CleanUp frees the context.
func (j *job) CleanUp() {
	if j.inner != nil {
		C.imageflow_context_destroy(j.inner)
		j.inner = nil
	}
}

// GetOutput from the context
func (job *job) GetOutput(id uint) ([]byte, error) {
	if job.CheckError() {
		return nil, job.ReadError()
	}

	var bufPtr *C.uint8_t
	var bufLen C.size_t

	result := C.imageflow_context_get_output_buffer_by_id(
		job.inner, C.int(id),
		&bufPtr, &bufLen)

	if !bool(result) {
		return nil, job.ReadError()
	}
	return C.GoBytes(unsafe.Pointer(bufPtr), C.int(bufLen)), nil
}

// ReadError from the context
func (job *job) ReadError() error {
	var written C.size_t
	byt := make([]byte, 512)
	for !bool(C.imageflow_context_error_write_to_buffer(job.inner, (*C.char)(unsafe.Pointer(&byt[0])), C.size_t(len(byt)), &written)) {
		byt = make([]byte, len(byt)*2)
	}
	return errors.New(string(byt[0:written]))
}
