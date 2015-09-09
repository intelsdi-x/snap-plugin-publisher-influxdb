//
// +build unit

package influx

import (
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
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
		Convey("ip.GetConfigPolicyNode() should return a config policy node", func() {
			config := ip.GetConfigPolicyNode()
			Convey("So config should not be nil", func() {
				So(config, ShouldNotBeNil)
			})
			Convey("So config should be a cpolicy.ConfigPolicyNode", func() {
				So(config, ShouldHaveSameTypeAs, cpolicy.ConfigPolicyNode{})
			})
		})
	})
}
