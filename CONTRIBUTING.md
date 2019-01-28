# Contributing guidelines

## How to become a contributor and submit your own code

### Contributor License Agreements

We'd love to accept your patches! Before we can take them, we have to jump a couple of legal hurdles.

Please fill out either the individual or corporate Contributor License Agreement (CLA).

  * If you are an individual writing original source code and you're sure you own the intellectual property, then you'll need to sign an individual CLA.
  * If you work for a company that wants to allow you to contribute your work, then you'll need to sign a corporate CLA.

To sign and submit a CLA, see the [CLA doc](https://git.k8s.io/community/CLA.md).

### Code styles
1. Go lang

   The code style of golang is the default IDE golang code style in [GoLand](https://www.jetbrains.com/go/?fromMenu).
1. Java

   The code style of java is the default IDE java code style in [Intellij IDEA](https://www.jetbrains.com/idea/).

### Projects
1. Foremast-barrelman

   Barrelman is a component to integrate with kubernetes. It watches the changes in K8s and trigger job for foremast-brain.
   
   ```bash
   # how to build a docker image
   cd foremast-barrelman
   make
   docker build .
   ```
2. Foremast-service
   
   Foremast-service is a component to provide APIs of foremast-brain, provide metrics proxy for foremast-ui.
   ```bash
   # how to build and build a docker image
   cd foremast-service
   ./build.sh
   docker build .
   ```
3. Foremast-spring-boot-k8s-metrics-starter
   
   Foremast-spring-boot-k8s-metrics-starter is a component to help users to enable the required metrics in their spring-boot applications.
   ```bash
   # How to make build
   mvn clean install
   ```
4. Foremast-UI
   
   [How to build](https://github.com/intuit/foremast/blob/master/app/README.md)
   
### Contributing A Patch

1. Submit an issue describing your proposed change to the repo in question.
1. The [repo owners](OWNERS) will respond to your issue promptly.
1. If your proposed change is accepted, and you haven't already done so, sign a Contributor License Agreement (see details above).
1. Fork the desired repo, develop and test your code changes.
1. Submit a pull request.

### Submitting a PR

- Fork the repo.

Do changes in your local repo, before committing your changes, make sure the compilation is green.

- Create your PR.

- Tests will automatically run for you.

- We will not merge any PR that is not passing tests.

- Any PR that changes user-facing behavior must have associated documentation in docs as well as release notes. API changes should be documented inline with protos as per the API contribution guidelines.

- Your PR title should be descriptive, and generally start with an existing issue. Examples:

  - "#XXX Some changes in some components"

- Your PR description should have details on what the PR does. If it fixes an existing issue it should end with "#XXX Fixes for some bugs".

- When all of the tests are passing and all other conditions described herein are satisfied, a maintainer will be assigned to review and merge the PR.

- Once you submit a PR, please do not rebase it. It's much easier to review if subsequent commits are new commits and/or merges. We squash rebase the final merged commit so the number of commits you have in the PR don't matter.

- We expect that once a PR is opened, it will be actively worked on until it is merged or closed. We reserve the right to close PRs that are not making progress. This is generally defined as no changes for 7 days. Obviously PRs that are closed due to lack of activity can be reopened later. Closing stale PRs helps us to keep on top of all of the work currently in flight.

- If a commit deprecates a feature, the commit message must mention what has been deprecated. Additionally, DEPRECATED.md must be updated as part of the commit.
er-example (for example making a new branch so that CI can pass) it is your responsibility to follow through with merging those changes back to master once the CI dance is done.

### Release cadence

- Currently we are targeting approximately quarterly official releases. We may change this based on customer demand.
- In general, master is assumed to be release candidate quality at all times for documented features. For undocumented or clearly under development features, use caution or ask about current status when running master. 
- Note that we currently do not provide formal docker images. Organizations are expected to build Foremast from source. This may change in the future if we get resources for maintaining packages.