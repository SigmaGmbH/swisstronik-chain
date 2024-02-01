//go:build linux && !muslc && arm64 && !sys_sgx_wrapper && !nosgx

package api

// #cgo LDFLAGS: -Wl,-rpath,${SRCDIR} -L${SRCDIR} -lsgx_wrapper.aarch64
import "C"
