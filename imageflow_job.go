// +build !windows

package imageflow

/*
#cgo LDFLAGS: -L./ -limageflow
#include "imageflow.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"unsafe"
)

// job to perform a task in imageflow
type job struct {
	inner  *C.struct_imageflow_context
	allocs []unsafe.Pointer
	err    bool
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
func (job *job) AddInput(id uint, b []byte) error {
	if job.CheckError() {
		return job.ReadError()
	}

	cb := C.CBytes(b)
	job.allocs = append(job.allocs, cb)

	result := C.imageflow_context_add_input_buffer(job.inner, C.int(id),
		(*C.uchar)(cb), C.ulong(len(b)), C.imageflow_lifetime_lifetime_outlives_function_call)

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
	jsize := C.ulong(len(message))

	cs := C.CString("v1/execute")
	cs_ptr := unsafe.Pointer(cs)

	cb := C.CBytes(message)
	cb_ptr := (*C.uchar)(cb)

	job.allocs = append(job.allocs, cb, cs_ptr)

	C.imageflow_context_send_json(job.inner, cs, cb_ptr, jsize)
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

// Frees the context and C allocations.
func (j *job) CleanUp() {
	C.imageflow_context_destroy(j.inner)
	for _, alloc := range j.allocs {
		C.free(alloc)
	}
}

// GetOutput from the context
func (job *job) GetOutput(id uint) ([]byte, error) {
	if job.CheckError() {
		return nil, job.ReadError()
	}

	size := C.size_t(unsafe.Sizeof(uintptr(0)))
	cb := C.malloc(size) // cbytes
	cb_ptr := (*C.uchar)(cb)
	defer C.free(cb)

	length := 0
	length_ptr := (*C.ulong)(unsafe.Pointer(&length))

	result := C.imageflow_context_get_output_buffer_by_id(
		job.inner, C.int(id),
		(&cb_ptr), length_ptr)

	if !bool(result) {
		return nil, job.ReadError()
	}
	return C.GoBytes((unsafe.Pointer)(cb_ptr), C.int(length)), nil
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
