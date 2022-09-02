package models

import (
	"bytes"
	"database/sql/driver"
	"encoding/csv"
	"fmt"
	"io"
	"math/big"
)

type TrackHeight struct {
	Name   string `gorm:"primary_key"`
	Height uint64
}

type EpochInfo struct {
	ID         uint64         `gorm:"primary_key"`
	Validators SQLStringArray `gorm:"type:varchar(4096)"`
}

type Validator struct {
	StakeAddress     string
	ConsensusAddress string  `gorm:"primary_key"`
	Commission       *BigInt `gorm:"type:varchar(64)"`
	TotalStake       *BigInt `gorm:"type:varchar(64)"`
	SelfStake        *BigInt `gorm:"type:varchar(64)"`
}

type StakeInfo struct {
	StakeAddress     string  `gorm:"primary_key"`
	ConsensusAddress string  `gorm:"primary_key"`
	Amount           *BigInt `gorm:"type:varchar(64)"`
}

type DoneTx struct {
	Hash   string `gorm:"primary_key"`
	Height uint64
}

type TotalGas struct {
	Height   uint64  `gorm:"primary_key"`
	TotalGas *BigInt `gorm:"type:varchar(64)"`
}

type GasFee struct {
	ID      uint64 `gorm:"primary_key"`
	Address string
	Height  uint64
	GasFee  *BigInt `gorm:"type:varchar(64)"`
}

type Rewards struct {
	Address string  `gorm:"primary_key"`
	Height  uint64  `gorm:"primary_key"`
	Amount  *BigInt `gorm:"type:varchar(64)"`
}

type AccumulatedRewards struct {
	Name   string  `gorm:"primary_key"`
	Amount *BigInt `gorm:"type:varchar(64)"`
}

type CommunityRate struct {
	Name   string  `gorm:"primary_key"`
	Amount *BigInt `gorm:"type:varchar(64)"`
}

// SQLStringArray is a string array stored in the database as comma separated values.
type SQLStringArray []string

// Scan implements the sql Scanner interface.
func (arr *SQLStringArray) Scan(src interface{}) error {
	if src == nil {
		*arr = nil
	}
	v, err := driver.String.ConvertValue(src)
	if err != nil {
		return fmt.Errorf("failed to scan StringArray")
	}
	str, ok := v.(string)
	if !ok {
		return nil
	}

	buf := bytes.NewBufferString(str)
	r := csv.NewReader(buf)
	ret, err := r.Read()
	if err != nil && err != io.EOF {
		return fmt.Errorf("badly formatted csv string array: %s", err)
	}
	*arr = ret
	return nil
}

// Value implements the driver Valuer interface.
func (arr SQLStringArray) Value() (driver.Value, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	err := w.Write(arr)
	if err != nil {
		return nil, fmt.Errorf("csv encoding of string array: %s", err)
	}
	w.Flush()
	return buf.String(), nil
}

type BigInt struct {
	big.Int
}

func NewBigInt(value *big.Int) *BigInt {
	return &BigInt{Int: *value}
}

func (bigInt *BigInt) Value() (driver.Value, error) {
	if bigInt == nil {
		return "null", nil
	}
	return bigInt.String(), nil
}

func (bigInt *BigInt) Scan(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("type error, not string")
	}
	if str == "null" || str == "nil" || str == "<nil>" || str == "" {
		return nil
	}
	data, ok := new(big.Int).SetString(str, 10)
	if !ok {
		return fmt.Errorf("not a valid big integer: %s", str)
	}
	bigInt.Int = *data
	return nil
}
