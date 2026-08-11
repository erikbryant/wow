package credentials

// Forked from https://github.com/keybase/go-keychain

import "C"
import (
	"fmt"
	"reflect"
	"time"
	"unsafe"
)

// See https://developer.apple.com/library/ios/documentation/Security/Reference/keychainservices/index.html for the APIs used below.
// Also see https://developer.apple.com/library/ios/documentation/Security/Conceptual/keychainServConcepts/01introduction/introduction.html .

/*
   #cgo LDFLAGS: -framework Security

   #include <Security/Security.h>

CFDictionaryRef CFDictionaryCreateSafe2(CFAllocatorRef allocator, const uintptr_t *keys, const uintptr_t *values, CFIndex numValues, const CFDictionaryKeyCallBacks *keyCallBacks, const CFDictionaryValueCallBacks *valueCallBacks) {
     return CFDictionaryCreate(allocator, (const void **)keys, (const void **)values, numValues, keyCallBacks, valueCallBacks);
}
*/
import "C"

// Error defines keychain errors
type Error int

var (
	// ErrorUnimplemented corresponds to errSecUnimplemented result code
	ErrorUnimplemented = Error(C.errSecUnimplemented)
	// ErrorParam corresponds to errSecParam result code
	ErrorParam = Error(C.errSecParam)
	// ErrorAllocate corresponds to errSecAllocate result code
	ErrorAllocate = Error(C.errSecAllocate)
	// ErrorNotAvailable corresponds to errSecNotAvailable result code
	ErrorNotAvailable = Error(C.errSecNotAvailable)
	// ErrorAuthFailed corresponds to errSecAuthFailed result code
	ErrorAuthFailed = Error(C.errSecAuthFailed)
	// ErrorDuplicateItem corresponds to errSecDuplicateItem result code
	ErrorDuplicateItem = Error(C.errSecDuplicateItem)
	// ErrorItemNotFound corresponds to errSecItemNotFound result code
	ErrorItemNotFound = Error(C.errSecItemNotFound)
	// ErrorInteractionNotAllowed corresponds to errSecInteractionNotAllowed result code
	ErrorInteractionNotAllowed = Error(C.errSecInteractionNotAllowed)
	// ErrorDecode corresponds to errSecDecode result code
	ErrorDecode = Error(C.errSecDecode)
	// ErrorNoSuchKeychain corresponds to errSecNoSuchKeychain result code
	ErrorNoSuchKeychain = Error(C.errSecNoSuchKeychain)
	// ErrorNoAccessForItem corresponds to errSecNoAccessForItem result code
	ErrorNoAccessForItem = Error(C.errSecNoAccessForItem)
	// ErrorReadOnly corresponds to errSecReadOnly result code
	ErrorReadOnly = Error(C.errSecReadOnly)
	// ErrorInvalidKeychain corresponds to errSecInvalidKeychain result code
	ErrorInvalidKeychain = Error(C.errSecInvalidKeychain)
	// ErrorDuplicateKeyChain corresponds to errSecDuplicateKeychain result code
	ErrorDuplicateKeyChain = Error(C.errSecDuplicateKeychain)
	// ErrorWrongVersion corresponds to errSecWrongSecVersion result code
	ErrorWrongVersion = Error(C.errSecWrongSecVersion)
	// ErrorReadonlyAttribute corresponds to errSecReadOnlyAttr result code
	ErrorReadonlyAttribute = Error(C.errSecReadOnlyAttr)
	// ErrorInvalidSearchRef corresponds to errSecInvalidSearchRef result code
	ErrorInvalidSearchRef = Error(C.errSecInvalidSearchRef)
	// ErrorInvalidItemRef corresponds to errSecInvalidItemRef result code
	ErrorInvalidItemRef = Error(C.errSecInvalidItemRef)
	// ErrorDataNotAvailable corresponds to errSecDataNotAvailable result code
	ErrorDataNotAvailable = Error(C.errSecDataNotAvailable)
	// ErrorDataNotModifiable corresponds to errSecDataNotModifiable result code
	ErrorDataNotModifiable = Error(C.errSecDataNotModifiable)
	// ErrorInvalidOwnerEdit corresponds to errSecInvalidOwnerEdit result code
	ErrorInvalidOwnerEdit = Error(C.errSecInvalidOwnerEdit)
	// ErrorUserCanceled corresponds to errSecUserCanceled result code
	ErrorUserCanceled = Error(C.errSecUserCanceled)
)

