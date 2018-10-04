package ReceiveSyslog

import (
	"bufio"
	"context"
	"fmt"
	"github.com/TIBCOSoftware/flogo-lib/core/trigger"
	"net"
	"strconv"
	"strings"
	"time"
)

const DATE_FORMAT_SYSLOG_3164 = "Jan 02 15:04:05";

type ServerTCP struct {
	port  int
	l     net.Listener
	m     map[string]*trigger.Handler
	delim string
}

func (s *ServerTCP) Start() error {

	var err error
	s.l, err = net.Listen("tcp4", strconv.Itoa(s.port))
	if err != nil {
		fmt.Println(err)
		return err
	}

	go s.listenLoop()
	return nil
}

func (s *ServerTCP) listenLoop() {
	for {
		c, err := s.l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go s.handleConnection(c)
	}
}

func (s *ServerTCP) Stop() error {
	s.l.Close()
	return nil
}

func (s *ServerTCP) handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		line, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		triggerData := map[string]interface{}{
			"body":      line,
			"eventTime": time.Now(),
			"sourceIP":  c.RemoteAddr().String(),
			"source":    c.RemoteAddr().String(),
			"message":   line,
		}

		for ; ; {
			if line[0] != '<' {
				break;
			}
			// valid syslog packet start
			idx := strings.Index(line, ">")
			if idx < 0 || idx > 4 {
				break;
			}
			pri, convErr := strconv.Atoi(line[1:idx])
			if convErr != nil {
				break;
			}
			triggerData["pri"] = pri
			if line[idx+1] == '1' && line[idx+1] == ' ' {
				break; // TODO Syslog 5124
			}
			idx++
			dateStr := line[idx : idx+len(DATE_FORMAT_SYSLOG_3164)]
			eventTime, errParser := time.Parse(DATE_FORMAT_SYSLOG_3164, dateStr)
			if errParser != nil {
				break
			}
			triggerData["eventTime"] = eventTime
			idx += len(DATE_FORMAT_SYSLOG_3164)
			if line[idx] != ' ' {
				break
			}
			idx++
			sp := strings.Index(line[idx:], " ")
			if sp < 0 {
				break
			}
			triggerData["source"] = line[idx : idx+sp]
			idx++
			triggerData["message"] = line[idx:]
			break;
		}

		for _, handler := range s.m {
			handler.Handle(context.Background(), triggerData);
		}
	}
}
