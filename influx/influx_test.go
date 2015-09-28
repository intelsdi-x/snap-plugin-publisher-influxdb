/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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

// +build unit
package influx

import (
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInfluxDBPlugin(t *testing.T) {
	Convey("Meta should return metadata for the plugin", t, func() {
		meta := Meta()
		So(meta.Name, ShouldResemble, name)
		So(meta.Version, ShouldResemble, version)
		So(meta.Type, ShouldResemble, plugin.PublisherPluginType)
	})

	Convey("Create InfluxPublisher", t, func() {
		ip := NewInfluxPublisher()
		Convey("So ip should not be nil", func() {
			So(ip, ShouldNotBeNil)
		})
		Convey("So ip should be of influxPublisher type", func() {
			So(ip, ShouldHaveSameTypeAs, &influxPublisher{})
		})
		Convey("ip.GetConfigPolicy() should return a config policy", func() {
			configPolicy := ip.GetConfigPolicy()
			Convey("So config policy should not be nil", func() {
				So(configPolicy, ShouldNotBeNil)
			})
			Convey("So config policy should be a cpolicy.ConfigPolicy", func() {
				So(configPolicy, ShouldHaveSameTypeAs, cpolicy.ConfigPolicy{})
			})
			testConfig := make(map[string]ctypes.ConfigValue)
			testConfig["host"] = ctypes.ConfigValueStr{Value: "localhost"}
			testConfig["port"] = ctypes.ConfigValueInt{Value: 8086}
			testConfig["user"] = ctypes.ConfigValueStr{Value: "root"}
			testConfig["password"] = ctypes.ConfigValueStr{Value: "root"}
			testConfig["database"] = ctypes.ConfigValueStr{Value: "test"}
			cfg, errs := configPolicy.Get([]string{""}).Process(testConfig)
			Convey("So config policy should process testConfig and return a config", func() {
				So(cfg, ShouldNotBeNil)
			})
			Convey("So testConfig processing should return no errors", func() {
				So(errs.HasErrors(), ShouldBeFalse)
			})
			testConfig["port"] = ctypes.ConfigValueStr{Value: "8086"}
			cfg, errs = configPolicy.Get([]string{""}).Process(testConfig)
			Convey("So config policy should not return a config after processing invalid testConfig", func() {
				So(cfg, ShouldBeNil)
			})
			Convey("So testConfig processing should return errors", func() {
				So(errs.HasErrors(), ShouldBeTrue)
			})
		})
	})
}
