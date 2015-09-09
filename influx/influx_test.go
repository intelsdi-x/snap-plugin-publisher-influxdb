//
// +build unit

package influx

import (
	"testing"

	"github.com/intelsdi-x/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInfluxDBPlugin(t *testing.T) {
	Convey("Meta should return metadata for the plugin", t, func() {
		meta := Meta()
		So(meta.Name, ShouldResemble, name)
		So(meta.Version, ShouldResemble, version)
		So(meta.Type, ShouldResemble, plugin.PublisherPluginType)
	})
}
