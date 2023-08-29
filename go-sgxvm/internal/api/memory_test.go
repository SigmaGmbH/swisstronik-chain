package api

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestMakeView(t *testing.T) {
	data := []byte{0xaa, 0xbb, 0x64}
	dataView := MakeView(data)
	require.Equal(t, cbool(false), dataView.is_nil)
	require.Equal(t, cusize(3), dataView.len)

	empty := []byte{}
	emptyView := MakeView(empty)
	require.Equal(t, cbool(false), emptyView.is_nil)
	require.Equal(t, cusize(0), emptyView.len)

	nilView := MakeView(nil)
	require.Equal(t, cbool(true), nilView.is_nil)
}

func TestCreateAndDestroyUnmanagedVector(t *testing.T) {
	// non-empty
	{
		original := []byte{0xaa, 0xbb, 0x64}
		unmanaged := NewUnmanagedVector(original)
		require.Equal(t, cbool(false), unmanaged.is_none)
		require.Equal(t, 3, int(unmanaged.len))
		require.GreaterOrEqual(t, 3, int(unmanaged.cap)) // Rust implementation decides this
		copy := CopyAndDestroyUnmanagedVector(unmanaged)
		require.Equal(t, original, copy)
	}

	// empty
	{
		original := []byte{}
		unmanaged := NewUnmanagedVector(original)
		require.Equal(t, cbool(false), unmanaged.is_none)
		require.Equal(t, 0, int(unmanaged.len))
		require.GreaterOrEqual(t, 0, int(unmanaged.cap)) // Rust implementation decides this
		copy := CopyAndDestroyUnmanagedVector(unmanaged)
		require.Equal(t, original, copy)
	}

	// none
	{
		var original []byte
		unmanaged := NewUnmanagedVector(original)
		require.Equal(t, cbool(true), unmanaged.is_none)
		// We must not make assumtions on the other fields in this case
		copy := CopyAndDestroyUnmanagedVector(unmanaged)
		require.Nil(t, copy)
	}
}

// Like the test above but without `NewUnmanagedVector` calls.
// Since only Rust can actually create them, we only test edge cases here.
//go:nocheckptr
func TestCopyDestroyUnmanagedVector(t *testing.T) {
	{
		// ptr, cap and len broken. Do not access those values when is_none is true
		invalid_ptr := unsafe.Pointer(uintptr(42))
		uv := ConstructUnmanagedVector(cbool(true), cu8_ptr(invalid_ptr), cusize(0xBB), cusize(0xAA))
		copy := CopyAndDestroyUnmanagedVector(uv)
		require.Nil(t, copy)
	}
	{
		// Capacity is 0, so no allocation happened. Do not access the pointer.
		invalid_ptr := unsafe.Pointer(uintptr(42))
		uv := ConstructUnmanagedVector(cbool(false), cu8_ptr(invalid_ptr), cusize(0), cusize(0))
		copy := CopyAndDestroyUnmanagedVector(uv)
		require.Equal(t, []byte{}, copy)
	}
}
