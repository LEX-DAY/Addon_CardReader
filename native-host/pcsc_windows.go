//go:build windows

package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ebfe/scard"
)

const pcscPollInterval = 400 * time.Millisecond

var getUIDAPDUs = [][]byte{
	{0xFF, 0xCA, 0x00, 0x00, 0x00},
	{0xFF, 0xCA, 0x00, 0x00, 0x04},
}

func (h *host) listenPCSC() {
	ctx, err := scard.EstablishContext()
	if err != nil {
		log.Printf("pcsc disabled: %v", err)
		return
	}
	defer ctx.Release()

	log.Printf("pcsc enabled")

	present := make(map[string]bool)
	lastRaw := make(map[string]string)
	lastListErr := ""

	for {
		readers, err := ctx.ListReaders()
		if err != nil {
			if isScardError(err, scard.ErrNoReadersAvailable) {
				time.Sleep(pcscPollInterval)
				continue
			}
			msg := err.Error()
			if msg != lastListErr {
				log.Printf("pcsc list readers: %v", err)
				lastListErr = msg
			}
			time.Sleep(2 * time.Second)
			continue
		}
		lastListErr = ""

		seen := make(map[string]struct{}, len(readers))
		for _, reader := range readers {
			seen[reader] = struct{}{}

			raw, source, err := readPCSCValue(ctx, reader)
			if err != nil {
				if isCardAbsentError(err) {
					present[reader] = false
					lastRaw[reader] = ""
					continue
				}
				log.Printf("pcsc read [%s]: %v", reader, err)
				continue
			}

			if raw == "" {
				continue
			}

			if !present[reader] || lastRaw[reader] != raw {
				log.Printf("pcsc card [%s]: %s (%s)", reader, raw, source)
				switch source {
				case "uid":
					// For ACR/Z-2 parity use W34B conversion path so output becomes e.g. 5DF50F46.
					h.handleInput("W34B", raw)
				case "atr":
					if isKnownPseudoATR(raw) {
						// Common ATR of contactless cards in CCID mode; not card-unique identifier.
						log.Printf("pcsc atr [%s] ignored (non-unique): %s", reader, raw)
						break
					}
					h.send(DecodeResult{Raw: raw})
				default:
					h.send(DecodeResult{Raw: raw})
				}
			}

			present[reader] = true
			lastRaw[reader] = raw
		}

		// Cleanup states for readers that disappeared from the system.
		for reader := range present {
			if _, ok := seen[reader]; !ok {
				delete(present, reader)
				delete(lastRaw, reader)
			}
		}

		time.Sleep(pcscPollInterval)
	}
}

func readPCSCValue(ctx *scard.Context, reader string) (raw string, source string, err error) {
	card, err := ctx.Connect(reader, scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		return "", "", err
	}
	defer card.Disconnect(scard.LeaveCard)

	status, err := card.Status()
	if err != nil {
		return "", "", err
	}

	var lastUIDErr error
	for _, apdu := range getUIDAPDUs {
		rsp, txErr := card.Transmit(apdu)
		if txErr == nil {
			if uid, ok := parseUIDResponse(rsp); ok {
				return uid, "uid", nil
			}
			lastUIDErr = fmt.Errorf("uid apdu % X returned unexpected status: % X", apdu, rsp)
			continue
		}
		lastUIDErr = txErr
	}

	atr := strings.ToUpper(hex.EncodeToString(status.Atr))
	if atr != "" {
		return atr, "atr", nil
	}

	if lastUIDErr != nil {
		return "", "", lastUIDErr
	}

	return "", "", errors.New("pcsc empty card data")
}

func parseUIDResponse(rsp []byte) (string, bool) {
	if len(rsp) < 4 {
		return "", false
	}
	sw1 := rsp[len(rsp)-2]
	sw2 := rsp[len(rsp)-1]
	if sw1 != 0x90 || sw2 != 0x00 {
		return "", false
	}
	uid := rsp[:len(rsp)-2]
	if len(uid) == 0 {
		return "", false
	}
	return strings.ToUpper(hex.EncodeToString(uid)), true
}

func isScardError(err error, target scard.Error) bool {
	var se scard.Error
	if errors.As(err, &se) {
		return se == target
	}
	return false
}

func isCardAbsentError(err error) bool {
	return isScardError(err, scard.ErrNoSmartcard) ||
		isScardError(err, scard.ErrRemovedCard) ||
		isScardError(err, scard.ErrTimeout) ||
		isScardError(err, scard.ErrCardUnsupported)
}

func isKnownPseudoATR(raw string) bool {
	atr := strings.ToUpper(strings.TrimSpace(raw))
	return strings.HasPrefix(atr, "3B8F8001804F0CA00000030603")
}
