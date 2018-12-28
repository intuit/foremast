# Foremast Brain
[![Build Status](https://api.travis-ci.org/intuit/foremast-brain.svg?branch=master)](https://www.travis-ci.org/intuit/foremast-brain)
[![Slack Chat](https://img.shields.io/badge/slack-live-orange.svg)](https://foremastio.slack.com/)

Foremast-brain makes health judgments of [Foremast](https://github.com/intuit/foremast), a service health detection and canary analysis system for Kubernetes. There are two main criteria that Foremast-brain evaluates:

1. Check if the baseline and current health metric have the same distribution pattern.
2. Calculate the historical model and detect current metric anomalies.

Foremast-brain will make a judgment, Healthy or Unhealthy, based on the evaluation result.

Please check out the [architecture and design](https://github.com/intuit/foremast/blob/master/docs/design.md) for more details.

## How to use

### Overwriting default algorithm and parameters

There are multiple sets of parameters that can be overwritten.

#### Machine learning algorithm related parameters -- used for post-deployment use cases

- `ML_ALGORITHM` -- Algorithm which you want to run. Please refer to [AI_MODEL](https://github.com/intuit/foremast-brain/blob/master/src/models/modelclass.py) for all the supported algorithms
- `MIN_HISTORICAL_DATA_POINT_TO_MEASURE` -- Minimum historical data points size
- `ML_BOUND` -- Measurement is upper bound, lower bound or upper and lower bound
- `ML_THRESHOLD` -- Machine learning algorithm threshold

#### Performance, fault-tolerant related parameters

- `MAX_STUCK_IN_SECONDS` -- Max process time until another foremast-brain process will take over and reprocess
- `MAX_CACHE_SIZE` -- Max cached model size

#### Pairwise algorithm parameters -- used for pre-deployment use cases

- `ML_PAIRWISE_ALGORITHM` -- There are multiple options: ALL, ANY, MANN_WHITE, WILCOXON, KRUSKAL, etc.
- `ML_PAIRWISE_THRESHOLD` -- Pairwise algorithm threshold
- `MIN_MANN_WHITE_DATA_POINTS` -- Minimum data points required by Mann-Whitney U algorithm
- `MIN_WILCOXON_DATA_POINTS` -- Minimum data points required by Wilcoxon algorithm
- `MIN_KRUSKAL_DATA_POINTS` -- Minimum data points required by Kruskal algorithm

### How to make changes:

You can add algorithm names and different parameters.
Please refer [foremast-brain](https://github.com/intuit/foremast/blob/master/deploy/foremast/3_judgement/foremast-brain.yaml) for details.

The following is an example of `ES_ENDPOINT`:

```sh
env:
- name: ES_ENDPOINT
  value: "http://elasticsearch-discovery.foremast.svc.cluster.local:9200"
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## License

This project is licensed under the Apache License - see the [LICENSE](LICENSE) file for details
