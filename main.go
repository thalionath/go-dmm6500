package main

import (
	"log"
	"time"

	"github.com/thalionath/go-dmm6500/dmm6500"
)

func run() {
	reader, err := dmm6500.NewReader(
		"10.10.10.130:5025",
		dmm6500.Settings{
			VoltageRange:    10,
			PowerLineCycles: 1, // 50 Hz / 1 = 50 Hz
			AvgFilterSize:   1, // Max 100
		},
	)

	if err != nil {
		log.Panic(err)
	}

	defer reader.Close()

	time.Sleep(2 * time.Second)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Print("Startup")

	for i := 0; i < 4; i++ {
		log.Printf("Run #%v", i)
		run()
		time.Sleep(2 * time.Second)
	}
}
