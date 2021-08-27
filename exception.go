package mbmaster

import "fmt"

// ref:https://www.simplymodbus.ca/exceptions.htm
func checkException(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("modbus error data length too short")
	}
	//ModbusError
	if (data[1] & 0x80) != 0 {
		return fmt.Errorf(exception(data[2]))
	}
	return nil
}

func exception(code byte) string {
	switch code {
	case 0x01:
		return fmt.Sprintf("modbus error (%02X illegal function)", code)
	case 0x02:
		return fmt.Sprintf("modbus error (%02X illegal data address)", code)
	case 0x03:
		return fmt.Sprintf("modbus error (%02X illegal data value)", code)
	case 0x04:
		return fmt.Sprintf("modbus error (%02X slave device failure)", code)
	case 0x05:
		return fmt.Sprintf("modbus error (%02X acknowledge)", code)
	case 0x06:
		return fmt.Sprintf("modbus error (%02X slave device busy)", code)
	case 0x07:
		return fmt.Sprintf("modbus error (%02X negative acknowledge)", code)
	case 0x08:
		return fmt.Sprintf("modbus error (%02X memory parity error)", code)
	case 0x10:
		return fmt.Sprintf("modbus error (%02X gateway path unavailable)", code)
	case 0x11:
		return fmt.Sprintf("modbus error (%02X gateway target device failed to respond)", code)

	}
	return fmt.Sprintf("modbus error (%02X error code not in list)", code)
}
