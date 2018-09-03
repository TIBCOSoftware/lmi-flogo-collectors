# SendSyslog
This activity allows you to send syslog messages using Syslog/UDP, Syslog/TCP or Syslog/TLS

## Installation
### Flogo Web
Add the activity using its GitHub reference <TBD>
### Flogo CLI
```bash
flogo add activity <TBD>
```

## Schema
Inputs and Outputs:

```json
{
  "inputs":[
     {
       "name": "protocol",
       "type": "string",
       "value": "udp",
       "required": true
     },
     {
       "name": "host",
       "type": "string",
       "value": "",
       "required" : true
     },
     {
       "name": "port",
       "type": "integer",
       "value": 514,
       "required" : false
     },
     {
       "name": "facility",
       "type": "integer",
       "value": 16,
       "required": true
     },
     {
       "name": "severity",
       "type": "integer",
       "value": 6,
       "required": true
     },
     {
       "name": "hostname",
       "type": "string",
       "value": null,
       "required": false
     },
     {
       "name": "appName",
       "type": "string",
       "value": null,
       "required": false
     },
     {
       "name": "msgId",
       "type": "string",
       "value": null,
       "required": false
     },
     {
       "name": "flowInfo",
       "type": "boolean",
       "value": "false"
     },
     {
       "name": "message",
       "type": "string",
       "value": "",
       "required" : true
     }
   ],
   "outputs": [
   ]
 }
```
## Settings
| Setting     | Required | Description |
|:------------|:---------|:------------|
| protocol     | True    | The protocol to use: udp, tcp, tls |
| host     | True    | The hostname/IP of the Syslog receiver |
| port     | False | The port of the Syslog receiver (default 514) |
| facility | True | The Syslog facility to use |
| severity | True | The Syslog severity to use |
| hostname | False | The hostname/IP to put in the Syslog header |
| appName | False | The app name to put in the Syslog header |
| msgId | False | The message ID to put in the Syslog header |
| flowInfo    | False    | If set to true this will append the flow information to the log message |
| message     | True    | The body of the message to send |

## Remarks

For the moment TLS connections are made in insecure mode, with no certificate validation.
```