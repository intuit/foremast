# Dynamic Auto-Scaling

## How it Works

### Flow
```
┌────────────────────────┐                                                       ┌──────────────────────┐
│                        │                                                       │ Score < 50 for Scale │
│     Determine TPS      │                                                       │         DOWN         │
│  seasonality & trend   │                                                       │ Score > 50 for Scale │
│                        │                                                       │          UP          │
└────────────────────────┘                                                       └──────────────────────┘
             │                                                                               ▲
             │                                     ┌────────────────┐                        │
             │                                     │ Use predicted  │                        │
             │                                 ┌──▶│     trend      │──────┐                 │
             ▼                                 │   └────────────────┘      │                 │
┌────────────────────────┐                     │                           ▼                 │
│                        │                     │              ┌─────────────────────────┐    │
│     Calculate TPS      │                    No              │      Reward function:   │    │
│ upper-lower bound for  │────▶  Anomaly?   ───┤              │     R(action) based on  │────┘
│       every 30m        │                     │              │        SLA metrics      │
│                        │                     │              └─────────────────────────┘
└────────────────────────┘                   Yes                           ▲
                                               │   ┌────────────────┐      │
                                               │   │  Use detected  │      │
                                               └──▶│trend w/ cycles │──────┘
                                                   └────────────────┘
```
### Steps
- Define a reward function __R(action)__ for application, using by default
`R(action) = M(latency) < SLA ? (action==UP?-1,1), (action==UP?1,-1)`:
  - Use a static criteria for `M(latency)`
  - Use dymanic criteria for `M(latency)`: 3 sigma of standard deviations
  - Use Min of above two
- Build model for historic `M(tps)` with detected seasnality, and trending.
- Every *PT* mins, Foremast predicts `M(tps)` for *PT* mins using derived model
- If real value of `M(tps)` falls in predicted range, use predicted trending as `Tr(tps)`.
Otherwise, observe *N* data points to determine anomaly trending as `Tr(tps)`
- Determine scaling recommendation:
  - If `Tr(tps)` is UP, return a score > 50 when __R(UP)__ == 1
  - If `Tr(tps)` is DOWN, return a score < 50 when __R(DOWN)__ == 1

### Reward Function
The default reward function R(action) considers two conditions:
1. SLA metrics
2. Resource allocation

If a SLA metrics is not provided, *Latency* will be used.
The rule is simple - reward function will reward lower resource allocation as long as SLA metrics is not vialated.

| R(action)  | Scale UP  | Scale DOWN  |
|---|---|---|
| Meet SLA | -1 | 1 |
| Not Meet SLA  | 1 | -1 |


## Samples

### A seasonal traffic pattern with noise
```
11.90  ┼╭─╮                         ╭╮ ╭╮                         ╭╮ ╭╮                        ╭╮
10.91  ┤│ │                         ││╭╯│                        ╭╯│╭╯│                        ││
 9.92  ┼╯ ╰╮                    ╭─╮╭╯╰╯ │                     ╭─╮│ ╰╯ │                      ╭─╯╰
 8.93  ┤   │╭─╮                 │ ╰╯    ╰╮ ╭╮                 │ ╰╯    ╰─╮╭╮                 ╭╯
 7.93  ┤   ╰╯ │               ╭╮│        ╰─╯╰╮               ╭╯         ╰╯│                ╭╯
 6.94  ┤      ╰╮              │╰╯            ╰╮             ╭╯            ╰╮               │
 5.95  ┤       ╰╮             │               │             │              ╰╮             ╭╯
 4.96  ┤        ╰╮           ╭╯               │          ╭╮╭╯               │            ╭╯
 3.97  ┤         │        ╭──╯                │╭─╮      ╭╯││                ╰╮╭╮      ╭╮ │
 2.97  ┤         ╰╮       │                   ╰╯ │      │ ╰╯                 ╰╯│      │╰─╯
 1.98  ┤          ╰╮   ╭╮╭╯                      │╭─╮ ╭╮│                      │╭╮   ╭╯
 0.99  ┤           ╰──╮│╰╯                       ╰╯ │ │╰╯                      ╰╯│  ╭╯
 0.00  ┤              ╰╯                            ╰─╯                          ╰──╯
```
A predicted trend will be used and based on reward function, it will return <50 score to scale down
whenever it sees downwards trending until it ensures the full utilization of resources.

