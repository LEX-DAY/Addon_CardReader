//go:build windows

package main

import (
	"bufio"
	"log"
	"strings"
	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

const serialScanInterval = 3 * time.Second
const defaultSerialBaud = 9600

func (h *host) listenSerial() {
	log.Printf("serial auto-detect enabled")
	for {
		ports, err := enumerator.GetDetailedPortsList()
		if err != nil {
			log.Printf("serial list ports: %v", err)
			time.Sleep(serialScanInterval)
			continue
		}

		for _, p := range ports {
			if p == nil || p.Name == "" {
				continue
			}
			if !isLikelyZ2SerialPort(p) {
				continue
			}
			if !h.claimSerialPort(p.Name) {
				continue
			}
			go h.runSerialReader(p)
		}

		time.Sleep(serialScanInterval)
	}
}

func (h *host) runSerialReader(p *enumerator.PortDetails) {
	defer h.releaseSerialPort(p.Name)

	mode := &serial.Mode{
		BaudRate: defaultSerialBaud,
	}
	port, err := serial.Open(p.Name, mode)
	if err != nil {
		log.Printf("serial open %s failed: %v", p.Name, err)
		return
	}
	defer port.Close()

	log.Printf("serial reader connected: %s (%s)", p.Name, p.Product)

	s := bufio.NewScanner(port)
	s.Split(scanCRLF)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		h.handleReaderLine("serial:"+p.Name, line)
	}

	if err := s.Err(); err != nil {
		log.Printf("serial read %s: %v", p.Name, err)
	}

	log.Printf("serial reader disconnected: %s", p.Name)
}

func scanCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' || data[i] == '\r' {
			return i + 1, data[:i], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func isLikelyZ2SerialPort(p *enumerator.PortDetails) bool {
	info := strings.ToLower(strings.Join([]string{
		p.Name,
		p.Product,
		p.SerialNumber,
		p.VID,
		p.PID,
	}, " "))

	if strings.Contains(info, "bluetooth") {
		return false
	}

	if strings.Contains(info, "z-2") ||
		strings.Contains(info, "rd_all") ||
		strings.Contains(info, "ironlogic") {
		return true
	}

	// Generic USB-Serial adapters often used by Z-2 converters.
	return p.IsUSB && strings.Contains(info, "serial")
}

func (h *host) claimSerialPort(name string) bool {
	h.serialMu.Lock()
	defer h.serialMu.Unlock()
	if _, exists := h.serialPorts[name]; exists {
		return false
	}
	h.serialPorts[name] = struct{}{}
	return true
}

func (h *host) releaseSerialPort(name string) {
	h.serialMu.Lock()
	defer h.serialMu.Unlock()
	delete(h.serialPorts, name)
}
