//go:build linux && !muslc && amd64 && !sys_sgx_wrapper && !nosgx && !attestationServer

package api

// #cgo LDFLAGS: -Wl,-rpath,${SRCDIR} -L${SRCDIR} -lsgx_wrapper_v1.0.5.x86_64
import "C"
