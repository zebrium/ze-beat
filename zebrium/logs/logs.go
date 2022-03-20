// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package logs

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/elastic/beats/v7/metricbeat/mb"
	"github.com/elastic/beats/v7/metricbeat/module/zebrium"
)

func init() {
	mb.Registry.MustAddMetricSet("zebrium", "logs", New,
		mb.WithHostParser(zebrium.HostParser),
		mb.DefaultMetricSet(),
	)
}

// MetricSet for fetching RabbitMQ logs metrics
type MetricSet struct {
	*zebrium.MetricSet
	lastTs []time.Time
}

// New creates new instance of MetricSet
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	config := defaultConfig
	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	ms, err := zebrium.NewMetricSet(base, zebrium.LogsPath)
	if err != nil {
		return nil, err
	}
	lastTs := make([]time.Time, len(ms.AccessTokens))
	for idx, _ := range ms.AccessTokens {
		lastTs[idx] = time.Now().UTC()
	}
	return &MetricSet{ms, lastTs}, nil
}

// Fetch metrics from zebrium logs
func (m *MetricSet) Fetch(r mb.ReporterV2) error {
	for idx, token := range m.AccessTokens {
		m.HTTP.SetHeader("Authentication", "Bearer "+token)
		m.HTTP.SetBody([]byte(fmt.Sprintf(`{"filter":["mi>%s","mi<%s"],"sort":["mi"]}`, m.lastTs[idx].Format(time.RFC3339Nano), time.Now().Add(-3*time.Minute).Format(time.RFC3339Nano))))
		content, err := m.HTTP.FetchContent()
		if err != nil {
			return errors.Wrap(err, "error in fetch")
		}
		evts, err := eventMapping(content, &m.lastTs[idx])
		if err != nil {
			return err
		}
		for _, evt := range evts {
			r.Event(evt)
		}
	}
	return nil
}
