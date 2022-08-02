# Creating Sumologic native SLO's and monitors from OpenSLO yaml configs

slogen supports native sumologic [SLO]() creation from spec version `v1` of OpenSLO. 

Samples can be found in the [samples/sumologic/v1](/samples/sumologic/v1) directory. 

They can be tried out as below from the root of the repo.

```shell
go run main.go samples/sumologic/v1 --clean --onlyNative=true --sloFolder=slogen-v1 --sloMonitorFolder=slogen-v
```

### OpenSLO SLI to Sumologic SLI fields mapping 

- the `name` value specified in `metadata:name` field of the OpenSLO spec is used as the SLO name
- Budgeting method `Occurrences` maps to `Request` based SLO's while `Timeslices` maps to `Window` based SLO's
- use `sumologic-logs` as `metricSource:type` for logs based Indicators, while `sumologic-metrics` for metrics based Indicators 

### Using Alerting Policies to create SLO monitors 
For now slogen doesn't support inline alerting conditional in SLO config yaml. 
Alerting policies need to specified separately as showed in the samples and then can be referred by name the SLO config. 

To specify the `critical` and `warning` threshold use the same condition `name` while changing the severity level as done in the samples. 


### alert notification targets

Notification targets can be specified in the OpenSLO [format](https://github.com/OpenSLO/OpenSLO#alertnotificationtarget) 
while using the annotations field for specifying the extra parameters required. 
These targets can then be referred in the alert policy config to generate the monitors.


Examples below

##### email based notification target

```yaml
apiVersion: openslo/v1
kind: AlertNotificationTarget
spec:
  description: Notifies by a mail message to the on-call devops mailing group
  target: email
metadata:
  name: OnCallDevopsMailNotification
  annotations:
    recipients: "agaurav@sumologic.com"
    subject: "Monitor Alert: {{TriggerType}} on {{Name}}"
    message_body: "Triggered {{TriggerType}} Alert on {{Name}}: {{QueryURL}}"
    time_zone: "PST"
    run_for_triggers: "Critical,ResolvedCritical"      
```

##### connection based notification target

```yaml
apiVersion: openslo/v1
kind: AlertNotificationTarget
spec:
  description: Notifies by a mail message to the on-call devops mailing group
  target: connection
metadata:
  name: OnCallDevopsSlackNotification
  annotations:
    connection_type : "Webhook"
    connection_id: "0000000000000431"
    run_for_triggers: "Critical,ResolvedCritical"
```


