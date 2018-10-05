# LMI log collector for TIBCO Flogo

This contains two activities that can be used with TIBCO Flogo for sending messages to a TIBCO LMI physical or virtual 
appliance.

SendSyslog uses Syslog (UDP/TCP/TLS) to send the messages.

SendUldp uses ULDP protocol to send the messages. 

ReceiveSyslog is a trigger, receiving Syslog messages.

Each activity is located under its own directory.

A good example on how to use these is provided in [https://community.tibco.com/wiki/using-flogo-collect-tibco-loglogic-lmi-sources](https://community.tibco.com/wiki/using-flogo-collect-tibco-loglogic-lmi-sources)

