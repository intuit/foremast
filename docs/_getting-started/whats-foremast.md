## What's Foremast?

<div style="text-align:center; margin:30px 0">
  <img src="{{ site.baseurl }}/assets/images/foremast-logo.png" class="img-fluid" alt="Foremast Logo">
</div>

An application’s health and its ability to serve customers is of primary importance for any business. Kubernetes provides elegant abstractions that ensure applications are resilient to infrastructure changes and failures. Foremast adds a layer of application resiliency to Kubernetes. By leveraging machine learning on application metrics and data, infrastructure data and various other data sources, Foremast provides intelligent observability in order to maintain application health during deployments and in steady state operations.

Foremast is an early warning system for detecting problems with the deployment of a new version of a service or component. Production deployments have used manual canary analysis for a few years now in various forms, be it A/B testing, phased rollout, or incremental rollout.

Foremast enables automated canary analysis that scores the health of new deployments on the basis of performance, functionality, and quality. In the case of rolling updates, the analysis should also be performed for the cluster as a whole to confirm the success of the upgrade for the whole application.

It addresses following problems in an enterprise environment of Kubernetes:

- Detect metrics spike or drop due to a deployment
- Detect impact to downstream services
- Automated remediation including alert, rollback etc
- Metrics anomaly Aggregated at service or API level
- Aggregate service health check across multiple K8s clusters

Check out the [architecture and design](https://github.intuit.com/dev-containers/foremast/blob/master/docs/design.md).