//go:build windows

package checker

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"netra-uptime/pkg/utils"
	"syscall"
	"time"
	"unsafe"
)

var (
	modiphlpapi         = syscall.NewLazyDLL("iphlpapi.dll")
	procIcmpCreateFile  = modiphlpapi.NewProc("IcmpCreateFile")
	procIcmpSendEcho    = modiphlpapi.NewProc("IcmpSendEcho")
	procIcmpCloseHandle = modiphlpapi.NewProc("IcmpCloseHandle")
)

type ipOptionInformation struct {
	Ttl         uint8
	Tos         uint8
	Flags       uint8
	OptionsSize uint8
	OptionsData uintptr
}

type icmpEchoReply struct {
	Address       uint32
	Status        uint32
	RoundTripTime uint32
	DataSize      uint16
	Reserved      uint16
	Data          uintptr
	Options       ipOptionInformation
}

func icmpEcho(ip net.IP, timeout time.Duration) (time.Duration, error) {
	if timeout <= 0 {
		return 0, errors.New("ping timed out")
	}

	if err := utils.IsRestrictedIP(ip); err != nil {
		return 0, err
	}

	ip4 := ip.To4()
	if ip4 == nil {
		return 0, errICMPUnavailable
	}

	destAddr := binary.LittleEndian.Uint32(ip4)

	handle, _, _ := procIcmpCreateFile.Call()
	if handle == ^uintptr(0) || handle == 0 {
		return 0, errICMPUnavailable
	}
	defer func() {
		_, _, _ = procIcmpCloseHandle.Call(handle)
	}()

	sendData := []byte("netra-uptime")
	replyBuf := make([]byte, unsafe.Sizeof(icmpEchoReply{})+uintptr(len(sendData))+16)

	ms := timeout.Milliseconds()
	var timeoutMs uint32
	if ms <= 0 {
		timeoutMs = 1000
	} else if ms > math.MaxUint32 {
		timeoutMs = math.MaxUint32
	} else {
		timeoutMs = uint32(ms)
	}

	start := time.Now()
	// #nosec G103 -- Win32 IcmpSendEcho requires raw pointers to in-memory byte slices
	ret, _, _ := procIcmpSendEcho.Call(
		handle,
		uintptr(destAddr),
		uintptr(unsafe.Pointer(&sendData[0])),
		uintptr(len(sendData)),
		0,
		uintptr(unsafe.Pointer(&replyBuf[0])),
		uintptr(len(replyBuf)),
		uintptr(timeoutMs),
	)

	if ret == 0 {
		return time.Since(start), errors.New("ping timed out: no ICMP echo reply received")
	}

	// #nosec G103 -- replyBuf is sized to hold icmpEchoReply plus payload
	reply := (*icmpEchoReply)(unsafe.Pointer(&replyBuf[0]))
	if reply.Status != 0 {
		return time.Since(start), fmt.Errorf("icmp echo failed with status %d", reply.Status)
	}

	if reply.RoundTripTime > 0 {
		return time.Duration(reply.RoundTripTime) * time.Millisecond, nil
	}
	return time.Since(start), nil
}
