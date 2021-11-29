### [v0.7](https://github.com/SumoLogic-Incubator/slogen/milestone/2?closed=1)

- **Feature** : Timeslice based budgeting (availability calculated w.r.t. good time windows)  
  - [Sample config](samples/openslo/ingest-lag-timeslice-budgeting.yaml)
  
- **Feature** : Track and filter your SLO by fields e.g. customerID, region etc
  - [Sample config](samples/openslo/ingest-lag-timeslice-budgeting.yaml)
  - [Screenshot](misc/SLO-breakdow.png)

- **Feature** : Overview of all SLO's configured for each service
  - [Screenshot](misc/service-overview.png)
  
- **Feature** : SLO budget forecasting
  - [Screenshot](misc/budget-forecast.png)

- **Feature** : subcommand to list connection id for use in alert notification field
  - `slogen list -c`

- **Fix** : "resource not found error" on changing service name for existing SLO

- **Fix** : tool unable to run behind https proxy
