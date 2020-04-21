
# Simple modbusRTU master tool

# Quick start

## install

    go get -u github.com/ka1hung/mbmaster

## simple example

    package main

    import (
        "log"
        "time"

        "github.com/ka1hung/mbmaster"
    )

    func main() {
        mb := mbmaster.NewMaster("COM3", 9600, time.Second)
        data, _ := mb.ReadReg(1, 0, 4)
        log.Println(data)

        mb.WriteReg(1, 0, 1)
        mb.WriteRegs(1, 1, []uint16{2, 3, 4})

        data, _ = mb.ReadReg(1, 0, 4)
        log.Println(data)
    }
    
