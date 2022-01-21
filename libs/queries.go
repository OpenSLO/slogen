package libs

const gaugeQueryPartForOccurrences = `
| sum(sliceGoodCount) as totalGood, sum(sliceTotalCount) as totalCount
| (totalGood/totalCount)*100 as SLO | format("%.2f%%",SLO) as sloStr
| fields SLO
`

const gaugeQueryPartForTimeslice = `
| timeslice 1m 
| sum(sliceGoodCount) as timesliceGoodCount, sum(sliceTotalCount) as timesliceTotalCount by _timeslice
| fillmissing timeslice(1m)
| if(timesliceTotalCount ==0, 1,(timesliceGoodCount/timesliceTotalCount))  as timesliceRatio
| order by _timeslice asc
| if(timesliceRatio >= {{.TimesliceRatioTarget}}, 1,0) as sliceHealthy
| 1 as timesliceOne
| sum(sliceHealthy) as healthySlices, sum(timesliceOne) as totalSlices
| (healthySlices/totalSlices)*100 as Availability
| fields Availability
`

const hourlyBurnQueryPartForOccurrences = `
| timeslice 60m 
| sum(sliceGoodCount) as tmGood, sum(sliceTotalCount) as tmCount  group by _timeslice
| tmGood/tmCount as tmSLO 
| (tmCount-tmGood) as tmBad 
| order by _timeslice asc
| total tmCount as totalCount  
| ((tmBad/tmCount)/(1-{{.Target}})) as hourlyBurnRate
| fields _timeslice, hourlyBurnRate | compare timeshift 1d
`

const hourlyBurnQueryPartForTimeslice = `
| timeslice 1m 
| sum(sliceGoodCount) as timesliceGoodCount, sum(sliceTotalCount) as timesliceTotalCount by _timeslice
| fillmissing timeslice(1m)
| if(timesliceTotalCount ==0, 1,(timesliceGoodCount/timesliceTotalCount))  as timesliceRatio
| if(timesliceRatio >= {{.TimesliceRatioTarget}}, 1,0) as sliceHealthy
| 1 as timesliceOne | _timeslice as _messagetime 
| timeslice 60m 
| sum(sliceHealthy) as healthySlices, sum(timesliceOne) as totalSlices by _timeslice
| order by _timeslice asc
| ((1 - healthySlices/totalSlices)/(1-{{.Target}})) as hourlyBurnRate
| fields _timeslice, hourlyBurnRate | compare timeshift 1d
`

const burnTrendQueryPartForOccurrences = `
| sum(sliceGoodCount) as totalGood, sum(sliceTotalCount) as totalCount 
| ((1 - totalGood/totalCount)/(1-{{.Target}}))*100 as BurnRate | fields BurnRate | compare timeshift 1d 7 
| fields BurnRate_7d,BurnRate_6d,BurnRate_5d,BurnRate_4d,BurnRate_3d,BurnRate_2d,BurnRate_1d,BurnRate
`

const burnTrendQueryPartForTimeslice = `
| timeslice 1m 
| sum(sliceGoodCount) as timesliceGoodCount, sum(sliceTotalCount) as timesliceTotalCount by _timeslice
| fillmissing timeslice(1m)
| if(timesliceTotalCount ==0, 1,(timesliceGoodCount/timesliceTotalCount))  as timesliceRatio
| if(timesliceRatio >= {{.TimesliceRatioTarget}}, 1,0) as sliceHealthy
| 1 as timesliceOne
| sum(sliceHealthy) as totalGood, sum(timesliceOne) as totalCount 
| ((1 - totalGood/totalCount)/(1-{{.Target}}))*100 as BurnRate | fields BurnRate | compare timeshift 1d 7 
| fields BurnRate_7d,BurnRate_6d,BurnRate_5d,BurnRate_4d,BurnRate_3d,BurnRate_2d,BurnRate_1d,BurnRate
`

const budgetLeftQueryPart = `
| timeslice 60m 
| sum(sliceGoodCount) as tmGood, sum(sliceTotalCount) as tmCount  group by _timeslice
| tmGood/tmCount as tmSLO 
| (tmCount-tmGood) as tmBad 
| order by _timeslice asc
| accum tmBad as runningBad  
| total tmCount as totalCount  
| totalCount*(1-{{.Target}}) as errorBudget
| (1-runningBad/errorBudget) as budgetRemaining 
| fields _timeslice, budgetRemaining
| predict budgetRemaining by 1h model=ar, forecast=800
| toLong(formatDate(_timeslice, "M")) as tmIndex 
| toLong(formatDate(now(), "M")) as monthIndex
| where  tmIndex = monthIndex | if(isNull(budgetRemaining),budgetRemaining_predicted,budgetRemaining) as budgetRemaining_predicted 
| fields _timeslice,budgetRemaining, budgetRemaining_predicted 
`

