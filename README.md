# Contents

1. [Overview](#1-overview)<br/>
   1.1. [Purpose](#11-purpose)<br/>
   1.2. [Definitions](#12-definitions)<br/>
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
   &nbsp;&nbsp;&nbsp;5.2.2. [Resolution Algorithm](#522-resolution-algorithm)<br/>
   &nbsp;&nbsp;&nbsp;5.2.2 [Results Pagination](#523-results-pagination)<br/>
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

TODO

## 1.2. Definitions

TODO

# 2. Configuration

The service is configurable using the environment variables:

| Variable     | Example value                                          | Description                                                                    |
|--------------|--------------------------------------------------------|--------------------------------------------------------------------------------|
| API_PORT     | `8081`                                                 | gRPC API port                                                                  |
| DB_URI       | `mongodb+srv://localhost/?retryWrites=true&w=majority` | DB connection URI                                                              |
| DB_NAME      | `subscriptions`                                        | DB name to store the data                                                      |
| DB_TBL_NAME  | `subscriptions`                                        | DB table name to store the tree data                                           |
| PATTERNS_URI | `http://localhost:8080`                                | [Patterns](https://github.com/meandros-messaging/patterns) dependency service URI |

# 3. Deployment

## 3.1. Prerequisites

A general note is that there should be a MongoDB cluster deployed to be used for storing the pattern data.
It's possible to obtain a free cluster for testing purposes using [Atlas](https://www.mongodb.com/atlas/database).

## 3.2. Bare

Preconditions:
1. Build patterns executive using ```make build```
2. Run the [patterns](https://github.com/meandros-messaging/patterns) dependency service

Then run the command:
```shell
API_PORT=8081 \
DB_URI=mongodb+srv://localhost/?retryWrites=true&w=majority \
DB_NAME=subscriptions \
DB_TBL_NAME=subscriptions \
PATTERNS_URI=http://localhost:8081 \
./subscriptions
```

## 3.3. Docker

TODO: run the patterns and subscriptions in the same network

alternatively, it's possible to build and run the new docker image in place using the command:
(note that the command below requires all env vars to be set in the file `env.txt`)

```shell
make run
```

## 3.4. K8s

TODO

### 3.4.1. Helm

TODO

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
  version: 42
  description: Orders that are not in Helsinki
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
  version: 0
  description: Messages that have foo=bar and reply-to address of John Doe
  includes:
    all: true
    matchers:
    - key: reply-to
      pattern: 
        code: john.doe@email.com
        regex: john.doe@email.com
      partial: false
    - key: foo
      pattern: 
        code: bar
        regex: bar
      partial: false
  excludes: {}
```

#### 5.2.1.1. Subscription

| Attribute   | Type                                 | Description                                                 |
|-------------|--------------------------------------|-------------------------------------------------------------|
| name        | String                               | Unique subscription name                                    |
| version     | Unsigned Integer                     | Subscription entry version for the optimistic lock purpose  |
| description | String                               | Human readable subscription description                     |
| includes    | [Matcher Group](#5212-matcher-group) | Matchers to include the subscription to query results       |
| excludes    | [Matcher Group](#5212-matcher-group) | Matchers to exclude the subscription from the query results |

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

### 5.2.2. Resolution Algorithm

Pseudocode:
```text
- for each ($key, $pattern_code) pair resolved from the external service by the input sorted $metadata map
    - query all subscriptions where
        name is greater than $cursor
            and 
        that have any matcher in the includes group where
            key == $key 
                and 
            pattern_code == $pattern_code
    - for each $subscription from the step (2)
        - if includes matchers group has all attribute set to true
            - check that all matchers where 
                (key, pattern_code) != ($key, $pattern_code)
                are matching the corresponding values in the input $metadata map
                if any not matching, skip the $subscription
        - for each matcher in the excludes group
            - assume $all_matches = true
            - if matches anything in the input $metadata map
                - if the excludes group has all set to false
                    - skip the $subscription
            - else if the excludes group has all set to true
                - interrupt a further check
        - if not excluded in the checks above
            - include the subscription into the $results
            - $cursor = $subscription.name
```

### 5.2.3. Results Pagination

The limit and cursor search parameters are used to support the results' pagination.

TODO

## 5.3. Limitations

| #     | Summary                        | Description                                                                                                                                   |
|-------|--------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| LIM-1 | Excluding only is not allowed  | A subscription should have at least 1 key entry with non `exclude` Action. Otherwise the subscription never matches anything in practice.     |
| LIM-2 | Multiple required key matchers | A subscription should have at least 2 key entries with `require` Action. If single only, then `require` type should be replaced with `match`. |

# 6. Contributing

## 6.1. Versioning

The service uses the [semantic versioning](http://semver.org/).
The single source of the version info is constant `version` declared in the [cmd/main.go](cmd/main.go) file.
The script [scripts/version.sh](scripts/version.sh) is used to extract that version.

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
