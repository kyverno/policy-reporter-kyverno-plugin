# Changelog

## 1.6.0

* Support BasicAuth for REST APIs and metrics
* Update to Go 1.21

## 1.5.1

* Add zap.Logger

## 1.5.0

* Implement Kyverno Compliance Reports
* Update to Go 1.19

## 1.4.2

* Block Reports: check for duplicated event IDs
* Go depdency updates

## 1.4.1

* Block Reports: add time to result properties

## 1.4.0 

* Implement LeaderElection for PolicyReport Management
* Refactor K8s Clients
* Update Depdencies and Workflow to Go 1.18

## 1.2.1 

* Bump Go Verstion to Go 1.17.8

## 1.2.0

* Support for linux/s390x Archs [[#13](https://github.com/kyverno/policy-reporter-kyverno-plugin/pull/13) by [skuethe](https://github.com/skuethe)]

## 1.1.1

* Bump Go Verstion to Go 1.17.6 [[#12](https://github.com/kyverno/policy-reporter-kyverno-plugin/pull/12) by [realshuting](https://github.com/realshuting)]

## 1.1.0

* Add VerifyImages information to `/policy` API
* Add new `verify-image-rules` API with a list of VerifyImages Rule definitions

## 1.0.0

* Use one HTTP Server with Port `8080` for Metrics and REST API
* Use informer instead of watch
* Improve Pub/Sub pattern
* Improve Healthz API

## 0.1.0

* Prometheus Metrics for active Policies
* Policy REST API