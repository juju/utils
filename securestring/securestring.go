// Copyright 2015 Canonical Ltd.
// Copyright 2015 Cloudbase Solutions SRL.
// Licensed under the AGPLv3, see LICENCE file for details.
//
// +build windows

package securestring

import (
	"encoding/hex"
	"strings"
	"syscall"
	"unsafe"

	"github.com/juju/errors"
)

var (
	cryptdll  = syscall.NewLazyDLL("Crypt32.dll")
	kerneldll = syscall.NewLazyDLL("Kernel32.dll")

	// Syntax:
	// procProtectData.Call(inputBlob *blob, dataDescription *string,
	//		optionalEntropy *blob, freeWorkingSpace *void, proptStruct
	//		*CRYPTPROTECT_PROMPTSTRUCT, dwflags uint, outputBlob *blob)
	//
	// Parameters:
	// In the calls made by the ConvertFrom-SecureString commandlet;
	// dataDescription, optionalEntropy, freeWorkingSpace and proptStruct
	// are set to their respective zero values.
	//
	// inputBlob contains the actual input, outputBlob is set to its zero
	// value, and dwflags is set to the default of 1.
	//
	// Return value:
	// A C-boolean; 1(true) if it succeeds, 0(false) if it fails
	procProtectData = cryptdll.NewProc("CryptProtectData")

	// Syntax:
	// procUnprotectData.Call(inputBlob *blob, dataDescription *string,
	//		optionalEntropy *blob, freeWorkingSpace *void, proptStruct
	//		*CRYPTPROTECT_PROMPTSTRUCT, dwflags uint, outputBlob *blob)
	//
	// Parameters:
	// In our case; dataDescription, optionalEntropy, freeWorkingSpace and
	// proptStruct are set to their respective zero values.
	// inputBlob contains the actual input, outputBlob is set to its zero
	// value, and dwflags is set to the default of 1.
	//
	// Return value:
	// A C-boolean; 1(true) if it succeeds, 0(false) if it fails.
	procUnprotectData = cryptdll.NewProc("CryptUnprotectData")

	// Syntax:
	// procLocalFree.Call(ptr *uint)
	//
	// Parameter:
	// An unsafe pointer of any type, for our purposes a *uint.
	//
	// Return value:
	// Pointer value. nil if it succeeds, ptr if it fails.
	procLocalFree = kerneldll.NewProc("LocalFree")
)

// blob is the struct type we shall be making the syscalls on. It contains a
// pointer to the start of the actual data and its respective length in bytes.
type blob struct {
	length uint32
	data   *byte
}

// getData is a helper method which fetches all the data pointed to by blob.data.
func (b *blob) getData() []byte {
	var fetched = make([]byte, b.length)

	// The built-in will copy the proper amount of data pointed to by blob.data
	// and put it in the new variable.
	// 1 << 30 is the largest possible slice size; it's pretty overkill but it
	// ensures we can read as most of very large data as physically possible.
	copy(fetched, (*[1 << 30]byte)(unsafe.Pointer(b.data))[:])

	return fetched
}

// Encrypt encrypts a provided string as input into a hexadecimal string.
// The output corresponds to the output of ConvertFrom-SecureString:
func Encrypt(input string) (string, error) {
	data := []byte(input)

	// For some reason; the cmdlet's calls automatically encrypts the bytes
	// with interwoven null characters, so we must account for this as follows:
	nulled := []byte{}
	for _, b := range data {
		nulled = append(nulled, b)
		nulled = append(nulled, 0)
	}

	inputBlob := blob{uint32(len(nulled)), &nulled[0]}
	entropyBlob := blob{}
	outputBlob := blob{}
	dwflags := 1

	res, _, err := procProtectData.Call(uintptr(unsafe.Pointer(&inputBlob)),
		uintptr(0), uintptr(unsafe.Pointer(&entropyBlob)), uintptr(0),
		uintptr(0), uintptr(uint(dwflags)),
		uintptr(unsafe.Pointer(&outputBlob)))
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outputBlob.data)))

	// check if result is 0 (C's false).
	if res == 0 {
		return "", errors.Trace(err)
	}

	output := outputBlob.getData()

	// The result is a slice of bytes, which we must encode into hexa
	// to match ConvertFrom-SecureString's output before returning it.
	return hex.EncodeToString(output), nil
}

// Decrypt converts the output from a call to ConvertFrom-SecureString
// back to the original input string and returns it.
func Decrypt(input string) (string, error) {
	// Trim spaces preemptively here.
	trimmed := strings.TrimSpace(input)

	// First we decode the hexadecimal string into a raw slice of bytes.
	data, err := hex.DecodeString(trimmed)
	if err != nil {
		return "", err
	}

	inputBlob := blob{uint32(len(data)), &data[0]}
	entropyBlob := blob{}
	outputBlob := blob{}
	dwflags := 1

	res, _, err := procUnprotectData.Call(uintptr(unsafe.Pointer(&inputBlob)),
		uintptr(0), uintptr(unsafe.Pointer(&entropyBlob)), uintptr(0),
		uintptr(0), uintptr(uint(dwflags)),
		uintptr(unsafe.Pointer(&outputBlob)))
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outputBlob.data)))

	// Check is result is 0 (C's false).
	if res == 0 {
		return "", errors.Trace(err)
	}

	output := outputBlob.getData()

	// As mentioned, the cmdlets infer working with data with interwoven
	// null characters, for which we must account for by removing them now:
	clean := []byte{}
	for _, b := range output {
		if b != 0 {
			clean = append(clean, b)
		}
	}

	return string(clean), nil
}
