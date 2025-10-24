package token_test

import (
	"encoding/json"
	"testing"

	"github.com/mulan-ext/auth/token"
)

var data = token.DefaultData{
	Token_:   "token",
	ID_:      1,
	Account_: "account",
	Roles_:   []string{"admin"},
	Items_:   map[string]any{"key": "value"},
}

func TestDataJSON(t *testing.T) {
	buf, err := json.Marshal(&data)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(buf))
	v := &token.DefaultData{}
	err = json.Unmarshal(buf, v)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
}

func TestDataSliceBinary(t *testing.T) {
	buf, err := data.Roles_.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(buf))
	v := &token.DefaultData{}
	err = v.Roles_.UnmarshalBinary(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v.Roles_)
}

func TestDataMapBinary(t *testing.T) {
	buf, err := data.Items_.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(buf))
	v := &token.DefaultData{}
	err = v.Items_.UnmarshalText(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v.Items_)
}
