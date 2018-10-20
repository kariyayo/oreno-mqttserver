package packet

import (
	"encoding/binary"
	"fmt"
)

type ConnectFlags struct {
	CleanSession bool
	WillFlag     bool
	WillQoS      byte
	WillRetain   bool
	PasswordFlag bool
	UserNameFlag bool
}

func (c *ConnectFlags) ToBytes() []byte {
	var b byte
	if c.CleanSession {
		b = onbit(b, 1)
	}
	if c.WillFlag {
		b = onbit(b, 2)
	}
	switch c.WillQoS {
	case 0:
	case 1:
		b = onbit(b, 3)
	case 2:
		b = onbit(b, 4)
	case 3:
		b = onbit(b, 3)
		b = onbit(b, 4)
	}
	if c.WillRetain {
		b = onbit(b, 5)
	}
	if c.PasswordFlag {
		b = onbit(b, 6)
	}
	if c.UserNameFlag {
		b = onbit(b, 7)
	}
	return []byte{b}
}

type ConnectVariableHeader struct {
	ProtocolName  string
	ProtocolLevel uint8
	ConnectFlags  ConnectFlags
	KeepAlive     uint16
}

func (h *ConnectVariableHeader) ToBytes() []byte {
	var result []byte
	result = append(result, 0)
	result = append(result, 4)
	result = append(result, []byte(h.ProtocolName)...)

	result = append(result, h.ProtocolLevel)

	result = append(result, h.ConnectFlags.ToBytes()...)

	keepAlive := make([]byte, 2)
	binary.BigEndian.PutUint16(keepAlive, h.KeepAlive)
	result = append(result, keepAlive...)

	return result
}

type ConnackVariableHeader struct {
	SessionPresent bool
	ReturnCode     uint8
}

type ConnectError struct {
	msg string
}

func (e *ConnectError) Error() string {
	return e.msg
}

func (h *ConnackVariableHeader) ToBytes() []byte {
	var result []byte
	if h.SessionPresent {
		result = append(result, 1)
	} else {
		result = append(result, 0)
	}
	result = append(result, h.ReturnCode)
	return result
}

func ToConnectVariableHeader(fixedHeader FixedHeader, bs []byte) (ConnectVariableHeader, []byte, error) {
	err := checkIsConnectPacket(fixedHeader, bs)
	if err != nil {
		return ConnectVariableHeader{}, nil, err
	}

	protocolLevel := bs[6]
	if protocolLevel != 4 {
		return ConnectVariableHeader{}, nil, &ConnectError{fmt.Sprintf("protocol level is not supported. it got is %v", protocolLevel)}
	}

	connectFlagsBytes := bs[7]
	connectFlags := ConnectFlags{}

	// TODO now, support only 1
	cleanSession := refbit(connectFlagsBytes, 1)
	if cleanSession != 1 {
		return ConnectVariableHeader{}, nil, &ConnectError{fmt.Sprintf("clean session value in connect flags must be 1. it got is %v", cleanSession)}
	}
	connectFlags.CleanSession = true

	// TODO now, support only 0
	willFlag := refbit(connectFlagsBytes, 2)
	if willFlag != 0 {
		return ConnectVariableHeader{}, nil, &ConnectError{fmt.Sprintf("will flag value in connect flags must be 0. it got is %v", willFlag)}
	}
	connectFlags.WillFlag = false

	// TODO now, support only QoS0
	willQoS2 := refbit(connectFlagsBytes, 3)
	willQoS1 := refbit(connectFlagsBytes, 4)
	if willQoS2 != 0 || willQoS1 != 0 {
		return ConnectVariableHeader{}, nil, &ConnectError{fmt.Sprintf("will QoS value in connect flags must be 0. it got is %v %v", willQoS1, willQoS2)}
	}
	connectFlags.WillQoS = 0

	// TODO now, support only 0
	willRetain := refbit(connectFlagsBytes, 5)
	if willRetain != 0 {
		return ConnectVariableHeader{}, nil, &ConnectError{fmt.Sprintf("will retain value in connect flags must be 0. it got is %v", willRetain)}
	}
	connectFlags.WillRetain = false

	// TODO now, support only 0
	passwordFlag := refbit(connectFlagsBytes, 6)
	if passwordFlag != 0 {
		return ConnectVariableHeader{}, nil, &ConnectError{fmt.Sprintf("password flag value in connect flags must be 0. it got is %v", passwordFlag)}
	}
	connectFlags.PasswordFlag = false

	// TODO now, support only 0
	userNameFlag := refbit(connectFlagsBytes, 7)
	if userNameFlag != 0 {
		return ConnectVariableHeader{}, nil, &ConnectError{fmt.Sprintf("user name flag value in connect flags must be 0. it got is %v", userNameFlag)}
	}
	connectFlags.UserNameFlag = false

	keepAlive := binary.BigEndian.Uint16(bs[8:10])

	result := ConnectVariableHeader{
		ProtocolName:  "MQTT",
		ProtocolLevel: 4,
		ConnectFlags:  connectFlags,
		KeepAlive:     keepAlive,
	}
	return result, bs[10:], nil
}

func checkIsConnectPacket(fixedHeader FixedHeader, bs []byte) error {
	if fixedHeader.PacketType != CONNECT {
		return fmt.Errorf("packet type is invalid. it got is %v", fixedHeader.PacketType)
	}

	protocolName := bs[:6]
	if !isValidProtocolName(protocolName) {
		return fmt.Errorf("protocol name is invalid. it got is %q", protocolName)
	}

	connectFlagsBytes := bs[7]
	reserved := connectFlagsBytes & 1
	if reserved != 0 {
		return fmt.Errorf("reserved value in connect flags must be 0. it got is %v", reserved)
	}
	return nil
}

func isValidProtocolName(protocolName []byte) bool {
	return len(protocolName) == 6 &&
		protocolName[0] == 0 && protocolName[1] == 4 &&
		protocolName[2] == 'M' && protocolName[3] == 'Q' && protocolName[4] == 'T' && protocolName[5] == 'T'
}
