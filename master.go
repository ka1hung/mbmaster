package mbmaster

import (
	"errors"
	"time"

	// "mbmaster/serial"
	"github.com/goburrow/serial"
)

//Master Modbus Master config
type Master struct {
	Comport  string
	BaudRate int
	DataBits int
	StopBits int
	Parity   string
	Timeout  time.Duration
}

// NewMaster creates a new Modbus Master config.
func NewMaster(com string, br int, timeout time.Duration) *Master {
	m := &Master{}
	m.Comport = com
	m.BaudRate = br
	m.DataBits = 8
	m.StopBits = 1
	m.Parity = "N"
	m.Timeout = timeout
	return m
}

//ReadCoil mdbus function 1 qurry and return []uint16
func (m *Master) ReadCoil(id uint8, addr uint16, leng uint16) ([]bool, error) {

	wbuf := []byte{id, 0x01, byte(addr >> 8), byte(addr), byte(leng >> 8), byte(leng)}

	//write
	res, err := Qurry(m, wbuf)
	if err != nil {
		return []bool{}, err
	}

	//convert
	result := []bool{}
	bc := res[2]
	for i := 0; i < int(bc); i++ {
		for j := 0; j < 8; j++ {
			if (res[3+i] & (byte(1) << byte(j))) != 0 {
				result = append(result, true)
			} else {
				result = append(result, false)
			}
		}
	}
	result = result[:leng]

	return result, nil
}

//ReadCoilIn mdbus function 2 qurry and return []uint16
func (m *Master) ReadCoilIn(id uint8, addr uint16, leng uint16) ([]bool, error) {

	wbuf := []byte{id, 0x02, byte(addr >> 8), byte(addr), byte(leng >> 8), byte(leng)}

	//write
	res, err := Qurry(m, wbuf)
	if err != nil {
		return []bool{}, err
	}

	//convert
	result := []bool{}
	bc := res[2]
	for i := 0; i < int(bc); i++ {
		for j := 0; j < 8; j++ {
			if (res[3+i] & (byte(1) << byte(j))) != 0 {
				result = append(result, true)
			} else {
				result = append(result, false)
			}
		}
	}
	result = result[:leng]

	return result, nil
}

//ReadReg mdbus function 3 qurry and return []uint16
func (m *Master) ReadReg(id uint8, addr uint16, leng uint16) ([]uint16, error) {

	wbuf := []byte{id, 0x03, byte(addr >> 8), byte(addr), byte(leng >> 8), byte(leng)}

	//write
	res, err := Qurry(m, wbuf)
	if err != nil {
		return []uint16{}, err
	}

	//convert
	result := []uint16{}
	for i := 0; i < int(leng); i++ {
		var b uint16
		b = uint16(res[i*2+3]) << 8
		b |= uint16(res[i*2+4])
		result = append(result, b)
	}

	return result, nil
}

//ReadRegIn mdbus function 4 qurry and return []uint16
func (m *Master) ReadRegIn(id uint8, addr uint16, leng uint16) ([]uint16, error) {

	wbuf := []byte{id, 0x04, byte(addr >> 8), byte(addr), byte(leng >> 8), byte(leng)}

	//write
	res, err := Qurry(m, wbuf)
	if err != nil {
		return []uint16{}, err
	}

	//convert
	result := []uint16{}
	for i := 0; i < int(leng); i++ {
		var b uint16
		b = uint16(res[i*2+3]) << 8
		b |= uint16(res[i*2+4])
		result = append(result, b)
	}

	return result, nil
}

//WriteCoil mdbus function 5 qurry and return []uint16
func (m *Master) WriteCoil(id uint8, addr uint16, data bool) error {

	var wbuf = []byte{}
	if data == true {
		wbuf = []byte{id, 0x5, byte(addr >> 8), byte(addr), 0xff, 0x00}
	} else {
		wbuf = []byte{id, 0x5, byte(addr >> 8), byte(addr), 0x00, 0x00}
	}

	//write
	_, err := Qurry(m, wbuf)
	if err != nil {
		return err
	}

	return nil
}

//WriteReg mdbus function 6 qurry and return []uint16
func (m *Master) WriteReg(id uint8, addr uint16, data uint16) error {

	wbuf := []byte{id, 0x06, byte(addr >> 8), byte(addr), byte(data >> 8), byte(data)}

	//write
	_, err := Qurry(m, wbuf)
	if err != nil {
		return err
	}

	return nil
}

//WriteCoils mdbus function 15(0x0f) qurry and return []uint16
func (m *Master) WriteCoils(id uint8, addr uint16, data []bool) error {
	wbuf := []byte{}
	if len(data)%8 == 0 {
		wbuf = []byte{id, 0x0f, byte(addr >> 8), byte(addr), byte(len(data) >> 8), byte(len(data)), byte(len(data) / 8)}
	} else {
		wbuf = []byte{id, 0x0f, byte(addr >> 8), byte(addr), byte(len(data) >> 8), byte(len(data)), byte(len(data)/8) + 1}
	}
	var tbuf byte
	for i := 0; i < len(data); i++ {
		if data[i] {
			tbuf |= byte(1 << uint(i%8))
		}

		if (i+1)%8 == 0 || i == len(data)-1 {
			wbuf = append(wbuf, tbuf)
			tbuf = 0
		}
	}

	//write
	_, err := Qurry(m, wbuf)
	if err != nil {
		return err
	}

	return nil
}

//WriteRegs mdbus function 16(0x10) qurry and return []uint16
func (m *Master) WriteRegs(id uint8, addr uint16, data []uint16) error {

	wbuf := []byte{id, 0x10, byte(addr >> 8), byte(addr), byte(len(data) >> 8), byte(len(data)), byte(len(data)) * 2}

	for i := 0; i < len(data); i++ {
		wbuf = append(wbuf, byte(data[i]>>8))
		wbuf = append(wbuf, byte(data[i]))
	}

	//write
	_, err := Qurry(m, wbuf)
	if err != nil {
		return err
	}

	return nil
}

//Qurry function
func Qurry(m *Master, data []byte) ([]byte, error) {
	result := []byte{}
	//0. check
	if len(data) < 6 {
		return result, errors.New("length not enough(<6)")
	}

	//1.open serial port
	port, err := serial.Open(
		&serial.Config{
			Address:  m.Comport,
			BaudRate: m.BaudRate,
			DataBits: m.DataBits,
			StopBits: m.StopBits,
			Parity:   m.Parity,
			Timeout:  m.Timeout,
		})
	if err != nil {
		return result, err
	}
	defer port.Close()

	//2.write
	data = CrcAppend(data) //append crc
	_, err = port.Write(data)
	if err != nil {
		return result, err
	}

	//3.read
	rlen := 8
	if data[1] == 1 || data[1] == 2 {
		bs := int(data[4])<<8 | int(data[5])
		if bs%8 == 0 {
			rlen = bs/8 + 5
		} else {
			rlen = bs/8 + 5 + 1
		}

	} else if data[1] == 3 || data[1] == 4 {
		rlen = int(data[4])<<8 | int(data[5])*2 + 5
	}
	total := 0

	for {
		time.Sleep(10 * time.Millisecond)
		b := make([]byte, 1024)
		resLen, err := port.Read(b)
		if err != nil {
			return result, err
		}
		result = append(result, b[:resLen]...)
		total += resLen
		if total >= rlen {
			break
		}
	}
	// 4. check crc
	if CrcCheck(result) == false {
		return []byte{}, errors.New("CRC Error")
	}
	// 5. check status
	if result[1] >= 0x81 {
		return result, errors.New("Qurry got error status")
	}
	return result, nil
}
