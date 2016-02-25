[![Build Status](https://travis-ci.org/intelsdi-x/snap-plugin-publisher-influxdb.svg?branch=master)](https://travis-ci.org/intelsdi-x/snap-plugin-publisher-influxdb)

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
- InfluxDB Plugin: v12 -> InfluxDB Version >= 0.9.4 -> snap version 0.13.0-beta and greater

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

###Configuration fields in task manifest for InfluxDB plugin
* **host** - InfluxDB host
* **port** - InfluxDB port
* **user** - InfluxDB user
* **password** - InfluxDB password
* **publish_timestamp** - optional parameter to define if plugin shall publish metric collection timestamp (default - parameter set to true) or not (parameter set to false). In the latter case timestamp will be set by InfluxDB. 

### Examples
Example of use snap-collector-mock1 collector plugin and InfluxDB publisher plugin to save collecting data in InfluxDB.

This is done from the snap directory.

In one terminal window, open the snap daemon (in this case with logging set to 1 and trust disabled):
```
$ $SNAP_PATH/bin/snapd -l 1 -t 0
```

In another terminal window:
Load snap-collector-mock1 collector plugin
```
$ $SNAP_PATH/bin/snapctl plugin load snap-collector-mock1
```

Load snap-plugin-publisher-influxdb publisher plugin
```
$ $SNAP_PATH/bin/snapctl plugin load snap-plugin-publisher-influxdb
```

See available metrics for your system
```
$ $SNAP_PATH/bin/snapctl metric list
```

Create a task manifest file (e.g. `task.json`):
```json
{
    "version": 1,
    "schedule": {
        "type": "simple",
        "interval": "10s"
    },
    "workflow": {
        "collect": {
            "metrics": {
                "/intel/mock/foo": {},
                "/intel/mock/bar": {},
                "/intel/mock/*/baz": {}
            },
             "config": {
                "/intel/mock": {
                    "user": "root",
                    "password": "secret"
                }
            },
            "process": null,
            "publish": [
                {
                    "plugin_name": "influx",
                    "config": {
                        "host": "host",
                        "port": 8086,
                        "database": "database",
                        "user": "user",
                        "password": "password",
                        "publish_timestamp": true
                    }
                }
            ]
        }
    }
}
```
Create task:
```
$ $SNAP_PATH/bin/snapctl task create -t task.json
Using task manifest to create task
Task created
ID: ae65bdc2-550a-4d0f-80a5-0b4c5aa98143
Name: Task-ae65bdc2-550a-4d0f-80a5-0b4c5aa98143
State: Running
```

Stop task:
```
$ $SNAP_PATH/bin/snapctl task stop ae65bdc2-550a-4d0f-80a5-0b4c5aa98143
Task stopped:
ID: ae65bdc2-550a-4d0f-80a5-0b4c5aa98143
```

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
