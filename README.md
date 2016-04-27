[![Build Status](https://travis-ci.com/intelsdi-x/snap-plugin-publisher-influxdb.svg?token=FkGfhS15Ai2yp19KAw41&branch=master)](https://travis-ci.com/intelsdi-x/snap-plugin-publisher-influxdb)

# snap publisher plugin - InfluxDB 

This plugin supports pushing metrics into an InfluxDB instance.

It's used in the [snap framework](http://github.com/intelsdi-x/snap).

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

* [golang 1.5+](https://golang.org/dl/) (needed only for building)

Support Matrix

- InfluxDB Plugin: v2 -> InfluxDB Version 0.9.1 -> snap version 0.2.0
- InfluxDB Plugin: v3 -> InfluxDB Version >= 0.9.1 -> snap version 0.2.0
- InfluxDB Plugin: v4 -> InfluxDB Version >= 0.9.1 -> snap version 0.3.x
- InfluxDB Plugin: v6 -> InfluxDB Version >= 0.9.1 -> snap version 0.8.0-beta
- InfluxDB Plugin: v7 -> InfluxDB Version >= 0.9.1 -> snap version 0.8.0-beta-114 and greater
- InfluxDB Plugin: v12 -> InfluxDB Version >= 0.9.3 -> snap version 0.8.0-beta-114 and greater

### Known Limitation

* InfluxDB (tested with InfluxDB 0.10.0) does not support uint64 as type of data. Metrics with uint64 type are converted to int64 by snap publisher plugin. uint64 values higher than maximum int64 value are converted to negative value and saved in InfluxDB. Overflow cases are logged.

### Installation

#### Download InfluxDB plugin binary:
You can get the pre-built binaries for your OS and architecture at snap's [GitHub Releases](https://github.com/intelsdi-x/snap/releases) page.

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
This builds the plugin in `/build/rootfs/`

### Configuration and Usage
* Set up the [snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started)
* Ensure `$SNAP_PATH` is exported  
`export SNAP_PATH=$GOPATH/src/github.com/intelsdi-x/snap/build`

## Documentation
<< @TODO

### Examples
<< @TODO

### Roadmap

There isn't a current roadmap for this plugin, but it is in active development. As we launch this plugin, we do not have any outstanding requirements for the next release.

If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-publisher-influxdb/issues/new) and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-publisher-influxdb/pulls).

## Community Support
This repository is one of **many** plugins in **snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support)

## Contributing
We love contributions! 

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License
[snap](http://github.com/intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements
* Author: [Justin Guidroz](https://github.com/geauxvirtual)
* Author: [Joel Cooklin](https://github.com/jcooklin)

And **thank you!** Your contribution, through code and participation, is incredibly important to us.
