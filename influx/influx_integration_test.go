// +build integration

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
package influx

import (
	"bytes"
	"encoding/gob"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInfluxPublish(t *testing.T) {
	config := make(map[string]ctypes.ConfigValue)

	Convey("Pulse Plugin InfluxDB integration testing with Influx", t, func() {
		var buf bytes.Buffer

		config["host"] = ctypes.ConfigValueStr{Value: os.Getenv("PULSE_INFLUXDB_HOST")}
		config["port"] = ctypes.ConfigValueInt{Value: 8086}
		config["user"] = ctypes.ConfigValueStr{Value: "root"}
		config["password"] = ctypes.ConfigValueStr{Value: "root"}
		config["database"] = ctypes.ConfigValueStr{Value: "test"}
		config["debug"] = ctypes.ConfigValueBool{Value: false}
		config["log-level"] = ctypes.ConfigValueStr{Value: "debug"}

		ip := NewInfluxPublisher()
		cp, _ := ip.GetConfigPolicy()
		cfg, _ := cp.Get([]string{""}).Process(config)
		tags := map[string]string{"zone": "red"}
		labels := []core.Label{}

		Convey("Publish integer metric", func() {
			metrics := []plugin.PluginMetricType{
				*plugin.NewPluginMetricType([]string{"foo"}, time.Now(), "127.0.0.1", tags, labels, 99),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.PulseGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish float metric", func() {
			metrics := []plugin.PluginMetricType{
				*plugin.NewPluginMetricType([]string{"bar"}, time.Now(), "127.0.0.1", tags, labels, 3.141),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.PulseGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish string metric", func() {
			metrics := []plugin.PluginMetricType{
				*plugin.NewPluginMetricType([]string{"qux"}, time.Now(), "127.0.0.1", tags, labels, "bar"),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.PulseGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish boolean metric", func() {
			metrics := []plugin.PluginMetricType{
				*plugin.NewPluginMetricType([]string{"baz"}, time.Now(), "127.0.0.1", tags, labels, true),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.PulseGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish multiple metrics", func() {
			metrics := []plugin.PluginMetricType{
				*plugin.NewPluginMetricType([]string{"foo"}, time.Now(), "127.0.0.1", tags, labels, 101),
				*plugin.NewPluginMetricType([]string{"bar"}, time.Now(), "127.0.0.1", tags, labels, 5.789),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.PulseGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

	})
}
