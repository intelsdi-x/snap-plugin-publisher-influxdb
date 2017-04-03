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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	Name       = "influxdb"
	Version    = 22
	PluginType = "publisher"
	maxInt64   = ^uint64(0) / 2
	separator  = "\U0001f422"

	// HTTP represents its string constant
	HTTP = "http"
	// UDP represents its string constant
	UDP = "udp"
)

var (
	// The maximum time a connection can sit around unused.
	maxConnectionIdle = time.Minute * 30
	// How frequently idle connections are checked
	watchConnectionWait = time.Minute * 15
	// Our connection pool
	connPool = make(map[string]*clientConnection)
	// Mutex for synchronizing connection pool changes
	m           = &sync.Mutex{}
	initialized = false
)

func init() {
	go watchConnections()
}

//NewInfluxPublisher returns an instance of the InfluxDB publisher
func NewInfluxPublisher() *InfluxPublisher {
	return &InfluxPublisher{}
}

type InfluxPublisher struct {
}

type point struct {
	ns     []string
	tags   map[string]string
	ts     time.Time
	fields map[string]interface{}
}

type configuration struct {
	host, database, user, password, retention, precision, scheme, logLevel string
	port                                                                   int64
	skipVerify, isMultiFields                                              bool
}

func getConfig(config plugin.Config) (configuration, error) {
	cfg := configuration{}
	var err error

	cfg.host, err = config.GetString("host")
	if err != nil {
		return cfg, fmt.Errorf("%s: %s", err, "host")
	}

	cfg.database, err = config.GetString("database")
	if err != nil {
		return cfg, fmt.Errorf("%s: %s", err, "database")
	}

	cfg.user, err = config.GetString("user")
	if err != nil {
		return cfg, fmt.Errorf("%s: %s", err, "user")
	}

	cfg.password, err = config.GetString("password")
	if err != nil {
		return cfg, fmt.Errorf("%s: %s", err, "password")
	}

	cfg.retention, err = config.GetString("retention")
	if err != nil {
		return cfg, fmt.Errorf("%s: %s", err, "retention")
	}

	cfg.scheme, err = config.GetString("scheme")
	if err != nil {
		return cfg, fmt.Errorf("%s: %s", err, "scheme")
	}

	cfg.logLevel, err = config.GetString("log-level")
	if err != nil {
		cfg.logLevel = "undefined"
	}

	cfg.port, err = config.GetInt("port")
	if err != nil {
		return cfg, fmt.Errorf("%s: %s", err, "port")
	}

	cfg.skipVerify, err = config.GetBool("skip-verify")
	if err != nil {
		return cfg, fmt.Errorf("%s: %s", err, "skip-verify")
	}

	cfg.isMultiFields, err = config.GetBool("isMultiFields")
	if err != nil {
		return cfg, fmt.Errorf("%s: %s", err, "isMultiFields")
	}

	return cfg, nil
}

func (ip *InfluxPublisher) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()

	policy.AddNewStringRule([]string{""}, "host", true)
	policy.AddNewIntRule([]string{""}, "port", false, plugin.SetDefaultInt(8086))
	policy.AddNewStringRule([]string{""}, "database", true)
	policy.AddNewStringRule([]string{""}, "user", true)
	policy.AddNewStringRule([]string{""}, "password", true)
	policy.AddNewStringRule([]string{""}, "retention", false, plugin.SetDefaultString("autogen"))
	policy.AddNewBoolRule([]string{""}, "skip-verify", false, plugin.SetDefaultBool(false))
	policy.AddNewStringRule([]string{""}, "precision", false, plugin.SetDefaultString("ns"))
	policy.AddNewBoolRule([]string{""}, "isMultiFields", false, plugin.SetDefaultBool(false))
	policy.AddNewStringRule([]string{""}, "scheme", false, plugin.SetDefaultString(HTTP))

	return *policy, nil
}

func watchConnections() {
	for {
		time.Sleep(watchConnectionWait)
		for k, c := range connPool {
			if time.Now().Sub(c.LastUsed) > maxConnectionIdle {
				m.Lock()
				// Close the connection
				c.closeClientConnection()
				// Remove from the pool
				delete(connPool, k)
				m.Unlock()
			}
		}
	}
}

