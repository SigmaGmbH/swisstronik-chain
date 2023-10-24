//go:build sys_sgx_wrapper && !nosgx

package api

// #cgo LDFLAGS: -lsgx_wrapper
import "C"
