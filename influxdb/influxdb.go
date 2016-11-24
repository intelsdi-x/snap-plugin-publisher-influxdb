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

package influxdb

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	name                      = "influxdb"
	version                   = 15
	pluginType                = plugin.PublisherPluginType
	maxInt64                  = ^uint64(0) / 2
	defaultTimestampPrecision = "s"
)

var (
	// The maximum time a connection can sit around unused.
	maxConnectionIdle = time.Minute * 30
	// How frequently idle connections are checked
	watchConnctionWait = time.Minute * 15
	// Our connection pool
	connPool = make(map[string]*clientConnection)
	// Mutex for synchronizing connection pool changes
	m = &sync.Mutex{}
)

func init() {
	go watchConnections()
}

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
	config.Add(r5)

	r6, err := cpolicy.NewStringRule("retention", false, "autogen")
	handleErr(err)
	r6.Description = "Influxdb retention policy"
	config.Add(r6)

	r7, err := cpolicy.NewBoolRule("https", false, false)
	handleErr(err)
	r7.Description = "Influxdb HTTPS connection"
	config.Add(r7)

	r8, err := cpolicy.NewBoolRule("skip-verify", false, false)
	handleErr(err)
	r8.Description = "Influxdb HTTPS Skip certificate verification"
	config.Add(r8)

	cp.Add([]string{""}, config)
	return cp, nil
}

func watchConnections() {
	for {
		time.Sleep(watchConnctionWait)
		for k, c := range connPool {

			if time.Now().Sub(c.LastUsed) > maxConnectionIdle {
				m.Lock()
				// Close the connection
				c.close()
				// Remove from the pool
				delete(connPool, k)
				m.Unlock()
			}
		}
	}
}

// Publish publishes metric data to influxdb
// currently only 0.9 version of influxdb are supported
func (f *influxPublisher) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	logger := getLogger(config)
	var metrics []plugin.MetricType

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

	con, err := selectClientConnection(config)
	if err != nil {
		logger.Error(err)
		return err
	}

	//Set up batch points
	bps, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:        config["database"].(ctypes.ConfigValueStr).Value,
		RetentionPolicy: config["retention"].(ctypes.ConfigValueStr).Value,
		Precision:       defaultTimestampPrecision,
	})

	for _, m := range metrics {
		tags := map[string]string{}
		ns := m.Namespace().Strings()

		isDynamic, indexes := m.Namespace().IsDynamic()
		if isDynamic {
			for i, j := range indexes {
				// The second return value from IsDynamic(), in this case `indexes`, is the index of
				// the dynamic element in the unmodified namespace. However, here we're deleting
				// elements, which is problematic when the number of dynamic elements in a namespace is
				// greater than 1. Therefore, we subtract i (the loop iteration) from j
				// (the original index) to compensate.
				//
				// Remove "data" from the namespace and create a tag for it
				ns = append(ns[:j-i], ns[j-i+1:]...)
				tags[m.Namespace()[j].Name] = m.Namespace()[j].Value
			}
		}

		// Add "unit"" if we do not already have a "unit" tag
		if _, ok := m.Tags()["unit"]; !ok {
			tags["unit"] = m.Unit()
		}

		// Process the tags for this metric
		for k, v := range m.Tags() {
			// Convert the standard tag describing where the plugin is running to "source"
			if k == core.STD_TAG_PLUGIN_RUNNING_ON {
				// Unless the "source" tag is already being used
				if _, ok := m.Tags()["source"]; !ok {
					k = "source"
				}
			}
			tags[k] = v
		}

		data := m.Data()

		//publishing of nil value causes errors
		if data == nil {
			log.Errorf("Received nil value of metric, this metric is not published, namespace: %s, timestamp: %s", m.Namespace().String(), m.Timestamp().String())
			continue
		}

		// NOTE: uint64 is specifically not supported by influxdb client due to potential overflow
		//without convertion of uint64 to int64, data with uint64 type will be saved as strings in influx database
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

	err = con.write(bps)
	if err != nil {
		logger.WithFields(log.Fields{
			"err":          err,
			"batch-points": bps,
		}).Error("publishing failed")
		// Remove connction from pool since something is wrong
		m.Lock()
		con.close()
		delete(connPool, con.Key)
		m.Unlock()
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

type clientConnection struct {
	Key      string
	Conn     *client.Client
	LastUsed time.Time
}

// Map the batch points write into client.Client
func (c *clientConnection) write(bps client.BatchPoints) error {
	return (*c.Conn).Write(bps)
}

// Map the close function into client.Client
func (c *clientConnection) close() error {
	return (*c.Conn).Close()
}

func selectClientConnection(config map[string]ctypes.ConfigValue) (*clientConnection, error) {
	// This is not an ideal way to get the logger but deferring solving this for a later date
	logger := getLogger(config)

	var prefix = "http"
	if config["https"].(ctypes.ConfigValueBool).Value {
		prefix = "https"
	}

	u, err := url.Parse(fmt.Sprintf("%s://%s:%d", prefix, config["host"].(ctypes.ConfigValueStr).Value, config["port"].(ctypes.ConfigValueInt).Value))
	if err != nil {
		return nil, err
	}

	// Pool changes need to be safe (read & write) since the plugin can be called concurrently by snapteld.
	m.Lock()
	defer m.Unlock()

	user := config["user"].(ctypes.ConfigValueStr).Value
	pass := config["password"].(ctypes.ConfigValueStr).Value
	db := config["database"].(ctypes.ConfigValueStr).Value
	skipVerify := config["skip-verify"].(ctypes.ConfigValueBool).Value
	key := connectionKey(u, user, db)

	// Do we have a existing client?
	if connPool[key] == nil {
		// create one and add to the pool
		con, err := client.NewHTTPClient(client.HTTPConfig{
			Addr:               u.String(),
			Username:           user,
			Password:           pass,
			InsecureSkipVerify: skipVerify,
		})

		if err != nil {
			return nil, err
		}

		cCon := &clientConnection{
			Key:      key,
			Conn:     &con,
			LastUsed: time.Now(),
		}
		// Add to the pool
		connPool[key] = cCon

		logger.Debug("Opening new InfluxDB connection[", user, "@", db, " ", u.String(), "]")
		return connPool[key], nil
	}
	// Update when it was accessed
	connPool[key].LastUsed = time.Now()
	// Return it
	logger.Debug("Using open InfluxDB connection[", user, "@", db, " ", u.String(), "]")
	return connPool[key], nil
}

func connectionKey(u *url.URL, user, db string) string {
	return fmt.Sprintf("%s:%s:%s", u.String(), user, db)
}
