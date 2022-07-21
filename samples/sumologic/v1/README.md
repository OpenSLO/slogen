# Creating Sumologic native SLO's and monitors from OpenSLO yaml configs

slogen support native sumologic slo creation for spec version v1 for OpenSLO. 

Samples can be found in the [samples/sumologic/v1/](samples/sumologic/v1/) directory. 


### OpenSLO SLI to Sumologic SLI fields mapping 

- the `name` value specified in `metadata:name` field of the OpenSLO spec is used as the SLO name
- Budgeting method `Occurrences` maps to `Request` based SLO's while `Timeslices` maps to `Window` based SLO's
- use `"sumologic-logs` as `metricSource:type` for logs based Indicators, while `sumologic-metrics` for metrics based Indicators 

### Using Alerting Policies to create SLO monitors 
For now slogen doesn't support inline alerting conditional in SLO config yaml. 
Alerting policies need to specified separately as showed in the samples and then can be referred by name the SLO config. 

To specify the `critical` and `warning` threshold use the same condition `name` while changing the severity level as done in the samples. 
