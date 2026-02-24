package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

type inboundMessage struct {
	Action string `json:"action"`
	Format string `json:"format,omitempty"`
	Raw    string `json:"raw,omitempty"`
}

type host struct {
	outMu sync.Mutex
}

func main() {
	cfg := parseFlags()
	if cfg.DoInstall {
		if err := runInstall(cfg); err != nil {
			log.Fatal(err)
		}
		return
	}

	log.SetOutput(os.Stderr)
	log.SetPrefix("card-reader-host ")

	h := &host{}
	go h.listenTCP(9099)
	go h.listenPCSC()

	if err := h.readNativeLoop(os.Stdin); err != nil {
		log.Fatal(err)
	}
}

func (h *host) readNativeLoop(r io.Reader) error {
	for {
		msg, err := readNativeMessage(r)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		var in inboundMessage
		if err := json.Unmarshal(msg, &in); err != nil {
			h.send(map[string]any{"error": err.Error()})
			continue
		}

		h.handleInput(in.Format, in.Raw)
	}
}

func (h *host) listenTCP(port int) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Printf("tcp listener failed: %v", err)
		return
	}
	defer ln.Close()
	log.Printf("listening TCP reader on %d", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go h.handleConn(conn)
	}
}

func (h *host) handleConn(c net.Conn) {
	defer c.Close()
	log.Printf("reader connected: %s", c.RemoteAddr())
	s := bufio.NewScanner(c)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}

		// Accept both "FORMAT:DATA" and plain raw lines from keyboard-wedge/bridge tools.
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			h.handleInput(parts[0], parts[1])
			continue
		}

		log.Printf("reader raw line (no format): %q", line)
		h.send(DecodeResult{Raw: line})
	}

	if err := s.Err(); err != nil {
		log.Printf("reader scan error: %v", err)
	}

	log.Printf("reader disconnected: %s", c.RemoteAddr())
}

func (h *host) handleInput(format, raw string) {
	format = strings.ToUpper(strings.TrimSpace(format))
	raw = strings.TrimSpace(raw)
	result := DecodeResult{Raw: raw}

	switch format {
	case "W34", "W34B":
		w34, err := decodeW34B(raw)
		if err != nil {
			h.send(map[string]any{"error": err.Error(), "raw": raw, "format": format})
			return
		}
		result.W34B = w34
	case "W26":
		w26, err := decodeW26(raw)
		if err != nil {
			h.send(map[string]any{"error": err.Error(), "raw": raw, "format": format})
			return
		}
		result.W26 = w26
	default:
		log.Printf("unsupported format %q, forwarding raw value", format)
		h.send(result)
		return
	}

	h.send(result)
}

func (h *host) send(v any) {
	h.outMu.Lock()
	defer h.outMu.Unlock()

	payload, err := json.Marshal(v)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(payload)))
	if _, err := os.Stdout.Write(lenBuf[:]); err != nil {
		log.Printf("stdout len write: %v", err)
		return
	}
	if _, err := os.Stdout.Write(payload); err != nil {
		log.Printf("stdout payload write: %v", err)
	}
}

func readNativeMessage(r io.Reader) ([]byte, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint32(lenBuf[:])
	if length == 0 {
		return []byte(`{}`), nil
	}
	buf := make([]byte, length)
	_, err := io.ReadFull(r, buf)
	return buf, err
}
