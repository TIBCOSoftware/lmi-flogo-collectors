# SendULDP
This activity allows you to send log messages to LMI using ULDP (realtime syslog messages).

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
      "name": "host",
      "type": "string",
      "value": "",
      "required" : true
    },
    {
      "name": "port",
      "type": "integer",
      "value": 5515,
      "required" : false
    },
    {
      "name": "origin",
      "type": "string",
      "value": null,
      "required": false
    },
    {
      "name": "deviceDomain",
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
      "name": "flattenJson",
      "type": "boolean",
      "value": "false",
      "required": false
    },
    {
      "name": "message",
      "type": "string",
      "value": "",
      "required" : true
    }
}
```
## Settings
| Setting     | Required | Description |
|:------------|:---------|:------------|
| host     | True    | The hostname/IP of the Syslog receiver |
| port     | False | The port of the Syslog receiver (default 514) |
| origin | False | The IP address to use in ULDP as message origin (defaults to host IP) |
| deviceDomain | False | The LMI device domain to use |
| appName | False | The app name to put in the Syslog header |
| msgId | False | The message ID to put in the Syslog header |
| flowInfo    | False    | If set to true this will append the flow information to the log message |
| flattenJson | False | If true, assumes the message is a JSON payload and flattens it to CSV |
| message     | True    | The body of the message to send |


## Remarks
There is no support for secure ULDP (TLS) for now
To work, the ULDP for GO library has to be added as a contribution (available as ZIP package in LMI supplemental disk)
```