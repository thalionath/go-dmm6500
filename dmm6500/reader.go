package dmm6500

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type Settings struct {
	VoltageRange    int
	PowerLineCycles float64
	AvgFilterSize   int
}

type reader struct {
	conn net.Conn
}

func NewReader(
	address string,
	s Settings,
) (*reader, error) {

	conn, err := net.Dial("tcp", address)

	if err != nil {
		return &reader{}, err
	}

	buffer := "defbuffer1"
	cmdRead := fmt.Sprintf("READ? \"%s\", READ, TSTamp", buffer)

	commands := []string{
		"*RST",         // This command resets the instrument settings to their default values and clears the reading buffers.
		"SYSTem:CLEar", // This command clears the event log
		"*CLS",         // This command clears the event registers and queues.
		// fmt.Sprintf("TRACe:MAKE \"%s\", 0", buffer),
		// fmt.Sprintf("TRACe:DELete \"%s\"", buffer),
		// fmt.Sprintf("TRACe:MAKE \"%s\", 0", buffer),
		// fmt.Sprintf("TRACe:FILL:MODE CONT, \"%s\"", buffer),
		// "DISPlay:BUFFer:ACTive \"defbuffer1\"",
		"SENS:FUNC \"VOLT:DC\"",
		fmt.Sprintf("SENS:VOLT:RANG %d", s.VoltageRange),
		// Auto Zero
		"SENS:VOLT:AZER ON",
		// Set the input impedance
		// Must be 10 MOhn for HV probe!
		"SENS:VOLT:INP MOHM10",
		// Sampling Frequency
		fmt.Sprintf("SENS:VOLT:NPLC %f", s.PowerLineCycles),
		// Enable the averaging filter
		// fmt.Sprintf("SENS:VOLT:AVER:COUN %d", s.AvgFilterSize),
		// "SENS:VOLT:AVER:TCON REP",
		// "SENS:VOLT:AVER ON",
		// TRACe:CLEar \"defbuffer1\"",
		// "TRACe:CLEar \"defbuffer2\"",
		// fmt.Sprintf("TRACe:CLEar \"%s\"", buffer),
		// "*WAI",
		// Kick off polling loop
		cmdRead,
	}

	for _, cmd := range commands {
		log.Println(cmd)
		cmdBytes := append([]byte(cmd), '\n')
		if _, err := conn.Write(cmdBytes); err != nil {
			conn.Close()
			return &reader{}, err
		}
	}

	scanner := bufio.NewScanner(conn)

	go func() {
		counter := 0
		for scanner.Scan() {
			counter++
			log.Printf("%d %v", counter, scanner.Text())
			cmdBytes := append([]byte(cmdRead), '\n')
			if _, err := conn.Write(cmdBytes); err != nil {
				log.Printf("error while writing %v", err)
				return
			}
		}
		log.Println("exiting scanner goroutine")
	}()

	return &reader{
		conn: conn,
	}, nil
}

func (r *reader) Close() {
	r.conn.Close()
}
