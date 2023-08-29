//go:build darwin && !sys_sgx_wrapper

package api

// #cgo LDFLAGS: -Wl,-rpath,${SRCDIR} -L${SRCDIR} -lsgx_wrapper
import "C"
