// +build medium

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package influxdb

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

func init() {
	//Do a ping to make sure the docker image actually came up. Otherwise this can fail Travis builds
	for i := 0; i < 3; i++ {
		resp, err := http.Get("http://" + os.Getenv("SNAP_INFLUXDB_HOST") + ":8086/ping")
		if err != nil || resp.StatusCode != 204 {
			//Try again after 3 seconds
			time.Sleep(3 * time.Second)
		} else {
			//Give the run.sh time to create the test database
			time.Sleep(5 * time.Second)
			return
		}
	}
	//If we got here, we failed to get to the server
	panic("Unable to connect to Influx host. Aborting test.")
}

func TestInfluxPublish(t *testing.T) {

	Convey("snap plugin InfluxDB integration testing with Influx", t, func() {
		var retention string

		if strings.HasPrefix(os.Getenv("INFLUXDB_VERSION"), "0.") {
			retention = "default"
		} else {
			retention = "autogen"
		}
		config := plugin.Config{
			"host":          os.Getenv("SNAP_INFLUXDB_HOST"),
			"skip-verify":   false,
			"user":          "root",
			"password":      "root",
			"database":      "test",
			"retention":     retention,
			"isMultiFields": false,
			"debug":         false,
			"log-level":     "debug",
			"precision":     "s",
		}

		config["scheme"] = HTTP
		config["port"] = int64(8086)
		tests(HTTP, config)

		config["scheme"] = UDP
		config["port"] = int64(4444)
		tests(UDP, config)
	})
}

func tests(scheme string, config plugin.Config) {
	ip := &InfluxPublisher{}
	tags := map[string]string{"zone": "red"}
	mcfg := map[string]interface{}{"field": "abc123"}

	Convey("Publish integer metric via "+scheme, func() {
		metrics := []plugin.Metric{
			{
				Namespace: plugin.NewNamespace("foo"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      99,
			},
		}
		err := ip.Publish(metrics, config)
		So(err, ShouldBeNil)
	})

	Convey("Publish float metric via "+scheme, func() {
		metrics := []plugin.Metric{
			{
				Namespace: plugin.NewNamespace("bar"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      3.141,
			},
		}
		err := ip.Publish(metrics, config)
		So(err, ShouldBeNil)
	})

	Convey("Publish uint64 metric via "+scheme, func() {
		metrics := []plugin.Metric{
			{
				Namespace: plugin.NewNamespace("uin"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      uint64(123),
			},
		}
		err := ip.Publish(metrics, config)
		So(err, ShouldBeNil)
	})

	Convey("Publish larger than uint64 metric via "+scheme, func() {
		metrics := []plugin.Metric{
			{
				Namespace: plugin.NewNamespace("lar"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      ^uint64(0)/2 + 1,
			},
		}
		err := ip.Publish(metrics, config)
		So(err, ShouldBeNil)
	})

	Convey("Publish string metric via "+scheme, func() {
		metrics := []plugin.Metric{
			{
				Namespace: plugin.NewNamespace("qux"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      "bar",
			},
		}
		err := ip.Publish(metrics, config)
		So(err, ShouldBeNil)
	})

	Convey("Publish boolean metric via "+scheme, func() {
		metrics := []plugin.Metric{
			{
				Namespace: plugin.NewNamespace("baz"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      true,
			},
		}
		err := ip.Publish(metrics, config)
		So(err, ShouldBeNil)
	})

	Convey("Publish multiple metrics via "+scheme, func() {
		metrics := []plugin.Metric{
			{
				Namespace: plugin.NewNamespace("foo"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      13,
			},
			{
				Namespace: plugin.NewNamespace("bar"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      2.718,
			},
		}
		err := ip.Publish(metrics, config)
		So(err, ShouldBeNil)
	})

	Convey("Publish dynamic metrics via "+scheme, func() {
		dynamicNS1 := plugin.NewNamespace("foo").
			AddDynamicElement("dynamic", "dynamic elem").
			AddStaticElement("bar")
		dynamicNS2 := plugin.NewNamespace("foo").
			AddDynamicElement("dynamic_one", "dynamic element one").
			AddDynamicElement("dynamic_two", "dynamic element two").
			AddStaticElement("baz")

		dynamicNS1[1].Value = "fooval"
		dynamicNS2[1].Value = "barval"
		dynamicNS2[2].Value = "bazval"

		metrics := []plugin.Metric{
			{
				Namespace: dynamicNS1,
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      17,
			},
			{
				Namespace: dynamicNS2,
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      23,
			},
		}
		err := ip.Publish(metrics, config)
		So(err, ShouldBeNil)
	})

	Convey("Publish nil value of metric via "+scheme, func() {
		metrics := []plugin.Metric{
			{
				Namespace: plugin.NewNamespace("baz"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "empty unit",
				Data:      nil,
			},
		}
		err := ip.Publish(metrics, config)
		So(err, ShouldBeNil)
	})

	Convey("Publish multiple fields to one metric via "+scheme, func() {
		config["isMultiFields"] = true
		metrics := []plugin.Metric{
			{
				Namespace: plugin.NewNamespace("a", "b", "x"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      123.456,
			},
			{
				Namespace: plugin.NewNamespace("a", "b", "y"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      987.654,
			},
			{
				Namespace: plugin.NewNamespace("a", "b", "z"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      18,
			},
			{
				Namespace: plugin.NewNamespace("a", "b", "z"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      512,
			},
		}
		err := ip.Publish(metrics, config)
		So(err, ShouldBeNil)
	})

	Convey("Publish multiple fields to two metrics via "+scheme, func() {
		config["isMultiFields"] = true
		ntags := map[string]string{"zone": "red", "light": "yellow"}
		metrics := []plugin.Metric{
			{
				Namespace: plugin.NewNamespace("influx", "x"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      333.6,
			},
			{
				Namespace: plugin.NewNamespace("influx", "y"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      666.3,
			},
			{
				Namespace: plugin.NewNamespace("influx", "z"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      tags,
				Unit:      "someunit",
				Data:      173,
			},
			{
				Namespace: plugin.NewNamespace("influx", "r"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      ntags,
				Unit:      "someunit",
				Data:      256,
			},
			{
				Namespace: plugin.NewNamespace("influx", "s"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      ntags,
				Unit:      "someunit",
				Data:      128,
			},
			{
				Namespace: plugin.NewNamespace("influx", "s"),
				Timestamp: time.Now(),
				Config:    mcfg,
				Tags:      ntags,
				Unit:      "someunit",
				Data:      64,
			},
		}
		err := ip.Publish(metrics, config)
		So(err, ShouldBeNil)
	})
}
