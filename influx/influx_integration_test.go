//
// +build integration

package influx

import (
	"bytes"
	"encoding/gob"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
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

		ip := NewInfluxPublisher()
		policy := ip.GetConfigPolicyNode()
		cfg, _ := policy.Process(config)

		Convey("Publish integer metric", func() {
			metrics := []plugin.PluginMetricType{
				*plugin.NewPluginMetricType([]string{"foo"}, time.Now(), "127.0.0.1", 99),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.PulseGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish float metric", func() {
			metrics := []plugin.PluginMetricType{
				*plugin.NewPluginMetricType([]string{"bar"}, time.Now(), "127.0.0.1", 3.141),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.PulseGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish string metric", func() {
			metrics := []plugin.PluginMetricType{
				*plugin.NewPluginMetricType([]string{"qux"}, time.Now(), "127.0.0.1", "bar"),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.PulseGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish boolean metric", func() {
			metrics := []plugin.PluginMetricType{
				*plugin.NewPluginMetricType([]string{"baz"}, time.Now(), "127.0.0.1", true),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.PulseGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

		Convey("Publish multiple metrics", func() {
			metrics := []plugin.PluginMetricType{
				*plugin.NewPluginMetricType([]string{"foo"}, time.Now(), "127.0.0.1", 101),
				*plugin.NewPluginMetricType([]string{"bar"}, time.Now(), "127.0.0.1", 5.789),
			}
			buf.Reset()
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := ip.Publish(plugin.PulseGOBContentType, buf.Bytes(), *cfg)
			So(err, ShouldBeNil)
		})

	})
}
