/*
 * Copyright © 2018. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */

package SendSyslog

import (
	"crypto/tls"
	"fmt"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// log is the default package logger
var log = logger.GetLogger("activity-SendSyslog")

// ActivitySendSyslog implementation
type ActivitySendSyslog struct {
	metadata      *activity.Metadata
	connectionMap *map[string]net.Conn
}

// NewActivity creates a new activity
func NewActivity(metadata *activity.Metadata) activity.Activity {
	ret := &ActivitySendSyslog{metadata: metadata}
	newMap := make(map[string]net.Conn)
	ret.connectionMap = &newMap
	return ret
}

// Metadata implements activity.Activity.Metadata
func (a *ActivitySendSyslog) Metadata() *activity.Metadata {
	return a.metadata
}

// Eval implements activity.Activity.Eval
func (a *ActivitySendSyslog) Eval(context activity.Context) (done bool, err error) {

	var useTcp = false
	var useTls = false

	wsProto := strings.ToLower(context.GetInput("protocol").(string))
	switch wsProto {
	case "udp":
	case "tcp":
		useTcp = true
	case "tls":
		useTcp = true
		useTls = true
	default:
		log.Errorf("Unsupported protocol: %s", wsProto)
		return false, nil
	}

	wsHost := context.GetInput("host").(string)
	iPort := context.GetInput("port").(int)

	var connKey string
	connKey = wsProto + ":" + wsHost + ":" + strconv.Itoa(iPort)

	var conn net.Conn

	if (*a.connectionMap)[connKey] == nil {
		if useTls {
			tlsConfig := tls.Config{}
			tlsConfig.ServerName = wsHost
			tlsConfig.InsecureSkipVerify = true
			log.Infof("Creating tls connection to %s:%d", wsHost, iPort)

			conn, err = tls.Dial("tcp", wsHost+":"+strconv.Itoa(iPort), &tlsConfig)
		} else {
			conn, err = net.Dial(wsProto, wsHost+":"+strconv.Itoa(iPort))
		}
		if err != nil {
			log.Errorf("Error while opening %s connection %v to %s:%d", wsProto, err, wsHost, iPort)
			return false, err
		}
		(*a.connectionMap)[connKey] = conn
	} else {
		conn = (*a.connectionMap)[connKey]
	}

	iFacility := context.GetInput("facility").(int)
	iSeverity := context.GetInput("severity").(int)
	iPriority := iFacility*8 + iSeverity

	t := time.Now()

	timeStr := t.Format(time.RFC3339)

	var hostname string

	if context.GetInput("hostname") != nil && len(context.GetInput("hostname").(string)) > 0 {
		hostname = context.GetInput("hostname").(string)
	} else {
		hostname, err = os.Hostname()
		if err != nil {
			hostname = "-"
		}
	}

	var appName string

	if context.GetInput("appName") != nil && len(context.GetInput("appName").(string)) > 0 {
		appName = context.GetInput("appName").(string)
		strings.Replace(appName, " ", "", -1)
		appName = "flogo-" + appName
	} else {
		arg0 := os.Args[0]
		arg0parts := strings.Split(arg0, "/")
		appName = arg0parts[len(arg0parts)-1]
	}

	procId := os.Getpid()
	var msgId string
	if context.GetInput("msgId") != nil && len(context.GetInput("msgId").(string)) > 0 {
		msgId = context.GetInput("msgId").(string)
		strings.Replace(msgId, " ", "", -1)
	} else {
		msgId = "-"
	}

	sd := "-"

	wsMessage := context.GetInput("message").(string)

	flowInfo := context.GetInput("flowInfo").(bool)

	if flowInfo {
		wsMessage = fmt.Sprintf("'%s' - FlowInstanceID [%s], Flow [%s], Task [%s]", wsMessage,
			context.FlowDetails().ID(), context.FlowDetails().Name(), context.TaskName())
	}

	syslogMsg := fmt.Sprintf("<%d>1 %s %s %s %d %s %s %s", iPriority, timeStr, hostname, appName, procId, msgId, sd, wsMessage)

	log.Infof("Sending to %s:%s:%d, payload '%s'", wsProto, wsHost, iPort, syslogMsg)

	if !useTcp {
		fmt.Fprintf(conn, "%s", syslogMsg)
	} else if !useTls {
		fmt.Fprintf(conn, "%s\n", syslogMsg)
	} else {
		msgLen := len(syslogMsg) //TODO check if this is the real size in bytes
		logger.Infof("msgLen=%d", msgLen)
		fmt.Fprintf(conn, "%d%s", msgLen, syslogMsg)
	}

	return true, nil
}
