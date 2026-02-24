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

func TestDecodeW26Pair(t *testing.T) {
	res, err := decodeW26("096,17669")
	if err != nil {
		t.Fatal(err)
	}
	if res.Facility != 96 || res.CardNumber != 17669 {
		t.Fatalf("unexpected values: %+v", res)
	}
	if len(res.Bits) != 26 {
		t.Fatalf("unexpected bits length: %d", len(res.Bits))
	}
}

func TestBuildW26BitsParity(t *testing.T) {
	bits := buildW26Bits(96, 17669)
	if bits != "10110000001000101000001011" {
		t.Fatalf("unexpected bits: %s", bits)
	}
}