const budgetLeftQueryTimeSlicesPart = `
| timeslice 1m 
| sum(sliceGoodCount) as timesliceGoodCount, sum(sliceTotalCount) as timesliceTotalCount by _timeslice
| fillmissing timeslice(1m)
| if(timesliceTotalCount ==0, 1,(timesliceGoodCount/timesliceTotalCount))  as timesliceRatio
| (timesliceGoodCount/timesliceTotalCount)  as timesliceRatio
| if(timesliceRatio >= {{.TimesliceRatioTarget}}, 1,0) as sliceHealthy
| 1 as timesliceOne | _timeslice as _messagetime 
| timeslice 60m 
| sum(sliceHealthy) as healthySlices, sum(timesliceOne) as totalSlices by _timeslice
| healthySlices/totalSlices as tmSLO 
| (totalSlices - healthySlices) as tmBad 
| order by _timeslice asc
| accum tmBad as runningBad  
| toLong(formatDate(now(), "M")) as monthIndex
| if(monthIndex == 12,0,1) as addToMonth 
| parseDate(format("2021-%d-01",toLong(monthIndex)), "yyyy-MM-dd") as ym
| parseDate(format("2021-%d-01",toLong(monthIndex+addToMonth)), "yyyy-MM-dd") as ymNext
| toLong(if(monthIndex == 12,31,(ymNext - ym)/(24*3600*1000))) as dayCount
| (dayCount*24*60)*(1-{{.Target}}) as errorBudget
| (1-runningBad/errorBudget) as budgetRemaining 
| fields _timeslice, budgetRemaining
| predict budgetRemaining by 1h model=ar, forecast=800
| toLong(formatDate(_timeslice, "M")) as tmIndex 
| toLong(formatDate(now(), "M")) as monthIndex
| where  tmIndex = monthIndex | if(isNull(budgetRemaining),budgetRemaining_predicted,budgetRemaining) as budgetRemaining_predicted 
| fields _timeslice,budgetRemaining, budgetRemaining_predicted 
`

const breakDownPanelQueryOccurrences = `
| sum(sliceGoodCount) as totalGood, sum(sliceTotalCount) as totalCount {{if ne .GroupByStr ""}}by {{.GroupByStr}} {{end}} 
| totalCount - totalGood as totalBad
| (totalGood/totalCount)*100 as Availability_Percentage
| totalCount*(1-{{.Target}}) as errorBudget
| (1-totalBad/errorBudget) as BudgetRemaining 
| BudgetRemaining*100 as %"Budget Remaining (%)"
| order by BudgetRemaining asc
| Availability_Percentage as %"Availability (%)" 
| fields {{if ne .GroupByStr ""}} {{.GroupByStr}}, {{end}} %"Availability (%)", %"Budget Remaining (%)"
`

const breakDownPanelQueryTimeslices = `
| timeslice 1m 
| sum(sliceGoodCount) as timesliceGoodCount, sum(sliceTotalCount) as timesliceTotalCount by _timeslice {{if ne .GroupByStr ""}}, {{.GroupByStr}} {{end}} 
| fillmissing timeslice(1m)
| if(timesliceTotalCount ==0, 1,(timesliceGoodCount/timesliceTotalCount))  as timesliceRatio
| (timesliceGoodCount/timesliceTotalCount)  as timesliceRatio
| if(timesliceRatio >= 0.9, 1,0) as sliceHealthy
| 1 as timesliceOne | _timeslice as _messagetime 
| sum(sliceHealthy) as healthySlices, sum(timesliceOne) as totalSlices {{if ne .GroupByStr ""}}by {{.GroupByStr}} {{end}} 
| (healthySlices/totalSlices)*100 as Availability_Percentage
| (totalSlices - healthySlices) as badSlices
| toLong(formatDate(now(), "M")) as monthIndex
| if(monthIndex == 12,0,1) as addToMonth 
| parseDate(format("2021-%d-01",toLong(monthIndex)), "yyyy-MM-dd") as ym
| parseDate(format("2021-%d-01",toLong(monthIndex+addToMonth)), "yyyy-MM-dd") as ymNext
| toLong(if(monthIndex == 12,31,(ymNext - ym)/(24*3600*1000))) as dayCount
| (dayCount*24*60)*(1-0.95) as errorBudget
| (1-badSlices/errorBudget) as BudgetRemaining 
| (BudgetRemaining)*errorBudget as DowntimeRemainingInMinutes 
| DowntimeRemainingInMinutes/60 as DowntimeRemainingInHours 
| DowntimeRemainingInMinutes%60 as DowntimeRemainingMinuteModulo 
| format("%2.0fh%2.0fm",DowntimeRemainingInHours,DowntimeRemainingMinuteModulo) as %"Budget Remaining (Time)"
| BudgetRemaining*100 as %"Budget Remaining (%)"
| order by BudgetRemaining asc
| Availability_Percentage as %"Availability (%)" 
| fields {{if ne .GroupByStr ""}} {{.GroupByStr}}, {{end}} %"Availability (%)", %"Budget Remaining (%)", %"Budget Remaining (Time)"
`

