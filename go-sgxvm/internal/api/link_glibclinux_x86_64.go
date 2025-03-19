//go:build linux && !muslc && amd64 && !sys_sgx_wrapper && !nosgx && !attestationServer && !checker

package api

// #cgo LDFLAGS: -Wl,-rpath,${SRCDIR} -L${SRCDIR} -lsgx_wrapper_v1.0.8.x86_64
import "C"
