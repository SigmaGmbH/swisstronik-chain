//go:build sys_sgx_wrapper

package api

// #cgo LDFLAGS: -lsgx_wrapper
import "C"
