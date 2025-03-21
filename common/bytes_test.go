package common

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestCopyBytes(t *testing.T) {
	input := []byte{1, 2, 3, 4}

	v := CopyBytes(input)
	if !bytes.Equal(v, input) {
		t.Fatal("not equal after copy")
	}
	v[0] = 99
	if bytes.Equal(v, input) {
		t.Fatal("result is not a copy")
	}
}

func TestLeftPadBytes(t *testing.T) {
	val := []byte{1, 2, 3, 4}
	padded := []byte{0, 0, 0, 0, 1, 2, 3, 4}

	if r := LeftPadBytes(val, 8); !bytes.Equal(r, padded) {
		t.Fatalf("LeftPadBytes(%v, 8) == %v", val, r)
	}
	if r := LeftPadBytes(val, 2); !bytes.Equal(r, val) {
		t.Fatalf("LeftPadBytes(%v, 2) == %v", val, r)
	}
}

func TestRightPadBytes(t *testing.T) {
	val := []byte{1, 2, 3, 4}
	padded := []byte{1, 2, 3, 4, 0, 0, 0, 0}

	if r := RightPadBytes(val, 8); !bytes.Equal(r, padded) {
		t.Fatalf("RightPadBytes(%v, 8) == %v", val, r)
	}
	if r := RightPadBytes(val, 2); !bytes.Equal(r, val) {
		t.Fatalf("RightPadBytes(%v, 2) == %v", val, r)
	}
}

func TestFromHex(t *testing.T) {
	input := "0x01"
	expected := []byte{1}
	result := FromHex(input)
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %x got %x", expected, result)
	}
}

func TestHexDecode(t *testing.T) {
	input := "0x1"
	_, err := hex.DecodeString(input)
	t.Log(err)
	if err == nil {
		t.Fatalf("hex.DecodeString(%v) should return an error", input)
	}
}

func TestIsHex(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
	}{
		{"", true},
		{"0", false},
		{"00", true},
		{"a9e67e", true},
		{"A9E67E", true},
		{"0xa9e67e", false},
		{"a9e67e001", false},
		{"0xHELLO_MY_NAME_IS_STEVEN_@#$^&*", false},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if ok := isHex(test.input); ok != test.ok {
				t.Errorf("isHex(%q) = %v, want %v", test.input, ok, test.ok)
			}
		})
	}
}

func TestFromHexOddLength(t *testing.T) {
	input := "0x1"
	expected := []byte{1}
	result := FromHex(input)
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %x got %x", expected, result)
	}
}

func TestNoPrefixShortHexOddLength(t *testing.T) {
	input := "1"
	expected := []byte{1}
	result := FromHex(input)
	if !bytes.Equal(expected, result) {
		t.Errorf("Expected %x got %x", expected, result)
	}
}

func TestTrimRightZeroes(t *testing.T) {
	tests := []struct {
		arr []byte
		exp []byte
	}{
		{FromHex("0x00ffff00ff0000"), FromHex("0x00ffff00ff")},
		{FromHex("0x00000000000000"), []byte{}},
		{FromHex("0xff"), FromHex("0xff")},
		{[]byte{}, []byte{}},
		{FromHex("0x00ffffffffffff"), FromHex("0x00ffffffffffff")},
	}
	for i, test := range tests {
		got := TrimRightZeroes(test.arr)
		if !bytes.Equal(got, test.exp) {
			t.Errorf("test %d, got %x exp %x", i, got, test.exp)
		}
	}
}

func TestParseHexOrString(t *testing.T) {
	tests := []struct {
		input string
		want  []byte
	}{
		{"0x123456", []byte{18, 52, 86}},
		{"123456", []byte{49, 50, 51, 52, 53, 54}},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got, err := ParseHexOrString(test.input)
			if err != nil {
				t.Errorf("ParseHexOrString(%v) failed: %v", test.input, err)
			}
			if !bytes.Equal(got, test.want) {
				t.Errorf("ParseHexOrString(%v) = %v, want %v", test.input, got, test.want)
			}
		})
	}
}
