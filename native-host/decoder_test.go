package main

import "testing"

func TestDecodeW34BExample(t *testing.T) {
	res, err := decodeW34B("46FF05D")
	if err != nil {
		t.Fatal(err)
	}
	if res.ExpandedHex != "5DF50F46" {
		t.Fatalf("unexpected transform: %s", res.ExpandedHex)
	}
}

func TestDecodeW26Bits(t *testing.T) {
	res, err := decodeW26("10110011000000010010100001")
	if err != nil {
		t.Fatal(err)
	}
	if res.Facility != 102 || res.CardNumber != 592 {
		t.Fatalf("unexpected values: %+v", res)
	}
}