func (k Error) Error() (msg string) {
	// SecCopyErrorMessageString is only available on OSX, so derive manually.
	// Messages derived from `$ security error $errcode`.
	switch k {
	case ErrorUnimplemented:
		msg = "Function or operation not implemented."
	case ErrorParam:
		msg = "One or more parameters passed to the function were not valid."
	case ErrorAllocate:
		msg = "Failed to allocate memory."
	case ErrorNotAvailable:
		msg = "No keychain is available. You may need to restart your computer."
	case ErrorAuthFailed:
		msg = "The user name or passphrase you entered is not correct."
	case ErrorDuplicateItem:
		msg = "The specified item already exists in the keychain."
	case ErrorItemNotFound:
		msg = "The specified item could not be found in the keychain."
	case ErrorInteractionNotAllowed:
		msg = "User interaction is not allowed."
	case ErrorDecode:
		msg = "Unable to decode the provided data."
	case ErrorNoSuchKeychain:
		msg = "The specified keychain could not be found."
	case ErrorNoAccessForItem:
		msg = "The specified item has no access control."
	case ErrorReadOnly:
		msg = "Read-only error."
	case ErrorReadonlyAttribute:
		msg = "The attribute is read-only."
	case ErrorInvalidKeychain:
		msg = "The keychain is not valid."
	case ErrorDuplicateKeyChain:
		msg = "A keychain with the same name already exists."
	case ErrorWrongVersion:
		msg = "The version is incorrect."
	case ErrorInvalidItemRef:
		msg = "The item reference is invalid."
	case ErrorInvalidSearchRef:
		msg = "The search reference is invalid."
	case ErrorDataNotAvailable:
		msg = "The data is not available."
	case ErrorDataNotModifiable:
		msg = "The data is not modifiable."
	case ErrorInvalidOwnerEdit:
		msg = "An invalid attempt to change the owner of an item."
	case ErrorUserCanceled:
		msg = "User canceled the operation."
	default:
		msg = "Keychain Error."
	}
	return fmt.Sprintf("%s (%d)", msg, k)
}

func checkError(errCode C.OSStatus) error {
	if errCode == C.errSecSuccess {
		return nil
	}
	return Error(errCode)
}

// kcAdd adds an Item to a Keychain
func kcAdd(service, name, value string) error {
	k := make(map[string]interface{})
	k["svce"] = service
	k["class"] = C.CFTypeRef(C.kSecClassGenericPassword)
	k["acct"] = name
	k["sync"] = C.CFTypeRef(C.kCFBooleanFalse)
	k["v_Data"] = value

	cfDict, err := convertMapToCFDictionary(k)
	if err != nil {
		return err
	}
	defer release(C.CFTypeRef(cfDict))

	errCode := C.SecItemAdd(cfDict, nil)
	return checkError(errCode)
}

// queryResult stores all possible results from queries.
// Not all fields are applicable all the time. Results depend on query.
type queryResult struct {
	// For generic application items
	Service string

	// For internet password items
	Server             string
	Protocol           string
	AuthenticationType string
	Port               int32
	Path               string

	Account          string
	AccessGroup      string
	Label            string
	Description      string
	Comment          string
	Data             []byte
	CreationDate     time.Time
	ModificationDate time.Time
}

// queryItemRef returns query result as CFTypeRef. You must release it when you are done.
func queryItemRef(attr map[string]interface{}) (C.CFTypeRef, error) {
	cfDict, err := convertMapToCFDictionary(attr)
	if err != nil {
		return 0, err
	}
	defer release(C.CFTypeRef(cfDict))

	var resultsRef C.CFTypeRef
	errCode := C.SecItemCopyMatching(cfDict, &resultsRef) //nolint
	if Error(errCode) == ErrorItemNotFound {
		return 0, nil
	}
	err = checkError(errCode)
	if err != nil {
		return 0, err
	}
	return resultsRef, nil
}

// kcQuery returns a single query result
func kcQuery(service, name string) (queryResult, error) {
	k := make(map[string]interface{})
	k["svce"] = service
	k["class"] = C.CFTypeRef(C.kSecClassGenericPassword)
	k["acct"] = name
	k["m_Limit"] = C.CFTypeRef(C.kSecMatchLimitOne)
	k["r_Data"] = true

	resultsRef, err := queryItemRef(k)
	if err != nil {
		return queryResult{}, err
	}
	if resultsRef == 0 {
		return queryResult{}, nil
	}
	defer release(resultsRef)

	results := make([]queryResult, 0, 1)

	typeID := C.CFGetTypeID(resultsRef)
	if typeID == C.CFDataGetTypeID() {
		b, err := cFDataToBytes(C.CFDataRef(resultsRef))
		if err != nil {
			return queryResult{}, err
		}
		item := queryResult{Data: b}
		results = append(results, item)
	} else {
		return queryResult{}, fmt.Errorf("invalid result type: %v", resultsRef)
	}

	return results[0], nil
}