// Publish publishes metric data to influxdb
// currently only 0.9 version of influxdb are supported
func (ip *InfluxPublisher) Publish(metrics []plugin.Metric, pluginConfig plugin.Config) error {
	config, err := getConfig(pluginConfig)
	if err != nil {
		return err
	}

	logger := getLogger(config)

	con, err := selectClientConnection(config)
	if err != nil {
		logger.Error(err)
		return err
	}

	//Set up batch points
	bps, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:        config.database,
		RetentionPolicy: config.retention,
		Precision:       config.precision,
	})

	isMultiFields := config.isMultiFields
	mpoints := map[string]point{}
	for _, m := range metrics {
		ns, tags := replaceDynamicElement(m)

		// Add "unit"" if we do not already have a "unit" tag
		if _, ok := m.Tags["unit"]; !ok {
			tags["unit"] = m.Unit
		}

		// Process the tags for this metric
		for k, v := range m.Tags {
			// Convert the standard tag describing where the plugin is running to "source"
			if k == "plugin_running_on" {
				// Unless the "source" tag is already being used
				if _, ok := m.Tags["source"]; !ok {
					k = "source"
				}
			}
			tags[k] = v
		}

		data := m.Data

		//publishing of nil value causes errors
		if data == nil {
			log.Errorf("Received nil value of metric, this metric will not be published, namespace: %s, timestamp: %s", strings.Join(m.Namespace.Strings(), "/"), m.Timestamp.String())
			continue
		}

		// NOTE: uint64 is specifically not supported by influxdb client due to potential overflow
		//without convertion of uint64 to int64, data with uint64 type will be saved as strings in influx database
		v, ok := m.Data.(uint64)
		if ok {
			data = int64(v)
			if v > maxInt64 {
				log.Errorf("Overflow during conversion uint64 to int64, value after conversion to int64: %d, desired uint64 value: %d ", data, v)
			}
		}

		if !isMultiFields {
			pt, err := client.NewPoint(strings.Join(ns, "/"), tags, map[string]interface{}{
				"value": data,
			}, m.Timestamp)
			if err != nil {
				logger.WithFields(log.Fields{
					"err":          err,
					"batch-points": bps.Points(),
					"point":        pt,
				}).Error("Publishing failed. Problem creating data point")
				return err
			}
			bps.AddPoint(pt)
		} else {
			groupCommonNamespaces(m, tags, mpoints)
		}
	}

	if isMultiFields {
		for _, p := range mpoints {
			pt, err := client.NewPoint(strings.Join(p.ns, "/"), p.tags, p.fields, p.ts)
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
	}

	err = con.write(bps)
	if err != nil {
		logger.WithFields(log.Fields{
			"err":          err,
			"batch-points": bps,
		}).Error("publishing failed")
		// Remove connction from pool since something is wrong
		m.Lock()
		con.closeClientConnection()
		delete(connPool, con.Key)
		m.Unlock()
		return err
	}
	logger.WithFields(log.Fields{
		"batch-points": bps.Points(),
	}).Debug("publishing metrics")

	return nil
}

func getLogger(config configuration) *log.Entry {
	logger := log.WithFields(log.Fields{
		"plugin-name":    Name,
		"plugin-version": Version,
		"plugin-type":    PluginType,
	})

	// default
	log.SetLevel(log.WarnLevel)

	levelValue := config.logLevel
	if levelValue != "undefined" {
		if level, err := log.ParseLevel(strings.ToLower(levelValue)); err == nil {
			log.SetLevel(level)
		} else {
			log.WithFields(log.Fields{
				"value":             strings.ToLower(levelValue),
				"acceptable values": "warn, error, debug, info",
			}).Warn("Invalid log-level config value")
		}
	}
	return logger
}

type clientConnection struct {
	Key      string
	Conn     *client.Client
	LastUsed time.Time
}

// Create database if it doesn't exist
// workaround: use http instead of client library because of the issue
// ref: https://github.com/influxdata/influxdb/issues/8108
func (c *clientConnection) initDB(u *url.URL, user, pass, db string) error {
	urlStr := fmt.Sprintf("%s/query", u.String())

	req, err := http.NewRequest("POST", urlStr, nil)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("CREATE DATABASE %s", db)
	params := req.URL.Query()
	params.Set("q", query)
	req.URL.RawQuery = params.Encode()

	req.SetBasicAuth(user, pass)

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	return nil
}