const pd = `
_view=slogen_tf_cloudcollector_cc_ingest_lag_v2 
| timeslice 1m 
| sum(sliceGoodCount) as timesliceGoodCount, sum(sliceTotalCount) as timesliceTotalCount by _timeslice
| fillmissing timeslice(1m)
| if(timesliceTotalCount ==0, 1,(timesliceGoodCount/timesliceTotalCount))  as timesliceRatio
| (timesliceGoodCount/timesliceTotalCount)  as timesliceRatio
| if(timesliceRatio >= 0.9, 1,0) as sliceHealthy
| 1 as timesliceOne | _timeslice as _messagetime 
| timeslice 60m 
| toLong(formatDate(now(), "M")) as monthIndex
| if(monthIndex == 12,0,1) as addToMonth 
| parseDate(format("2021-%d-01",toLong(monthIndex)), "yyyy-MM-dd") as ym
| parseDate(format("2021-%d-01",toLong(monthIndex+addToMonth)), "yyyy-MM-dd") as ymNext
| toLong(if(monthIndex == 12,31,(ymNext - ym)/(24*3600*1000))) as dayCount
| sum(sliceHealthy) as healthySlices, sum(timesliceOne) as totalSlices by _timeslice | predict healthySlices by 1d forecast=30 
| if(isNull(healthySlices) ,healthySlices_predicted,healthySlices) as forecasted_slices | fields _timeslice,healthySlices,forecasted_slices
`

/*
_view=slogen_tf_tsat_v2_anomaly_compute_delay_v2
| timeslice 1m
| fillmissing timeslice(1m)
| sum(sliceGoodCount) as timesliceGoodCount, sum(sliceTotalCount) as timesliceTotalCount by _timeslice
| if(timesliceTotalCount ==0, 1,(timesliceGoodCount/timesliceTotalCount))  as timesliceRatio
| if(timesliceRatio >= 0.9, 1,0) as sliceHealthy | _timeslice as _messagetime
| 1 as timesliceOne
| timeslice 60m
| sum(sliceHealthy) as healthySlices, sum(timesliceOne) as totalSlices by _timeslice
| healthySlices/totalSlices as tmSLO
| (totalSlices - healthySlices) as badSlices
| order by _timeslice asc
| accum badSlices as runningBad
| toLong(formatDate(now(), "M")) as monthIndex
| if(monthIndex == 12,0,1) as addToMonth
| parseDate(format("2021-%d-01",toLong(monthIndex)), "yyyy-MM-dd") as ym
| parseDate(format("2021-%d-01",toLong(monthIndex+addToMonth)), "yyyy-MM-dd") as ymNext
| toLong(if(monthIndex == 12,31,(ymNext - ym)/(24*3600*1000))) as dayCount
| (dayCount*24*60)*(1-0.8) as errorBudget
| (1-runningBad/errorBudget) as budgetRemaining
| fields _timeslice, budgetRemaining

_view=slogen_tf_cloudcollector_cc_ingest_lag_v2
| timeslice 1m
| sum(sliceGoodCount) as timesliceGoodCount, sum(sliceTotalCount) as timesliceTotalCount by _timeslice
| fillmissing timeslice(1m)
| if(timesliceTotalCount ==0, 1,(timesliceGoodCount/timesliceTotalCount))  as timesliceRatio
| (timesliceGoodCount/timesliceTotalCount)  as timesliceRatio
| if(timesliceRatio >= 0.9, 1,0) as sliceHealthy
| 1 as timesliceOne | _timeslice as _messagetime
| timeslice 60m
| sum(sliceHealthy) as healthySlices, sum(timesliceOne) as totalSlices by _timeslice | predict healthySlices by 1d forecast=30 |  toLong(formatDate(now(), "M")) as monthIndex | toLong(formatDate(now(), "M")) as dayIndex
| if(monthIndex == 12,0,1) as addToMonth
| parseDate(format("2021-%d-01",toLong(monthIndex)), "yyyy-MM-dd") as ym
| parseDate(format("2021-%d-01",toLong(monthIndex+addToMonth)), "yyyy-MM-dd") as ymNext
| toLong(if(monthIndex == 12,31,(ymNext - ym)/(24*3600*1000))) as dayCount
| if(isNull(healthySlices) ,healthySlices_predicted,healthySlices) as forecasted_slices | formatDate(_timeslice,"dd") as m | where m = dayCount
*/
