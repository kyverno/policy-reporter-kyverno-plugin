# Changelog

## 1.2.0

* Support for linux/s390x Archs [[#13](https://github.com/kyverno/policy-reporter-kyverno-plugin/pull/13) by [skuethe](https://github.com/skuethe)]

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