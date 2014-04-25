package util

import (
	"net"
	"time"
)

const ntpServer string = "10.pool.ntp.org"

func Time() (*time.Time, error) {
	raddr, err := net.ResolveUDPAddr("udp", ntpServer+":123")
	if err != nil {
		return nil, err
	}
	
	con, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}

	defer con.Close()
	con.SetDeadline(time.Now().Add(10 * time.Second))

	data := make([]byte, 48)
	data[0] = 3<<3 | 3

	_, err = con.Write(data)
	if err != nil {
		return nil, err
	}

	_, err = con.Read(data)
	if err != nil {
		return nil, err
	}

	var sec, frac uint64
	sec = uint64(data[43]) | uint64(data[42])<<8 | uint64(data[41])<<16 | uint64(data[40])<<24
	frac = uint64(data[47]) | uint64(data[46])<<8 | uint64(data[45])<<16 | uint64(data[44])<<24

	nsec := sec * 1e9
	nsec += (frac * 1e9) >> 32

	t := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(nsec)).Local()

	return &t, nil
}

