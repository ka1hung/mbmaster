
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
        // read
        log.Println(mb.ReadCoil(1, 0, 11))   //func1
        log.Println(mb.ReadCoilIn(1, 0, 11)) //func2
        log.Println(mb.ReadReg(1, 0, 11))    //func3
        log.Println(mb.ReadRegIn(1, 1, 11))  //func4

        // write
        log.Println(mb.WriteCoil(1, 120, true))                       //func5
        log.Println(mb.WriteCoils(1, 122, []bool{true, false, true})) //func15
        log.Println(mb.WriteReg(1, 0, 123))                           //func6
        log.Println(mb.WriteRegs(1, 1, []uint16{1, 2, 3}))            //func16
    }
    
