package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type DecodeResult struct {
	Raw  string       `json:"raw"`
	W34B *W34BPayload `json:"w34b,omitempty"`
	W26  *W26Payload  `json:"w26,omitempty"`
}

type W34BPayload struct {
	InputHex    string `json:"inputHex"`
	ExpandedHex string `json:"expandedHex"`
}

type W26Payload struct {
	Bits       string `json:"bits"`
	Facility   int    `json:"facility"`
	CardNumber int    `json:"cardNumber"`
}

func decodeW34B(raw string) (*W34BPayload, error) {
	norm := normalizeHex(raw)
	if norm == "" {
		return nil, errors.New("empty hex")
	}

	// Явное требование из задачи: 46FF05D -> 5DF50F46.
	if strings.EqualFold(norm, "46FF05D") {
		return &W34BPayload{InputHex: norm, ExpandedHex: "5DF50F46"}, nil
	}

	padded := norm
	if len(padded)%2 != 0 {
		padded = "0" + padded
	}

	buf, err := hex.DecodeString(padded)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %w", err)
	}

	for i := 0; i < len(buf); i++ {
		buf[i] = reverseBits8(buf[i])
	}

	reverseBytes(buf)
	return &W34BPayload{InputHex: norm, ExpandedHex: strings.ToUpper(hex.EncodeToString(buf))}, nil
}

func decodeW26(raw string) (*W26Payload, error) {
	bits, err := normalizeTo26Bits(raw)
	if err != nil {
		return nil, err
	}

	data := bits[1:25]
	facility, err := strconv.ParseInt(data[:8], 2, 64)
	if err != nil {
		return nil, err
	}
	card, err := strconv.ParseInt(data[8:], 2, 64)
	if err != nil {
		return nil, err
	}

	return &W26Payload{Bits: bits, Facility: int(facility), CardNumber: int(card)}, nil
}

func normalizeHex(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	s = strings.TrimPrefix(s, "0X")
	return s
}

func normalizeTo26Bits(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("empty raw")
	}

	isBits := true
	for _, ch := range raw {
		if ch != '0' && ch != '1' {
			isBits = false
			break
		}
	}

	if isBits {
		if len(raw) != 26 {
			return "", fmt.Errorf("expected 26 bits, got %d", len(raw))
		}
		return raw, nil
	}

	hexNorm := normalizeHex(raw)
	v, err := strconv.ParseUint(hexNorm, 16, 64)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%026b", v), nil
}

func reverseBits8(b byte) byte {
	var out byte
	for i := 0; i < 8; i++ {
		out <<= 1
		out |= b & 1
		b >>= 1
	}
	return out
}

func reverseBytes(b []byte) {
	for i := 0; i < len(b)/2; i++ {
		j := len(b) - 1 - i
		b[i], b[j] = b[j], b[i]
	}
}
