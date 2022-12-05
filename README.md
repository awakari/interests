# Contents

1. [Overview](#1-overview)<br/>
   1.1. [Purpose](#11-purpose)<br/>
   1.2. [Definitions](#12-definitions)<br/>
   &nbsp;&nbsp;&nbsp;1.2.1. [Pattern](#121-pattern)<br/>
   &nbsp;&nbsp;&nbsp;1.2.2. [Matchers](#122-matcher)<br/>
   &nbsp;&nbsp;&nbsp;1.2.3. [Matcher Group](#123-matcher-group)<br/>
   &nbsp;&nbsp;&nbsp;1.2.4. [Subscription](#124-subscription)<br/>
2. [Configuration](#2-configuration)<br/>
3. [Deployment](#3-deployment)<br/>
   3.1. [Prerequisites](#31-prerequisites)<br/>
   3.2. [Bare](#32-bare)<br/>
   3.3. [Docker](#33-docker)<br/>
   3.4. [K8s](#34-k8s)<br/>
   &nbsp;&nbsp;&nbsp;3.4.1. [Helm](#341-helm) <br/>
4. [Usage](#4-usage)<br/>
5. [Design](#5-design)<br/>
   5.1. [Requirements](#51-requirements)<br/>
   5.2. [Approach](#52-approach)<br/>
   &nbsp;&nbsp;&nbsp;5.2.1. [Data Schema](#521-data-schema)<br/>
   &nbsp;&nbsp;&nbsp;5.2.2 [Results Pagination](#522-results-pagination)<br/>
   5.3. [Limitations](#53-limitations)<br/>
6. [Contributing](#6-contributing)<br/>
   6.1. [Versioning](#61-versioning)<br/>
   6.2. [Issue Reporting](#62-issue-reporting)<br/>
   6.3. [Building](#63-building)<br/>
   6.4. [Testing](#64-testing)<br/>
   &nbsp;&nbsp;&nbsp;6.4.1. [Functional](#641-functional)<br/>
   &nbsp;&nbsp;&nbsp;6.4.2. [Performance](#642-performance)<br/>
   6.5. [Releasing](#65-releasing)<br/>

# 1. Overview

## 1.1. Purpose

The main function is to find all subscriptions by a linked wildcard. Wildcard may be found by the sample text input 
using [matchers](https://github.com/meandros-messaging/matchers) service, hence, it makes possible to resolve all 
matching wildcard subscriptions.

## 1.2. Definitions

### 1.2.1. Pattern

See the [definition here](https://github.com/meandros-messaging/matchers#122-pattern).

### 1.2.2. Matcher

Same as [matcher](https://github.com/meandros-messaging/matchers#123-matcher) but with an additional `partial` flag.
The `partial` flag controls whether the matcher should be used to match the sample text parts matching or complete only.

### 1.2.3. Matcher Group

Matcher group is just a set of matchers with an additional `all` flag. This `all` flag controls whether all matchers in
this group should match the sample input or any is enough. 

### 1.2.4. Subscription

Subscription is a named bundle of matchers with assigned routes and human-readable description. Matchers are listed in 
two matcher groups:
* `Includes`: matchers those should match the input data, all or any (matcher group attribute), completely or partially (matcher attribute).
* `Excludes`: matchers those should not match the input data, all or any, completely or partially.
These two matcher groups control whether the entire subscription matches the input data or not.

# 2. Configuration

The service is configurable using the environment variables:

| Variable                           | Example value                                          | Description                                                                                          |
|------------------------------------|--------------------------------------------------------|------------------------------------------------------------------------------------------------------|
| API_PORT                           | `8080`                                                 | gRPC API port                                                                                        |
| DB_URI                             | `mongodb+srv://localhost/?retryWrites=true&w=majority` | DB connection URI                                                                                    |
| DB_NAME                            | `subscriptions`                                        | DB name to store the data                                                                            |
| DB_TABLE_NAME                      | `subscriptions`                                        | DB table name to store the tree data                                                                 |
| API_MATCHERS_URI_EXCLUDES_COMPLETE | `matchers-excludes-complete:8080`                      | Excluding complete [matchers](https://github.com/meandros-messaging/matchers) dependency service URI |
| API_MATCHERS_URI_EXCLUDES_PARTIAL  | `matchers-excludes-partial:8080`                       | Excluding partial [matchers](https://github.com/meandros-messaging/matchers) dependency service URI  |
| API_MATCHERS_URI_INCLUDES_COMPLETE | `matchers-includes-complete:8080`                      | Including complete [matchers](https://github.com/meandros-messaging/matchers) dependency service URI |
| API_MATCHERS_URI_INCLUDES_PARTIAL  | `matchers-includes-partial:8080`                       | Including partial [matchers](https://github.com/meandros-messaging/matchers) dependency service URI  |

# 3. Deployment

## 3.1. Prerequisites

A general note is that there should be a MongoDB cluster deployed to be used for storing the pattern data.
It's possible to obtain a free cluster for testing purposes using [Atlas](https://www.mongodb.com/atlas/database).

## 3.2. Bare

Preconditions:
1. Build patterns executive using ```make build```
2. Run the [patterns](https://github.com/meandros-messaging/matchers) dependency services (x4: includes/excludes, complete/partial)

Then run the command:
```shell
API_PORT=8081 \
DB_URI=mongodb+srv://localhost/?retryWrites=true&w=majority \
DB_NAME=subscriptions \
DB_TABLE_NAME=subscriptions \
API_MATCHERS_URI_EXCLUDES_COMPLETE=http://localhost:8081 \
API_MATCHERS_URI_EXCLUDES_PARTIAL=http://localhost:8082 \
API_MATCHERS_URI_INCLUDES_COMPLETE=http://localhost:8083 \
API_MATCHERS_URI_INCLUDES_PARTIAL=http://localhost:8084 \
./subscriptions
```

## 3.3. Docker

TODO: run the matchers (x4) and subscriptions in the same network

alternatively, it's possible to build and run the new docker image in place using the command:
(note that the command below requires all env vars to be set in the file `env.txt`)

```shell
make run
```

## 3.4. K8s

TODO

### 3.4.1. Helm

Create a helm package from the sources:
```shell
helm package helm/subscriptions/
```

Install the helm chart:
```shell
helm install subscriptions ./subscriptions-<CHART_VERSION>.tgz \
  --values helm/subscriptions/values-db-uri.yaml
```

where
* `values-db-uri.yaml` contains the value override for the DB URI
* `<CHART_VERSION>` is the helm chart version

# 4. Usage

The service provides basic gRPC interface to perform the operation on subscriptions.
There's a [Kreya](https://kreya.app/) gRPC client application that can be used for the testing purpose.

TODO: screenshot

The service provides few basic operations on subscriptions.

TODO: operations subsections

# 5. Design

## 5.1. Requirements

| #     | Summary          | Description                                                                         |
|-------|------------------|-------------------------------------------------------------------------------------|
| REQ-1 | Basic matching   | Resolve subscriptions matching the input value                                      |
| REQ-2 | Logic            | Support subscription logics for the multiple key-value matches (*and*, *or*, *not*) |
| REQ-3 | Partial matching | Split input metadata values to lexemes and match against each lexeme                |
| REQ-4 | Pagination       | Support query results pagination                                                    |

## 5.2. Approach

### 5.2.1. Data Schema

Subscriptions are stored in the single table under the denormalized schema.

Example data:

```yaml
- name: subscription0
  description: Anything related to orders that are not in Helsinki
  routes:
  - devnull
  includes:
    all: false
    matchers:
    - key: subject
      pattern:
        code: orders
        regex: orders
      partial: true
  excludes:
    all: false
    matchers:
    - key: location
      pattern:
        code: Helsinki
        regex: Helsinki 
      partial: false
```

```yaml
- name: subscription1
  description: Messages that have both high priority and reply-to address of John Doe
  routes:
  - john_doe
  - important
  includes:
    all: true
    matchers:
    - key: reply-to
      pattern: 
        code: john.doe@email.com
        regex: john.doe@email.com
      partial: false
    - key: priority
      pattern: 
        code: high
        regex: high
      partial: false
  excludes: {}
```

#### 5.2.1.1. Subscription

| Attribute   | Type                                 | Description                                                  |
|-------------|--------------------------------------|--------------------------------------------------------------|
| name        | String                               | Unique subscription name                                     |
| description | String                               | Human readable subscription description                      |
| routes      | Array of String                      | Destination routes to use for the matching messages delivery |
| includes    | [Matcher Group](#5212-matcher-group) | Matchers to include the subscription to query results        |
| excludes    | [Matcher Group](#5212-matcher-group) | Matchers to exclude the subscription from the query results  |

#### 5.2.1.2. Matcher Group

| Attribute | Type                              | Description                                                                         |
|-----------|-----------------------------------|-------------------------------------------------------------------------------------|
| all       | Boolean                           | Defines whether **all** matchers in the group should match or **any** is sufficient |
| matchers  | Array of [Matcher](#5213-matcher) | Set of matchers in the group                                                        |

#### 5.2.1.3. Matcher

| Attribute | Type                     | Description                                                                                                   |
|-----------|--------------------------|---------------------------------------------------------------------------------------------------------------|
| key       | String                   | Metadata key                                                                                                  |
| pattern   | [Pattern](#5214-pattern) | Metadata value matching pattern                                                                               |
| partial   | Boolean                  | If `true`, then allowed match any lexeme in a tokenized metadata value. Otherwise, entire value should match. |

#### 5.2.1.4. Pattern

| Attribute | Type          | Description                                                                                            |
|-----------|---------------|--------------------------------------------------------------------------------------------------------|
| code      | Array of byte | Unique pattern path in the [patterns tree](https://github.com/meandros-messaging/patterns#52-approach) |
| regex     | String        | A regular expression to finally filter the resolved subscription candidates                            |

### 5.2.2. Results Pagination

The limit and cursor search parameters are used to support the results' pagination.

## 5.3. Limitations

| #     | Summary                        | Description                                                                                                                                   |
|-------|--------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| LIM-1 | Excluding only is not allowed  | A subscription should have at least 1 in `includes` matcher group. Otherwise the subscription never matches anything in practice.             |

# 6. Contributing

## 6.1. Versioning

The service uses the [semantic versioning](http://semver.org/).
The single source of the version info is the git tag:
```shell
git describe --tags --abbrev=0
```

## 6.2. Issue Reporting

TODO

## 6.3. Building

```shell
make build
```
Generates the sources from proto files, compiles and creates the `patterns` executable.

## 6.4. Testing

### 6.4.1. Functional

```shell
make test
```

### 6.4.2. Performance

TODO

## 6.5. Releasing

To release a new version (e.g. `1.2.3`) it's enough to put a git tag:
```shell
git tag -v1.2.3
git push --tags
```

The corresponding CI job is started to build a docker image and push it with the specified tag (+latest).
