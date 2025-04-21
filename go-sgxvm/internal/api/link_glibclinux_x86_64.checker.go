//go:build linux && !muslc && amd64 && !sys_sgx_wrapper && !nosgx && checker

package api

// #cgo LDFLAGS: -Wl,-rpath,${SRCDIR} -L${SRCDIR} -lsgx_checker_wrapper_v1.0.1.x86_64
import "C"
