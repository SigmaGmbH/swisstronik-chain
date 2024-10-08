//go:build linux && muslc && !sys_sgx_wrapper

package api

// #cgo LDFLAGS: -Wl,-rpath,${SRCDIR} -L${SRCDIR} -lsgx_wrapper_v1.0.6_muslc
import "C"
