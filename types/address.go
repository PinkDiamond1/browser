package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"

	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

const AddressLength = 20

var addressT = reflect.TypeOf(Address{})

// Address represents the 20 byte address of an  account.
type Address [AddressLength]byte

func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

// Bytes gets the string representation of the underlying address.
func (a Address) Bytes() []byte { return a[:] }

// Big converts an address to a big integer.
func (a Address) Big() *big.Int { return new(big.Int).SetBytes(a[:]) }

// Hash converts an address to a hash by left-padding it with zeros.
func (a Address) Hash() Hash { return BytesToHash(a[:]) }

// Hex returns an EIP55-compliant hex string representation of the address.
func (a Address) Hex() string {
	uncheckSummed := hex.EncodeToString(a[:])
	sha := sha3.NewLegacyKeccak256()
	sha.Write([]byte(uncheckSummed))
	hash := sha.Sum(nil)

	result := []byte(uncheckSummed)
	for i := 0; i < len(result); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}
	return "0x" + string(result)
}

// String implements fmt.Stringer.
func (a Address) String() string {
	return a.Hex()
}

// Format implements fmt.Formatter, forcing the byte slice to be formatted as is,
// without going through the stringer interface used for logging.
func (a Address) Format(s fmt.State, c rune) {
	_, _ = fmt.Fprintf(s, "%"+string(c), a[:])
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a) it will panic.
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

// MarshalText returns the hex representation of a.
func (a Address) MarshalText() ([]byte, error) {
	return hexutil.Bytes(a[:]).MarshalText()
}

func (a Address) Compare(x Address) int {
	return bytes.Compare(a.Bytes(), x.Bytes())
}

// UnmarshalText parses a hash in hex syntax.
func (a *Address) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Address", input, a[:])
}

// UnmarshalJSON parses a hash in hex syntax.
func (a *Address) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(addressT, input, a[:])
}

const PubKeyLength = 65

type PubKey [PubKeyLength]byte

var pubKeyT = reflect.TypeOf(PubKey{})

//Bytes return bytes
func (p PubKey) Bytes() []byte { return p[:] }

// Big converts a hash to a big integer.
func (p PubKey) Big() *big.Int { return new(big.Int).SetBytes(p[:]) }

// Hex converts a hash to a hex string.
func (p PubKey) Hex() string { return hexutil.Encode(p[:]) }

//SetBytes set bytes to publicKey
func (p *PubKey) SetBytes(key []byte) {
	if len(key) > len(p) {
		key = key[len(key)-PubKeyLength:]
	}
	copy(p[PubKeyLength-len(key):], key)
}

// String implements fmt.Stringer.
func (p PubKey) String() string {
	return p.Hex()
}

// MarshalText returns the hex representation of a.
func (p PubKey) MarshalText() ([]byte, error) {
	return hexutil.Bytes(p[:]).MarshalText()
}

// UnmarshalText parses a hash in hex syntax.
func (p *PubKey) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("PubKey", input, p[:])
}

// UnmarshalJSON parses a hash in hex syntax.
func (p *PubKey) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(pubKeyT, input, p[:])
}

// Compare returns an integer comparing two byte slices lexicographically.
// The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
// A nil argument is equivalent to an empty slice.
func (p PubKey) Compare(x PubKey) int {
	return bytes.Compare(p.Bytes(), x.Bytes())
}

// HexToPubKey returns PubKey with byte values of s.
func HexToPubKey(s string) PubKey { return BytesToPubKey(FromHex(s)) }

// BytesToPubKey returns PubKey with value b.
func BytesToPubKey(b []byte) PubKey {
	var a PubKey
	a.SetBytes(b)
	return a
}
