# Receive Syslog trigger

This trigger receives syslog messages.

## installation
```
flogo install github.com/TIBCOSoftware/lmi-flogo-collectors/ReceiveSyslog
```

## Settings

### Trigger
|Setting|Description|
|-----|----|
|protocol|protocol to use, only tcp is supported at the moment|
|port|port to bind for listening to incoming syslog messages|

### Handler
|Setting|Description|
|-----|----|
|regex|Filter messages (not implemented yet)|

## Schema
|name|type|Description|
|----|---|---|
|body|string| The whole message received|
|severity|integer|The severity|
|facility|integer|The facility|
|sourceIP|string|The source of the syslog connection|
|source|string|The source from the syslog header|
|eventTime|long|The time of the message as a UNIX timestamp|
|message|string|The syslog message without its header|

## Notes
Only Syslog in RFC 3164 format is properly supported right now