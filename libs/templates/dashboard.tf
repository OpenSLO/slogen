data "sumologic_admin_recommended_folder" "admin_folder" {}

data "sumologic_personal_folder" "personalFolder" {}

resource "sumologic_folder" "slo_folder" {
  name        = "SLO Tracker"
  description = "Your SLO dashboards"
  parent_id   = data.sumologic_personal_folder.personalFolder.id
}

resource "sumologic_folder" "tsat" {
  name        = "tsat"
  description = "tsat trackers"
  parent_id   = sumologic_folder.slo_folder.id
}


resource "sumologic_dashboard" "api-dashboard" {
  title            = "Anomaly generation"
  description      = "Tracks objective : checks anomaly are being calculated and with acceptable delays"
  folder_id        = sumologic_folder.tsat.id
  refresh_interval = 300
  theme            = "Light"

  time_range {
    begin_bounded_time_range {
      from {
        literal_time_range {
          range_name = "today"
        }
      }
    }
  }

  topology_label_map {
    data {
      label  = "service"
      values = ["tsat"]
    }
  }

  ## text panel
  panel {
    text_panel {
      key                                         = "text-panel-01"
      title                                       = "Service Health"
      visual_settings                             = jsonencode(
      {
        "text" : {
          "verticalAlignment" : "top",
          "horizontalAlignment" : "left",
          "fontSize" : 12
        }
      }
      )
      keep_visual_settings_consistent_with_parent = true
      text                                        = <<-EOF
                ## Service Level Objective

                Use this dashboard to track reliability of TSAT service. It contains following panels:

                1. Current availability
                3. Weekly Burn rate
                3. Daily Burn rate
                4. Monthly Forecasted value
            EOF
    }
  }

  ## search panel - log query
  panel {
    sumo_search_panel {
      key                                         = "search-panel-01"
      title                                       = "Hourly Burn Rate"
      description                                 = ""
      # stacked time series
      visual_settings                             = jsonencode(
      {
        "general" : {
          "mode" : "timeSeries",
          "type" : "area",
          "displayType" : "stacked",
          "markerSize" : 5,
          "lineDashType" : "solid",
          "markerType" : "square",
          "lineThickness" : 1
        },
        "title" : {
          "fontSize" : 14
        },
        "legend" : {
          "enabled" : true,
          "verticalAlign" : "bottom",
          "fontSize" : 12,
          "maxHeight" : 50,
          "showAsTable" : false,
          "wrap" : true
        }
      }
      )
      keep_visual_settings_consistent_with_parent = true
      query {
        query_string = <<QUERY
_sourceCategory=tsat-batcher | json auto | where msg="finished drift genQuery run for active customers"
| timeslice 1h
| if ( !isNull(duration) and duration > 320 ,1,0) as isGood
| sum(isGood) as totalGood, count as totalCount by _timeslice
| totalGood/totalCount as SLO
QUERY
        query_type   = "Logs"
        query_key    = "A"
      }
      time_range {
        begin_bounded_time_range {
          from {
            relative_time_range {
              relative_time = "-12h"
            }
          }
        }
      }
    }
  }

  ## search panel - metrics query
  panel {
    sumo_search_panel {
      key                                         = "metrics-panel-01"
      title                                       = "Total Availability"
      description                                 = "good requests/total requests"
      # pie chart
      visual_settings                             = jsonencode(
      {
        "general" : {
          "mode" : "distribution",
          "type" : "pie",
          "displayType" : "default",
          "fillOpacity" : 1,
          "startAngle" : 270,
          "innerRadius" : "40%",
          "maxNumOfSlices" : 10,
          "aggregationType" : "sum"
        },
        "title" : {
          "fontSize" : 14
        },
      }
      )
      keep_visual_settings_consistent_with_parent = true
      query {
        query_string = <<QUERY
_sourceCategory=tsat-batcher | json auto | where msg="finished drift genQuery run for active customers"
| if ( !isNull(duration) and duration > 320 ,1,0) as isGood
| sum(isGood) as totalGood, count as totalCount by _timeslice
QUERY
        query_type   = "Logs"
        query_key    = "A"
      }
      time_range {
        begin_bounded_time_range {
          from {
            literal_time_range {
              range_name = "today"
            }
          }
        }
      }
    }
  }

  ## search panel - multiple metrics queries
  panel {
    sumo_search_panel {
      key                                         = "metrics-panel-02"
      title                                       = "Violations"
      description                                 = "budget left"
      # time series with line of dash dot type
      visual_settings                             = jsonencode(
      {
        "general" : {
          "mode" : "timeSeries",
          "type" : "line",
          "displayType" : "smooth",
          "markerSize" : 5,
          "lineDashType" : "dashDot",
          "markerType" : "none",
          "lineThickness" : 1
        },
        "title" : {
          "fontSize" : 14
        },
      }
      )
      keep_visual_settings_consistent_with_parent = true
      query {
        query_string       = "metric=Proc_CPU nite-api-1"
        query_type         = "Metrics"
        query_key          = "A"
        metrics_query_mode = "Basic"
        metrics_query_data {
          metric           = "Proc_CPU"
          filter {
            key      = "_sourcehost"
            negation = false
            value    = "nite-api-1"
          }
          aggregation_type = "None"
        }
      }
      query {
        query_string       = "metric=Proc_CPU nite-api-2"
        query_type         = "Metrics"
        query_key          = "B"
        metrics_query_mode = "Basic"
        metrics_query_data {
          metric           = "Proc_CPU"
          filter {
            key      = "_sourcehost"
            negation = false
            value    = "nite-api-2"
          }
          aggregation_type = "None"
        }
      }
      time_range {
        begin_bounded_time_range {
          from {
            relative_time_range {
              relative_time = "-1h"
            }
          }
        }
      }
    }
  }

  ## layout
  layout {
    grid {
      layout_structure {
        key       = "text-panel-01"
        structure = "{\"height\":5,\"width\":24,\"x\":0,\"y\":0}"
      }
      layout_structure {
        key       = "search-panel-01"
        structure = "{\"height\":10,\"width\":12,\"x\":0,\"y\":5}"
      }
      layout_structure {
        key       = "metrics-panel-01"
        structure = "{\"height\":10,\"width\":12,\"x\":12,\"y\":5}"
      }
      layout_structure {
        key       = "metrics-panel-02"
        structure = "{\"height\":10,\"width\":24,\"x\":0,\"y\":25}"
      }
    }
  }

}