// Check if database exists
// workaround: use http instead of client library because of the issue
// ref: https://github.com/influxdata/influxdb/issues/8108
func (c *clientConnection) dbExists(u *url.URL, user, pass, db string) bool {
	urlStr := fmt.Sprintf("%s/query", u.String())

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return false
	}

	query := fmt.Sprintf("SHOW DATABASES")
	params := req.URL.Query()
	params.Set("q", query)
	req.URL.RawQuery = params.Encode()

	req.SetBasicAuth(user, pass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}

	if resp.Body != nil {
		bodyText, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false
		}

		if strings.Contains(string(bodyText), db) {
			return true
		}
		resp.Body.Close()
	}
	return false
}

// Map the batch points write into client.Client
func (c *clientConnection) write(bps client.BatchPoints) error {
	return (*c.Conn).Write(bps)
}

// Map the close function into client.Client
func (c *clientConnection) closeClientConnection() error {
	return (*c.Conn).Close()
}

func selectClientConnection(config configuration) (*clientConnection, error) {
	// This is not an ideal way to get the logger but deferring solving this for a later date
	logger := getLogger(config)

	scheme := config.scheme

	u, err := url.Parse(fmt.Sprintf("%s://%s:%d", scheme, config.host, config.port))
	if err != nil {
		logger.Error("Error parsing URL")
		return nil, err
	}

	// Pool changes need to be safe (read & write) since the plugin can be called concurrently by snapteld.
	m.Lock()
	defer m.Unlock()

	user := config.user
	pass := config.password
	db := config.database
	key := connectionKey(u, user, db)

	// Do we have a existing client?
	if connPool[key] == nil {
		// create one and add to the pool
		var con client.Client
		var err error
		if scheme != UDP {
			con, err = client.NewHTTPClient(client.HTTPConfig{
				Addr:               u.String(),
				Username:           user,
				Password:           pass,
				InsecureSkipVerify: config.skipVerify,
			})
		} else {
			con, err = client.NewUDPClient(client.UDPConfig{
				Addr: u.Host,
			})
		}

		if err != nil {
			return nil, err
		}

		cCon := &clientConnection{
			Key:      key,
			Conn:     &con,
			LastUsed: time.Now(),
		}
		if !initialized && scheme != UDP {
			err = cCon.initDB(u, user, pass, db)
			if err != nil {
				return nil, err
			}
			if cCon.dbExists(u, user, pass, db) {
				initialized = true
			}
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

func replaceDynamicElement(m plugin.Metric) ([]string, map[string]string) {
	tags := map[string]string{}
	ns := m.Namespace.Strings()

	isDynamic, indexes := m.Namespace.IsDynamic()
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
			tags[m.Namespace[j].Name] = m.Namespace[j].Value
		}
	}
	return ns, tags
}

// groupCommonNamespaces groups common namespaces, those that differ at the leaf, into one data point with multiple influx fields.
func groupCommonNamespaces(m plugin.Metric, tags map[string]string, mpoints map[string]point) {
	// Slices to the second to last
	elems, tag := replaceDynamicElement(m)
	s2l := elems[:len(elems)-1]
	if len(s2l) == 0 {
		s2l = elems
	}

	// Appends tag keys
	mkeys := []string{}
	for k, v := range tags {
		tag[k] = v
		mkeys = append(mkeys, k, v)
	}
	// Appends namespace prefix
	mkeys = append(mkeys, s2l...)

	// Converts the map keys to a string key
	sk := strings.Join(mkeys, separator)

	// Groups fields by the namespace common prefix and tags
	fieldName := elems[len(elems)-1]
	if p, ok := mpoints[sk]; !ok {
		mpoints[sk] = point{
			ns:     s2l,
			tags:   tag,
			ts:     m.Timestamp,
			fields: map[string]interface{}{fieldName: m.Data},
		}
	} else {
		p.fields[fieldName] = m.Data
	}
}
