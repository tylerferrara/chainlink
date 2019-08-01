package assets

import (
	"database/sql/driver"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/willf/pad"
)

func format(i *big.Int, precision int) string {
	v := "1" + pad.Right("", precision, "0")
	d := &big.Int{}
	d.SetString(v, 10)
	r := &big.Rat{}
	r.SetFrac(i, d)
	return fmt.Sprintf("%v", r.FloatString(precision))
}

// Link contains a field to represent the smallest units of LINK
type Link big.Int

// NewLink returns a new struct to represent LINK from it's smallest unit
func NewLink(w int64) *Link {
	return (*Link)(big.NewInt(w))
}

// String returns Link formatted as a string.
func (l *Link) String() string {
	if l == nil {
		return "0"
	}
	return format((*big.Int)(l), 18)
}

// SetInt64 delegates to *big.Int.SetInt64
func (l *Link) SetInt64(w int64) *Link {
	return (*Link)((*big.Int)(l).SetInt64(w))
}

// ToHash returns a 32 byte representation of this value
func (l *Link) ToHash() common.Hash {
	return common.BigToHash((*big.Int)(l))
}

// Set delegates to *big.Int.Set
func (l *Link) Set(x *Link) *Link {
	il := (*big.Int)(l)
	ix := (*big.Int)(x)

	w := il.Set(ix)
	return (*Link)(w)
}

// SetString delegates to *big.Int.SetString
func (l *Link) SetString(s string, base int) (*Link, bool) {
	w, ok := (*big.Int)(l).SetString(s, base)
	return (*Link)(w), ok
}

// Cmp defers to big.Int Cmp
func (l *Link) Cmp(y *Link) int {
	return (*big.Int)(l).Cmp((*big.Int)(y))
}

// Add defers to big.Int Add
func (l *Link) Add(x, y *Link) *Link {
	il := (*big.Int)(l)
	ix := (*big.Int)(x)
	iy := (*big.Int)(y)

	return (*Link)(il.Add(ix, iy))
}

// Text defers to big.Int Text
func (l *Link) Text(base int) string {
	return (*big.Int)(l).Text(base)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (l *Link) MarshalText() ([]byte, error) {
	return (*big.Int)(l).MarshalText()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (l *Link) UnmarshalText(text []byte) error {
	if _, ok := l.SetString(string(text), 10); !ok {
		return fmt.Errorf("assets: cannot unmarshal %q into a *assets.Link", text)
	}
	return nil
}

// IsZero returns true when the value is 0 and false otherwise
func (l *Link) IsZero() bool {
	zero := big.NewInt(0)
	return (*big.Int)(l).Cmp(zero) == 0
}

// Symbol returns LINK
func (*Link) Symbol() string {
	return "LINK"
}

// Value returns the Link value for serialization to database.
func (l Link) Value() (driver.Value, error) {
	b := (big.Int)(l)
	return b.String(), nil
}

// Scan reads the database value and returns an instance.
func (l *Link) Scan(value interface{}) error {
	switch temp := value.(type) {
	case string:
		_, ok := l.SetString(temp, 10)
		if !ok {
			return fmt.Errorf("Unable to scan Link string from %s", temp)
		}
		return nil
	case int64:
		return fmt.Errorf("Unable to convert %v of %T to Link, is the sql type set to varchar?", value, value)
	default:
		return fmt.Errorf("Unable to convert %v of %T to Link", value, value)
	}
}

// Eth contains a field to represent the smallest units of ETH
type Eth big.Int

// NewEth returns a new struct to represent ETH from it's smallest unit
func NewEth(w int64) *Eth {
	return (*Eth)(big.NewInt(w))
}

// Cmp delegates to *big.Int.Cmp
func (e *Eth) Cmp(y *Eth) int {
	return e.ToInt().Cmp(y.ToInt())
}

func (e *Eth) String() string {
	return format(e.ToInt(), 18)
}

// SetInt64 delegates to *big.Int.SetInt64
func (e *Eth) SetInt64(w int64) *Eth {
	return (*Eth)(e.ToInt().SetInt64(w))
}

// SetString delegates to *big.Int.SetString
func (e *Eth) SetString(s string, base int) (*Eth, bool) {
	w, ok := e.ToInt().SetString(s, base)
	return (*Eth)(w), ok
}

// MarshalText implements the encoding.TextMarshaler interface.
func (e *Eth) MarshalText() ([]byte, error) {
	return e.ToInt().MarshalText()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (e *Eth) UnmarshalText(text []byte) error {
	if _, ok := e.SetString(string(text), 10); !ok {
		return fmt.Errorf("assets: cannot unmarshal %q into a *assets.Eth", text)
	}
	return nil
}

// IsZero returns true when the value is 0 and false otherwise
func (e *Eth) IsZero() bool {
	zero := big.NewInt(0)
	return e.ToInt().Cmp(zero) == 0
}

// Symbol returns ETH
func (*Eth) Symbol() string {
	return "ETH"
}

// ToInt returns the Eth value as a *big.Int.
func (e *Eth) ToInt() *big.Int {
	return (*big.Int)(e)
}
