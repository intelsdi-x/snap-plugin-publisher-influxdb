// +build large

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
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
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
	config := make(map[string]ctypes.ConfigValue)

	Convey("snap plugin InfluxDB integration testing with Influx", t, func() {
		var buf bytes.Buffer

		config["host"] = ctypes.ConfigValueStr{Value: os.Getenv("SNAP_INFLUXDB_HOST")}
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

		Convey("Publish integer metric", func() {
			metrics := []plugin.MetricType{
				*plugin.NewMetricType(core.NewNamespace("foo"), time.Now(), tags, "some unit", 99),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish float metric", func() {
			metrics := []plugin.MetricType{
				*plugin.NewMetricType(core.NewNamespace("bar"), time.Now(), tags, "some unit", 3.141),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish string metric", func() {
			metrics := []plugin.MetricType{
				*plugin.NewMetricType(core.NewNamespace("qux"), time.Now(), tags, "some unit", "bar"),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish boolean metric", func() {
			metrics := []plugin.MetricType{
				*plugin.NewMetricType(core.NewNamespace("baz"), time.Now(), tags, "some unit", true),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish multiple metrics", func() {
			metrics := []plugin.MetricType{
				*plugin.NewMetricType(core.NewNamespace("foo"), time.Now(), tags, "some unit", 101),
				*plugin.NewMetricType(core.NewNamespace("bar"), time.Now(), tags, "some unit", 5.789),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.SnapGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

	})
}
