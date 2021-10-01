package libs

const gaugeQueryPart = `
| sum(sliceGoodCount) as totalGood, sum(sliceTotalCount) as totalCount
| (totalGood/totalCount)*100 as SLO | format("%.2f%%",SLO) as sloStr
| fields SLO
`

const hourlyBurnQueryPart = `
| timeslice 60m 
| sum(sliceGoodCount) as tmGood, sum(sliceTotalCount) as tmCount  group by _timeslice
| tmGood/tmCount as tmSLO 
| (tmCount-tmGood) as tmBad 
| order by _timeslice asc
| accum tmBad as runningBad  
| total tmCount as totalCount  
| totalCount*(1-0.9) as errorBudget
| (1-runningBad/errorBudget) as budgetRemaining 
| ((tmBad/tmCount)/(1-0.9)) as hourlyBurnRate
| fields _timeslice, hourlyBurnRate | compare timeshift 1d
`

const burnTrendQueryPart = `
| sum(sliceGoodCount) as totalGood, sum(sliceTotalCount) as totalCount 
| (1 - totalGood/totalCount)/0.1 as BurnRate | fields BurnRate | compare timeshift 1d 7 
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
| totalCount*(1-0.9) as errorBudget
| (1-runningBad/errorBudget) as budgetRemaining 
| ((tmBad/tmCount)/(1-0.9)) as hourlyBurnRate
| fields _timeslice, budgetRemaining
`
