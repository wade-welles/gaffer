package main

import (
	"flag"
	"fmt"
	"go.uber.org/zap"
	"os"
)

func failOnErr(err error) {
	if err != nil {
		fmt.Println("error: ", err.Error())
		os.Exit(1)
	}
}

func main() {
	logger, _ := zap.NewDevelopment()
	var (
		mode  = flag.String("mode", "", "server mode [server, master, agent, zookeeper]")
		dbStr = flag.String("db", "./gaffer.db", "database connection string")
	)
	flag.Parse()
	switch *mode {
	case "server":
		db, err := NewSQLStore(*dbStr, logger)
		failOnErr(err)
		defer db.Close()
		server := NewServer(db, logger)
		server.Anonymous = true
		failOnErr(server.Serve())
	case "master":
	case "agent":
	case "zookeeper":
	default:
		flag.PrintDefaults()
		failOnErr(fmt.Errorf("Invalid server mode %s", *mode))
	}
	/*
			p := process{
				cmd:  exec.Command("/home/kevin/repos/go/src/github.com/vektorlab/gaffer/script.sh"),
				env:  map[string]string{"COOL": "BEANS"},
				err:  make(chan error),
				quit: make(chan struct{}, 1),
				log:  logger,
			}
			if err := p.Start(); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		loop:
			for {
				select {
				case err := <-p.err:
					fmt.Println("Uh oh", err.Error())
					break loop
				case <-p.quit:
					fmt.Println("quittin")
					break loop
				default:
					time.Sleep(100 * time.Millisecond)
				}
			}
	*/
}
