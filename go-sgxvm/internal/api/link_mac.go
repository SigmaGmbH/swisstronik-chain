//go:build darwin && !sys_sgx_wrapper && !nosgx

package api

// #cgo LDFLAGS: -Wl,-rpath,${SRCDIR} -L${SRCDIR} -lsgx_wrapper_v1.0.2
import "C"
