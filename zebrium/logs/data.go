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
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/metricbeat/mb"
)

type StatsMi struct {
	Customer      string    `json:"customer" db:"customer" validate:"nonzero"`
	Deployment_id string    `json:"deployment_id" db:"deployment_id" validate:"nonzero"`
	Evt_ct        int64     `json:"evt_ct" db:"evt_ct" validate:"nonzero"`
	Evt_error_ct  int64     `json:"evt_error_ct" db:"evt_error_ct" validate:"nonzero"`
	Evt_rare_ct   int64     `json:"evt_rare_ct" db:"evt_rare_ct" validate:"nonzero"`
	Mi            time.Time `json:"mi" db:"mi" validate:"nonzero"`
	Svc_grp       string    `json:"svc_grp" db:"svc_grp" validate:"nonzero"`
}

type Error struct {
	Code       int64             `json:"code" db:"code" validate:"nonzero"`
	Data       []json.RawMessage `json:"data" db:"data" validate:"nonzero"`
	Message    string            `json:"message" db:"message" validate:"nonzero"`
	RetryAfter int64             `json:"retryAfter,omitempty" db:"retry_after"`
}

type Meta struct {
	Data            []StatsMi `json:"data" db:"data" validate:"nonzero"`
	Error           Error     `json:"error" db:"error" validate:"nonzero"`
	Op              string    `json:"op" db:"op" validate:"regexp=^(create|read|update|delete)$,nonzero"`
	SoftwareRelease string    `json:"softwareRelease" db:"software_release" validate:"nonzero"`
}

func eventMapping(content []byte, lastTs *time.Time) ([]mb.Event, error) {
	meta := Meta{}
	err := json.Unmarshal(content, &meta)
	if err != nil {
		return []mb.Event{}, errors.Wrap(err, "error in Unmarshal")
	}
	if len(meta.Data) > 0 {
		fmt.Println("")
	}

	events := []mb.Event{}
	for _, bucket := range meta.Data {
		if bucket.Mi.After(*lastTs) {
			*lastTs = bucket.Mi
		}
		fmt.Println(bucket)
		event := mb.Event{
			Namespace: "logs",
			Timestamp: bucket.Mi,
			MetricSetFields: common.MapStr{
				"deployment_id": bucket.Deployment_id,
				"service_group": bucket.Svc_grp,
				"total":         common.MapStr{"count": bucket.Evt_ct},
				"errors":        common.MapStr{"count": bucket.Evt_error_ct},
				"anamalies":     common.MapStr{"count": bucket.Evt_rare_ct},
			},
		}
		events = append(events, event)
	}
	return events, nil
}
