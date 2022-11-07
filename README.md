DISCONTINUATION OF PROJECT. 

This project will no longer be maintained by Intel.

This project has been identified as having known security escapes.

Intel has ceased development and contributions including, but not limited to, maintenance, bug fixes, new releases, or updates, to this project.  

Intel no longer accepts patches to this project.

# DISCONTINUATION OF PROJECT 

**This project will no longer be maintained by Intel.  Intel will not provide or guarantee development of or support for this project, including but not limited to, maintenance, bug fixes, new releases or updates.  Patches to this project are no longer accepted by Intel. If you have an ongoing need to use this project, are interested in independently developing it, or would like to maintain patches for the community, please create your own fork of the project.**


[![Build Status](https://travis-ci.org/intelsdi-x/snap-plugin-publisher-influxdb.svg?branch=master)](https://travis-ci.org/intelsdi-x/snap-plugin-publisher-influxdb)

# Snap publisher plugin - InfluxDB

This plugin supports pushing metrics into an InfluxDB instance.

It's used in the [Snap framework](http://github.com/intelsdi-x/snap).

1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
  * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
  * [Collected Metrics](#collected-metrics)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license)
6. [Acknowledgements](#acknowledgements)

## Getting Started

### System Requirements

* [golang 1.6+](https://golang.org/dl/) (needed only for building)

Support Matrix

Influxdb | Influxdb Publisher | Snap
-----|-----|-----
1.0 | 16 | 1.0.0
1.1 | 16 | 1.0.0
1.1.1 | 16 | 1.0.0

### Known Limitation

* InfluxDB (tested with InfluxDB 1.0) does not support uint64 as type of data. Metrics with uint64 type are converted to int64 by Snap publisher plugin. uint64 values higher than maximum int64 value are converted to negative value and saved in InfluxDB. Overflow cases are logged.

### Installation

#### Download InfluxDB plugin binary:
You can get the pre-built binaries for your OS and architecture at plugin's [GitHub Releases](https://github.com/intelsdi-x/snap-plugin-publisher-influxdb/releases) page.

#### To build the plugin binary:
Fork https://github.com/intelsdi-x/snap-plugin-publisher-influxdb

Clone repo into `$GOPATH/src/github.com/intelsdi-x/`:

```
$ git clone https://github.com/<yourGithubID>/snap-plugin-publisher-influxdb.git
```

Build the plugin by running make within the cloned repo:
```
$ make
```
This builds the plugin in `./build`

### Configuration and Usage
* Set up the [Snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started)

## Documentation

The plugin expects you to provide the following parameters:
 - `host`
 - `database`
 - `user`
 - `password`

You can also set the following options if needed:
 - `skip-verify` defaults to `false` (boolean). Set to true to complain if the certificate used is not issued by a trusted CA.
 - `precision` defaults to `s` (string). The value can be changed to any of the following: n,u,ms,s,m,h. This will determine the precision of timestamps.
 - `isMultiFields` defaults to `false` (boolean). When it's true, plugin groups common namespaces, those that differ at the leaf and have same tags including values, into one data point with multiple influx fields.  
 - `port` defaults to `8086` which works with `http` and `https`. The port is `4444` for udp in the example.
 - `scheme` defaults to `http`.
   - `http`
   - `https`
   - `udp`
 - `retention` defaults to `autogen`, it indicates [retention policy](https://docs.influxdata.com/influxdb/v1.0/concepts/key_concepts/#retention-policy)
  for database with specified duration which determines how long InfluxDB keeps the data, for more information read
   [Retention Policy Management](https://docs.influxdata.com/influxdb/v1.0/query_language/database_management/#retention-policy-management).

### Examples

See [examples/tasks](https://github.com/intelsdi-x/snap-plugin-publisher-influxdb/tree/master/examples/tasks) folder for examples.  

Here are samples to illustrate the differences for `isMultiFields` flag. When *isMultiFields* is `false` which is the default setting, 
you have to query each measurement. While *isMultiFields* is `true`, plugin groups the common namespaces, those that differ at the leaf and have same tags including values, into one data point with multiple influx fields; you query the common namespace.

**Sample** *`isMultiField=false`*
```
select * from "/intel/psutil/load/load1"
```

| time | source | unit | value |
|---------------------|---------------|---------|-------|
| 1483997727411599704 | egu-mac01.lan | Load/1M | 1.76 |
| 1483997728412178616 | egu-mac01.lan | Load/1M | 1.76 |


**Sample** *`isMultiField=true`*
```
select * from "/intel/psutil/load"
```

| time | load1 | load15 | load5 | source | unit |
|---------------------|-------|--------|-------|---------------|---------|
| 1483996289995839909 | 2.05 |  |  | egu-mac01.lan | Load/1M |
| 1483996289995839909 |  | 6.21 |  | egu-mac01.lan | Load/1M |
| 1483996289995839909 |  |  | 5.26 | egu-mac01.lan | Load/1M |

### Roadmap

There isn't a current roadmap for this plugin, but it is in active development. As we launch this plugin, we do not have any outstanding requirements for the next release.

If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-publisher-influxdb/issues/new) and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-publisher-influxdb/pulls).

## Community Support
This repository is one of **many** plugins in **Snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support)

## Contributing
We love contributions! 

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License
[Snap](http://github.com/intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements
* Author: [Justin Guidroz](https://github.com/geauxvirtual)
* Author: [Joel Cooklin](https://github.com/jcooklin)

And **thank you!** Your contribution, through code and participation, is incredibly important to us.
