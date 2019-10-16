package types

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

// Name represents the account name
type Name string

// IsValidName verifies whether a string can represent a valid name or not.
func IsValidName(s string) bool {
	nameCheck := fmt.Sprintf("^[a-z0-9]{2,%d}(\\.[a-z0-9]{1,%d}){0,%d}$", 16, 10, 2)
	return regexp.MustCompile(nameCheck).MatchString(s)
}

// StrToName  returns Name with string of s.
func StrToName(s string) Name {
	n, err := parseName(s)
	if err != nil {
		panic(err)
	}
	return n
}

func parseName(s string) (Name, error) {
	var n Name
	if !n.SetString(s) {
		return n, fmt.Errorf("invalid name %v", s)
	}
	return n, nil
}

// BytesToName returns Name with value b.
func BytesToName(b []byte) (Name, error) {
	return parseName(string(b))
}

// BigToName returns Name with byte values of b.
func BigToName(b *big.Int) (Name, error) { return BytesToName(b.Bytes()) }

// SetString  sets the name to the value of b..
func (n *Name) SetString(s string) bool {
	//if !IsValidName(s) {
	//	return false
	//}
	*n = Name(s)
	return true
}

// UnmarshalText parses a hash in hex syntax.
func (n *Name) UnmarshalText(input []byte) error {
	return n.UnmarshalJSON(input)
}

// UnmarshalJSON parses a hash in hex syntax.
func (n *Name) UnmarshalJSON(data []byte) error {
	input := strings.TrimSpace(string(data))
	if len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"' {
		input = input[1: len(input)-1]
	}
	if len(input) > 0 {
		dec, err := parseName(string(input))
		if err != nil {
			return err
		}
		*n = dec
	}
	return nil
}

//// EncodeRLP implements rlp.Encoder
//func (n *Name) EncodeRLP(w io.Writer) error {
//	str := n.String()
//	if len(str) != 0 {
//		if _, err := parseName(str); err != nil {
//			return err
//		}
//	}
//	rlp.Encode(w, str)
//	return nil
//}
//
//// DecodeRLP implements rlp.Decoder
//func (n *Name) DecodeRLP(s *rlp.Stream) error {
//	var str string
//	err := s.Decode(&str)
//	if err == nil {
//		if len(str) != 0 {
//			name, err := parseName(str)
//			if err != nil {
//				return err
//			}
//			*n = name
//		}
//	}
//	return err
//}

// String implements fmt.Stringer.
func (n Name) String() string {
	return string(n)
}

// Big converts a name to a big integer.
func (n Name) Big() *big.Int { return new(big.Int).SetBytes([]byte(n.String())) }
