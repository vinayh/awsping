# awsping
Console tool to check the latency to each AWS region

[![Go Report Card](https://goreportcard.com/badge/github.com/ekalinin/awsping)](https://goreportcard.com/report/github.com/ekalinin/awsping)
[![codecov](https://codecov.io/gh/ekalinin/awsping/branch/master/graph/badge.svg)](https://codecov.io/gh/ekalinin/awsping)
[![Go Reference](https://pkg.go.dev/badge/github.com/ekalinin/awsping.svg)](https://pkg.go.dev/github.com/ekalinin/awsping)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/ekalinin/awsping)

# ToC

* [Usage](#usage)
  * [Test via TCP](#test-via-tcp)
  * [Test via HTTP](#test-via-http)
  * [Test via HTTPS](#test-via-https)
  * [Test several times](#test-several-times)
  * [Verbose mode](#verbose-mode)
  * [Get Help](#get-help)
* [Get binary file](#get-binary-file)
* [Build from sources](#build-from-sources)
* [Use with Docker](#use-with-docker)
  * [Build a Docker image](#build-a-docker-image)
  * [Run the Docker image](#run-the-docker-image)

# Usage

## Test via TCP

Output shows each region as `<region-code> (<location>)` followed by the
average latency. Regions whose endpoint can't be reached are reported as
`timeout` (or a short error) and sort to the bottom.

```bash
➥ ./awsping
eu-west-2 (London)                          1.71 ms
eu-west-3 (Paris)                           7.33 ms
eu-west-1 (Ireland)                        10.85 ms
eu-central-1 (Frankfurt)                   13.36 ms
us-east-1 (N. Virginia)                    78.44 ms
ca-central-1 (Central)                     85.02 ms
ap-south-1 (Mumbai)                       128.78 ms
sa-east-1 (São Paulo)                     185.57 ms
ap-northeast-1 (Tokyo)                    231.80 ms
ap-southeast-2 (Sydney)                   278.39 ms
ap-southeast-6 (New Zealand)              296.84 ms
me-south-1 (Bahrain)                        timeout
```

## Test via HTTP

```bash
➥ ./awsping -http
eu-west-2 (London)                         27.06 ms
eu-west-3 (Paris)                          31.33 ms
eu-west-1 (Ireland)                        42.99 ms
eu-central-1 (Frankfurt)                   44.62 ms
us-east-1 (N. Virginia)                   167.56 ms
ca-central-1 (Central)                    176.35 ms
ap-south-1 (Mumbai)                       383.13 ms
sa-east-1 (São Paulo)                     382.40 ms
ap-northeast-1 (Tokyo)                    497.25 ms
ap-southeast-2 (Sydney)                   572.68 ms
ap-southeast-6 (New Zealand)              621.30 ms
me-south-1 (Bahrain)                        timeout
```

## Test via HTTPS

```bash
➥ ./awsping -https
eu-west-2 (London)                        189.68 ms
eu-west-3 (Paris)                         196.65 ms
eu-west-1 (Ireland)                       197.75 ms
eu-central-1 (Frankfurt)                  212.76 ms
us-east-1 (N. Virginia)                   273.25 ms
ca-central-1 (Central)                    273.32 ms
ap-south-1 (Mumbai)                       580.50 ms
sa-east-1 (São Paulo)                     605.01 ms
ap-northeast-1 (Tokyo)                    765.88 ms
ap-southeast-2 (Sydney)                   860.53 ms
ap-southeast-6 (New Zealand)              964.32 ms
me-south-1 (Bahrain)                        timeout
```

## Test several times

```bash
➥ ./awsping -repeats 3
eu-west-2 (London)                          2.56 ms
eu-west-3 (Paris)                           8.60 ms
eu-west-1 (Ireland)                        10.19 ms
eu-central-1 (Frankfurt)                   13.53 ms
us-east-1 (N. Virginia)                    78.95 ms
ca-central-1 (Central)                     80.72 ms
ap-south-1 (Mumbai)                       184.53 ms
sa-east-1 (São Paulo)                     187.18 ms
ap-northeast-1 (Tokyo)                    234.30 ms
ap-southeast-2 (Sydney)                   278.47 ms
ap-southeast-6 (New Zealand)              302.51 ms
me-south-1 (Bahrain)                        timeout
```

## Verbose mode

Verbose level 1 adds an index, the region code column, and the full
region name (`Region Group (Location)`):

```bash
➥ ./awsping -repeats 3 -verbose 1
      Code            Region                                      Latency
    0 eu-west-2       Europe (London)                             2.04 ms
    1 eu-west-3       Europe (Paris)                              7.85 ms
    2 eu-west-1       Europe (Ireland)                           10.72 ms
    3 eu-central-1    Europe (Frankfurt)                         13.06 ms
    4 eu-central-2    Europe (Zurich)                            17.96 ms
    5 eu-south-2      Europe (Spain)                             20.60 ms
    6 eu-south-1      Europe (Milan)                             23.50 ms
    7 eu-north-1      Europe (Stockholm)                         27.25 ms
    8 il-central-1    Israel (Tel Aviv)                          60.08 ms
    9 us-east-1       US East (N. Virginia)                      78.69 ms
   10 ca-central-1    Canada (Central)                           84.34 ms
   11 us-east-2       US East (Ohio)                             89.57 ms
```

Verbose level 2 also breaks out per-try latencies. If a try errors,
that cell is shown as `-`; if every try errors, the average column
shows the error (e.g. `timeout`):

```bash
➥ ./awsping -repeats 3 -verbose 2
      Code            Region                                  Try #1          Try #2          Try #3     Avg Latency
    0 eu-west-2       Europe (London)                        1.95 ms         2.11 ms         2.02 ms         2.03 ms
    1 eu-west-3       Europe (Paris)                        10.84 ms         7.36 ms         9.65 ms         9.28 ms
    2 eu-west-1       Europe (Ireland)                      12.04 ms        11.12 ms        11.93 ms        11.70 ms
    3 eu-central-1    Europe (Frankfurt)                    13.59 ms        13.14 ms        13.27 ms        13.34 ms
    4 eu-central-2    Europe (Zurich)                       17.67 ms        17.56 ms        17.98 ms        17.74 ms
    5 eu-south-2      Europe (Spain)                        20.05 ms        19.95 ms        20.86 ms        20.29 ms
    6 eu-south-1      Europe (Milan)                        22.27 ms        22.39 ms        22.63 ms        22.43 ms
    7 eu-north-1      Europe (Stockholm)                    27.31 ms        27.19 ms        27.14 ms        27.22 ms
    8 il-central-1    Israel (Tel Aviv)                     62.50 ms        62.99 ms        60.41 ms        61.97 ms
    9 us-east-1       US East (N. Virginia)                 78.73 ms        80.20 ms        79.13 ms        79.35 ms
   10 ca-central-1    Canada (Central)                      85.30 ms        85.26 ms        84.41 ms        84.99 ms
   11 us-east-2       US East (Ohio)                        90.43 ms        88.41 ms        89.97 ms        89.60 ms
```

## Get Help

```bash
➜ ./awsping -h
Usage of ./awsping:
  -http
    	Use http transport (default is tcp)
  -https
    	Use https transport (default is tcp)
  -list-regions
    	Show list of regions
  -repeats int
    	Number of repeats (default 1)
  -service string
    	AWS Service: ec2, sdb, sns, sqs, ... (default "dynamodb")
  -v	Show version
  -verbose int
    	Verbosity level
```

# Get binary file

```bash
$ wget https://github.com/ekalinin/awsping/releases/latest/download/awsping_linux_amd64.tar.gz
$ tar xzvf awsping_linux_amd64.tar.gz
$ chmod +x awsping
$ ./awsping -v
```

# Build from sources

```bash
➥ make build
```

# Use with Docker
## Build a Docker image

```
$ docker build -t awsping .
```

## Run the Docker image
```
$ docker run --rm awsping
```

Arguments can be used as mentioned in the _Usage_ section.

i.e.:
```
$ docker run --rm awsping -repeats 3 -verbose 2
```
