/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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
	"fmt"
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"

	"github.com/influxdb/influxdb/client/v2"
	str "github.com/intelsdi-x/snap-plugin-utilities/strings"
)

const (
	name                      = "influx"
	version                   = 12
	pluginType                = plugin.PublisherPluginType
	maxInt64                  = ^uint64(0) / 2
	defaultTimestampPrecision = "s"
)

// Meta returns a plugin meta data
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
}

//NewInfluxPublisher returns an instance of the InfuxDB publisher
func NewInfluxPublisher() *influxPublisher {
	return &influxPublisher{}
}

type influxPublisher struct {
}

func (f *influxPublisher) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()

	r1, err := cpolicy.NewStringRule("host", true)
	handleErr(err)
	r1.Description = "Influxdb host"
	config.Add(r1)

	r2, err := cpolicy.NewIntegerRule("port", true)
	handleErr(err)
	r2.Description = "Influxdb port"
	config.Add(r2)

	r3, err := cpolicy.NewStringRule("database", true)
	handleErr(err)
	r3.Description = "Influxdb db name"
	config.Add(r3)

	r4, err := cpolicy.NewStringRule("user", true)
	handleErr(err)
	r4.Description = "Influxdb user"
	config.Add(r4)

	r5, err := cpolicy.NewStringRule("password", true)
	handleErr(err)
	r5.Description = "Influxdb password"
	config.Add(r4)

	cp.Add([]string{""}, config)
	return cp, nil
}

// Publish publishes metric data to influxdb
// currently only 0.9 version of influxdb are supported
func (f *influxPublisher) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	logger := getLogger(config)
	var metrics []plugin.PluginMetricType

	switch contentType {
	case plugin.SnapGOBContentType:
		dec := gob.NewDecoder(bytes.NewBuffer(content))
		if err := dec.Decode(&metrics); err != nil {
			logger.WithFields(log.Fields{
				"err": err,
			}).Error("decoding error")
			return err
		}
	default:
		logger.Errorf("unknown content type '%v'", contentType)
		return fmt.Errorf("Unknown content type '%s'", contentType)
	}

	u, err := url.Parse(fmt.Sprintf("http://%s:%d", config["host"].(ctypes.ConfigValueStr).Value, config["port"].(ctypes.ConfigValueInt).Value))
	if err != nil {
		logger.Fatal(err)
		return err
	}

	con, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     u.String(),
		Username: config["user"].(ctypes.ConfigValueStr).Value,
		Password: config["password"].(ctypes.ConfigValueStr).Value,
	})

	if err != nil {
		logger.Fatal(err)
		return err
	}

	//Set up batch points
	bps, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:        config["database"].(ctypes.ConfigValueStr).Value,
		RetentionPolicy: "default",
		Precision:       defaultTimestampPrecision,
	})

	for _, m := range metrics {
		ns := m.Namespace()
		tags := map[string]string{"source": m.Source()}
		if m.Labels_ != nil {
			for _, label := range m.Labels_ {
				tags[label.Name] = m.Namespace()[label.Index]
				ns = str.Filter(
					ns,
					func(n string) bool {
						return n != label.Name
					},
				)
			}
		}
		for k, v := range m.Tags() {
			tags[k] = v
		}

		// NOTE: uint64 is specifically not supported by influxdb client due to potential overflow
		//without convertion of uint64 to int64, data with uint64 type will be saved as strings in influx database
		data := m.Data()
		v, ok := m.Data().(uint64)
		if ok {
			data = int64(v)
			if v > maxInt64 {
				log.Errorf("Overflow during conversion uint64 to int64, value after conversion to int64: %d, desired uint64 value: %d ", data, v)
			}
		}
		pt, err := client.NewPoint(strings.Join(ns, "/"), tags, map[string]interface{}{
			"value": data,
		}, m.Timestamp())
		if err != nil {
			logger.WithFields(log.Fields{
				"err":          err,
				"batch-points": bps.Points(),
				"point":        pt,
			}).Error("Publishing failed. Problem creating data point")
			return err
		}
		bps.AddPoint(pt)
	}

	err = con.Write(bps)
	if err != nil {
		logger.WithFields(log.Fields{
			"err":          err,
			"batch-points": bps,
		}).Error("publishing failed")
		return err
	}
	logger.WithFields(log.Fields{
		"batch-points": bps.Points(),
	}).Debug("publishing metrics")

	return nil
}

func handleErr(e error) {
	if e != nil {
		panic(e)
	}
}

func getLogger(config map[string]ctypes.ConfigValue) *log.Entry {
	logger := log.WithFields(log.Fields{
		"plugin-name":    name,
		"plugin-version": version,
		"plugin-type":    pluginType.String(),
	})

	// default
	log.SetLevel(log.WarnLevel)

	if debug, ok := config["debug"]; ok {
		switch v := debug.(type) {
		case ctypes.ConfigValueBool:
			if v.Value {
				log.SetLevel(log.DebugLevel)
				return logger
			}
		default:
			logger.WithFields(log.Fields{
				"field":         "debug",
				"type":          v,
				"expected type": "ctypes.ConfigValueBool",
			}).Error("invalid config type")
		}
	}

	if loglevel, ok := config["log-level"]; ok {
		switch v := loglevel.(type) {
		case ctypes.ConfigValueStr:
			switch strings.ToLower(v.Value) {
			case "warn":
				log.SetLevel(log.WarnLevel)
			case "error":
				log.SetLevel(log.ErrorLevel)
			case "debug":
				log.SetLevel(log.DebugLevel)
			case "info":
				log.SetLevel(log.InfoLevel)
			default:
				log.WithFields(log.Fields{
					"value":             strings.ToLower(v.Value),
					"acceptable values": "warn, error, debug, info",
				}).Warn("invalid config value")
			}
		default:
			logger.WithFields(log.Fields{
				"field":         "log-level",
				"type":          v,
				"expected type": "ctypes.ConfigValueStr",
			}).Error("invalid config type")
		}
	}

	return logger
}