Noise as long as not exceeding bound of the prediction, will be ignored.


### A seasonal traffic pattern with anomaly
```
10.00  ┼───╮                                     ╭──────╮                                     ╭──
 9.05  ┼   ╰─╮                                 ╭─╯      ╰─╮                                 ╭─╯
 8.10  ┤     ╰─╮                               │          ╰─╮                             ╭─╯
 7.15  ┤       ╰─╮                             │            ╰─╮                         ╭─╯
 6.20  ┤         ╰╮                       ╭╮   │              ╰╮                       ╭╯
 5.25  ┤          ╰─╮                   ╭─╯│   │               ╰─╮                   ╭─╯
 4.30  ┤            ╰╮                 ╭╯  │   │                 ╰╮                 ╭╯
 3.35  ┤             ╰─╮             ╭─╯   │╭╮╭╯                  ╰─╮             ╭─╯
 2.40  ┤               ╰╮           ╭╯     ││╰╯                     ╰╮           ╭╯
 1.45  ┤                ╰──╮     ╭──╯      ││                        ╰──╮     ╭──╯
 0.50  ┤                   ╰─────╯         ╰╯                           ╰─────╯

```
A predicted trend will be used for most of the time. Anomaly in middle will be detected as it
exceeds the lower bound of prediction. In that case, it will observe several cycles to ensure the
downwards trending and kick off scale down. However, it will also detect the fast recovery of traffic
and decides to scale up if SLA is impacted.

### An anomaly without seasonality
```
10.00  ┤                    ╭╮     ╭╮
 9.00  ┤                    ││  ╭╮ ││
 8.00  ┤                    ││  ││ ││   ╭─╮
 7.00  ┤                    ││  ││ ││ ╭╮│ │
 6.00  ┤                    │╰╮ ││ ││ │││ │   ╭─╮
 5.00  ┤                    │ ╰╮│╰╮││ │││ │   │ │
 4.00  ┤                    │  ││ │││╭╯││ ╰╮  │ │
 3.00  ┤                    │  ││ ╰╯││ ╰╯  │  │ │
 2.00  ┼──╮╭─╮╭╮  ╭╮╭─╮ ╭╮╭╮│  ╰╯   ││     │╭─╯ │╭╮   ╭╮╭╮   ╭╮ ╭─╮╭──╮╭╮  ╭╮╭─╮╭╮ ╭──╮╭╮ ╭─╮
 1.00  ┤  ╰╯ ╰╯╰──╯╰╯ ╰─╯╰╯╰╯       ╰╯     ╰╯   ╰╯╰───╯╰╯╰───╯╰─╯ ╰╯  ╰╯╰──╯╰╯ ╰╯╰─╯  ╰╯╰─╯ ╰────
```
Anomaly will be detected as a upwards trend or downwards trend. Upwards will only trigger scale up when it sees SLA impact.

## Breath duration
Breath duration is a period of time left to give the system a chance to reach a
stable state after each scaling decision:
- Breath-up is time to wait after each scaling up decision.
- Breath-down is time to wait after each scaling down decision.

Breath-up is smaller than breath-down to allow the system to scale-up rapidly to 
cope with burst workload. Breath-down is larger than breath-up duration in order
to avoid uncertain scaling down which could cause degradation in the system.

## Feedback
Some metrics will be used to feedback into the flow:
- Number of Scale up/down decisions in last N minutes to avoid flipping
- ...
