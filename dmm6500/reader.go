package dmm6500

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"
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
		// cmdRead,
	}

	for _, cmd := range commands {
		log.Println(cmd)
		cmdBytes := append([]byte(cmd), '\n')
		if _, err := conn.Write(cmdBytes); err != nil {
			conn.Close()
			return &reader{}, err
		}
	}

	// The DMM does not seem to reset some TCP connection related
	// buffer when connecting to it.
	// If we do not wait for the response to a READ? command in the
	// previous connection but instead disconnect the connection
	// immediately, then the response will be sent unsolicited
	// after the subsequent connection is established.
	if err := flushInput(conn, 100*time.Millisecond); err != nil {
		conn.Close()
		return &reader{}, err
	}

	go func() {
		counter := 0
		cmdBytes := append([]byte(cmdRead), '\n')
		for {
			counter++
			if _, err := conn.Write(cmdBytes); err != nil {
				log.Printf("error while writing %v", err)
			}
			line, err := readResponse(conn)
			if err != nil {
				log.Println(err)
				break
			}
			log.Printf("%d %v", counter, line)
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

func readResponse(conn net.Conn) (string, error) {

	// timeout should depend on averager settings
	if err := conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		return "", err
	}

	line := make([]byte, 0, 128)

	for {
		buffer := make([]byte, 128)
		n, err := conn.Read(buffer)

		if err != nil {
			return "", err
		}

		if n == 0 {
			return "", errors.New("connection closed")
		}

		for i := 0; i < n; i++ {
			if buffer[i] == '\n' {
				line = append(line, buffer[:i]...)
				return string(line), nil
			}
		}

		line = append(line, buffer...)
	}
}

// flushInput reads all data from the connection until
// there is silence for a while
func flushInput(conn net.Conn, silence time.Duration) error {
	discarder := make([]byte, 4096)
	for {
		if err := conn.SetReadDeadline(time.Now().Add(silence)); err != nil {
			return err
		}

		n, err := conn.Read(discarder)

		if e, ok := err.(net.Error); ok && e.Timeout() {
			return nil
		} else if err != nil {
			return err
		}

		if n == 0 {
			return errors.New("connection closed")
		}
	}
}
