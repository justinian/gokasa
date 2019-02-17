package gokasa

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
)

const (
	default_port int = 9999
)

var (
	ErrOutputTooShort = fmt.Errorf("Output buffer too short")
	ErrBadPacket      = fmt.Errorf("Received a bad packet from the device")
)

/// The HS300 Kasa smart strip
type PowerStrip struct {
	address string

	Alias     string  `json:"alias"`
	DeviceId  string  `json:"deviceId"`
	ErrorCode int     `json:"err_code"`
	MAC       string  `json:"mac"`
	Plugs     []*Plug `json:"children"`
}

func NewPowerStrip(address string) (*PowerStrip, error) {
	if !strings.Contains(address, ":") {
		address = fmt.Sprintf("%s:%d", address, default_port)
	}

	ps := &PowerStrip{address: address}
	if err := ps.RefreshSystemInfo(); err != nil {
		return nil, err
	}

	return ps, nil
}

func (ps *PowerStrip) RefreshSystemInfo() error {
	command := []byte(`{"system":{"get_sysinfo":{}}}`)
	respbytes, err := ps.sendUDPCommand(command)
	if err != nil {
		return err
	}

	var response map[string]map[string]*PowerStrip
	if err := json.Unmarshal(respbytes, &response); err != nil {
		return err
	}

	if system, ok := response["system"]; ok {
		if info, ok := system["get_sysinfo"]; ok {
			info.address = ps.address
			*ps = *info
			for _, p := range ps.Plugs {
				p.powerStrip = ps
			}
			return nil
		}
	}

	return ErrBadPacket
}

func (ps *PowerStrip) sendTCPCommand(command []byte) ([]byte, error) {
	output := make([]byte, len(command)+4)
	binary.BigEndian.PutUint32(output, uint32(len(command)))
	if err := rotorCommand(command, output[4:]); err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", ps.address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.Write(output)
	if err != nil {
		return nil, err
	}

	lenbuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lenbuf); err != nil {
		return nil, err
	}

	resplen := binary.BigEndian.Uint32(lenbuf)
	if resplen > 4096 {
		return nil, ErrBadPacket
	}

	response := make([]byte, resplen)
	if _, err := io.ReadFull(conn, response); err != nil {
		return nil, err
	}

	return derotorCommand(response), nil
}

func (ps *PowerStrip) sendUDPCommand(command []byte) ([]byte, error) {
	output := make([]byte, len(command))
	if err := rotorCommand(command, output); err != nil {
		return nil, err
	}

	conn, err := net.Dial("udp", ps.address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.Write(output)
	if err != nil {
		return nil, err
	}

	response := make([]byte, 4096)
	n, err := conn.Read(response)
	if err != nil {
		return nil, err
	}

	return derotorCommand(response[:n]), nil
}

func (ps *PowerStrip) setStates(plugs []*Plug, s int) error {
	plug_ids := make([]string, len(plugs))
	for i, p := range plugs {
		p.State = s
		plug_ids[i] = fmt.Sprintf(`"%s%s"`, ps.DeviceId, p.Id)
	}

	f := `{"context":{"child_ids":[%s]},"system":{"set_relay_state":{"state":%d}}}`
	c := []byte(fmt.Sprintf(f, strings.Join(plug_ids, ","), s))

	respbytes, err := ps.sendTCPCommand(c)
	if err != nil {
		return err
	}

	var kr KasaResponse
	if err := json.Unmarshal(respbytes, &kr); err != nil {
		return err
	}

	return kr.GetErrors()
}
