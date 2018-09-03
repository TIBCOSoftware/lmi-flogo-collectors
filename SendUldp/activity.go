/*
 * Copyright © 2018. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */

 package SendUldp

import (
	"fmt"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"os"
	"strconv"
	"strings"
	"time"
	"TIBCOSoftware/loglmi/uldp"
	"net"
)

// log is the default package logger
var log = logger.GetLogger("activity-SendUldp")

// ActivitySendSyslog implementation
type ActivitySendUldp struct {
	metadata      *activity.Metadata
	connectionMap *map[string]*uldp.Sender
}

// NewActivity creates a new activity
func NewActivity(metadata *activity.Metadata) activity.Activity {
	ret := &ActivitySendUldp{metadata: metadata}
	newMap := make(map[string]*uldp.Sender)
	ret.connectionMap = &newMap
	return ret
}

// Metadata implements activity.Activity.Metadata
func (a *ActivitySendUldp) Metadata() *activity.Metadata {
	return a.metadata
}

// Eval implements activity.Activity.Eval
func (a *ActivitySendUldp) Eval(context activity.Context) (done bool, err error) {

	wsHost := context.GetInput("host").(string)
	iPort := context.GetInput("port").(int)

	var connKey string
	connKey = "uldp:" + wsHost + ":" + strconv.Itoa(iPort)

	var sender *uldp.Sender

	if (*a.connectionMap)[connKey] == nil {
		settings := uldp.CreateConnectionSettings(wsHost, false)
		if context.GetInput("origin") != nil {
			originStr := context.GetInput("origin").(string)
			originIpAddr, err := net.ResolveIPAddr("ip", originStr)
			if err == nil {
				settings.Origin = originIpAddr
			}
		}
		if context.GetInput("deviceDomain") != nil {
			settings.DomainName = context.GetInput("deviceDomain").(string)
		}
		mySender, err := uldp.CreateSender(settings)
		if err != nil {
			return false, err
		}
		sender = &mySender
		err = sender.Connect()
		if err != nil {
			return false, err
		}
		(*a.connectionMap)[connKey] = sender
	} else {
		sender = (*a.connectionMap)[connKey]
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

	flattenJson := context.GetInput("flattenJson").(bool)
	if (flattenJson) {
		wsMessage = toFlatText(wsMessage)
	}

	flowInfo := context.GetInput("flowInfo").(bool)

	if flowInfo {
		wsMessage = fmt.Sprintf("'%s' - FlowInstanceID [%s], Flow [%s], Task [%s]", wsMessage,
			context.FlowDetails().ID(), context.FlowDetails().Name(), context.TaskName())
	}

	syslogMsg := fmt.Sprintf("<%d>1 %s %s %s %d %s %s %s", iPriority, timeStr, hostname, appName, procId, msgId, sd, wsMessage)

	log.Infof("Sending to %s:%s:%d, payload '%s'", "uldp", wsHost, iPort, syslogMsg)

	logMesssage := uldp.CreateUldpSyslogMessage(t, nil, syslogMsg)

	err = sender.SendMessage(&logMesssage)
	if err != nil {
		return false, err
	}

	err = sender.Flush()
	if err != nil {
		return true, err
	}

	return true, nil
}
