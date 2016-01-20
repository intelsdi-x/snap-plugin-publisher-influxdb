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
	"fmt"
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"

	"github.com/influxdb/influxdb/client"
	str "github.com/intelsdi-x/snap-plugin-utilities/strings"
)

const (
	name       = "influx"
	version    = 8
	pluginType = plugin.PublisherPluginType
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
		handleErr(err)
	}

	conf := client.Config{
		URL:       *u,
		Username:  config["user"].(ctypes.ConfigValueStr).Value,
		Password:  config["password"].(ctypes.ConfigValueStr).Value,
		UserAgent: "snap-publisher",
	}

	con, err := client.NewClient(conf)
	if err != nil {
		logger.Fatal(err)
	}

	_, ver, err := con.Ping()
	if err != nil {
		logger.WithFields(log.Fields{
			"metrics":      metrics,
			"config":       config,
			"influxdb-ver": ver,
		}).Error("influxdb connection failed")
		handleErr(err)
	}

	pts := make([]client.Point, len(metrics))
	for i, m := range metrics {
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
		pts[i] = client.Point{
			Measurement: strings.Join(ns, "/"),
			Time:        m.Timestamp(),
			Tags:        tags,
			Fields: map[string]interface{}{
				"value": m.Data(),
			},
			Precision: "s",
		}
	}

	bps := client.BatchPoints{
		Points:          pts,
		Database:        config["database"].(ctypes.ConfigValueStr).Value,
		RetentionPolicy: "default",
	}

	_, err = con.Write(bps)
	if err != nil {
		logger.WithFields(log.Fields{
			"err":          err,
			"batch-points": bps,
		}).Error("publishing failed")
	}
	logger.WithFields(log.Fields{
		"batch-points": bps,
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