// kcDelete removes an entry from the keychain
func kcDelete(service, name string) error {
	k := make(map[string]interface{})
	k["svce"] = service
	k["class"] = C.CFTypeRef(C.kSecClassGenericPassword)
	k["acct"] = name

	cfDict, err := convertMapToCFDictionary(k)
	if err != nil {
		return err
	}
	defer release(C.CFTypeRef(cfDict))

	errCode := C.SecItemDelete(cfDict)
	return checkError(errCode)
}

// release releases memory pointed to by a CFTypeRef.
func release(ref C.CFTypeRef) {
	C.CFRelease(ref)
}

// bytesToCFData will return a CFDataRef and if non-nil, must be released with
// release(ref).
func bytesToCFData(b []byte) (C.CFDataRef, error) {
	var p *C.UInt8
	if len(b) > 0 {
		p = (*C.UInt8)(&b[0])
	}
	cfData := C.CFDataCreate(C.kCFAllocatorDefault, p, C.CFIndex(len(b)))
	if cfData == 0 {
		return 0, fmt.Errorf("CFDataCreate failed")
	}
	return cfData, nil
}

// cFDataToBytes converts CFData to bytes.
func cFDataToBytes(cfData C.CFDataRef) ([]byte, error) {
	return C.GoBytes(unsafe.Pointer(C.CFDataGetBytePtr(cfData)), C.int(C.CFDataGetLength(cfData))), nil
}

// mapToCFDictionary will return a CFDictionaryRef and if non-nil, must be
// released with release(ref).
func mapToCFDictionary(m map[C.CFTypeRef]C.CFTypeRef) (C.CFDictionaryRef, error) {
	var keys, values []C.uintptr_t
	for key, value := range m {
		keys = append(keys, C.uintptr_t(key))
		values = append(values, C.uintptr_t(value))
	}
	numValues := len(values)
	var keysPointer, valuesPointer *C.uintptr_t
	if numValues > 0 {
		keysPointer = &keys[0]
		valuesPointer = &values[0]
	}
	var cfDict C.CFDictionaryRef
	cfDict = C.CFDictionaryCreateSafe2(C.kCFAllocatorDefault, keysPointer, valuesPointer, C.CFIndex(numValues),
		&C.kCFTypeDictionaryKeyCallBacks, &C.kCFTypeDictionaryValueCallBacks) //nolint
	if cfDict == 0 {
		return 0, fmt.Errorf("CFDictionaryCreate failed")
	}
	return cfDict, nil
}

// stringToCFString will return a CFStringRef and if non-nil, must be released with
// release(ref).
func stringToCFString(s string) (C.CFStringRef, error) {
	bytes := []byte(s)
	var p *C.UInt8
	if len(bytes) > 0 {
		p = (*C.UInt8)(&bytes[0])
	}
	return C.CFStringCreateWithBytes(C.kCFAllocatorDefault, p, C.CFIndex(len(s)), C.kCFStringEncodingUTF8, C.false), nil
}

// convertMapToCFDictionary converts a map to a CFDictionary and if non-nil,
// must be released with release(ref).
func convertMapToCFDictionary(attr map[string]interface{}) (C.CFDictionaryRef, error) {
	m := make(map[C.CFTypeRef]C.CFTypeRef)
	for key, i := range attr {
		var valueRef C.CFTypeRef
		switch val := i.(type) {
		default:
			return 0, fmt.Errorf("unsupported value type: %v", reflect.TypeOf(i))
		case C.CFTypeRef:
			valueRef = val
		case bool:
			if val {
				valueRef = C.CFTypeRef(C.kCFBooleanTrue)
			} else {
				valueRef = C.CFTypeRef(C.kCFBooleanFalse)
			}
		case []byte:
			bytesRef, err := bytesToCFData(val)
			if err != nil {
				return 0, err
			}
			valueRef = C.CFTypeRef(bytesRef)
			defer release(valueRef)
		case string:
			stringRef, err := stringToCFString(val)
			if err != nil {
				return 0, err
			}
			valueRef = C.CFTypeRef(stringRef)
			defer release(valueRef)
		}
		keyRef, err := stringToCFString(key)
		if err != nil {
			return 0, err
		}
		m[C.CFTypeRef(keyRef)] = valueRef
		defer release(C.CFTypeRef(keyRef))
	}

	cfDict, err := mapToCFDictionary(m)
	if err != nil {
		return 0, err
	}
	return cfDict, nil
}
