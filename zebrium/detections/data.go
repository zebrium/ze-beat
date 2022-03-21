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

package detections

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/metricbeat/mb"
)

type EnumInciSummaryInci_significance string

const (
	EnumInciSummaryInci_significancelow    EnumInciSummaryInci_significance = "low"
	EnumInciSummaryInci_significancemedium EnumInciSummaryInci_significance = "medium"
	EnumInciSummaryInci_significancehigh   EnumInciSummaryInci_significance = "high"
)

type Word struct {
	B int64  `json:"b" db:"b" validate:"nonzero"`
	S int64  `json:"s" db:"s" validate:"nonzero"`
	W string `json:"w" db:"w" validate:"nonzero"`
}

type InciSummary struct {
	Deployment_id     string                           `json:"deployment_id" db:"deployment_id" validate:"nonzero"`
	Inci_id           string                           `json:"inci_id" db:"inci_id" validate:"nonzero"`
	Inci_itype_occ    int64                            `json:"inci_itype_occ" db:"inci_itype_occ" validate:"nonzero"`
	Inci_report_url   string                           `json:"inci_report_url" db:"inci_report_url" validate:"nonzero"`
	Inci_significance EnumInciSummaryInci_significance `json:"inci_significance" db:"inci_significance" validate:"regexp=^(low|medium|high)$,nonzero"`
	Inci_svc_grps     string                           `json:"inci_svc_grps" db:"inci_svc_grps" validate:"nonzero"`
	Inci_ts           time.Time                        `json:"inci_ts" db:"inci_ts" validate:"nonzero"`
	Inci_words        []Word                           `json:"inci_words" db:"inci_words" validate:"nonzero"`
	Inci_words_str    string                           `json:"inci_words_str" db:"inci_words_str" validate:"nonzero"`
	Itype_id          string                           `json:"itype_id" db:"itype_id" validate:"nonzero"`
	Itype_title       string                           `json:"itype_title" db:"itype_title" validate:"nonzero"`
}

type Error struct {
	Code       int64             `json:"code" db:"code" validate:"nonzero"`
	Data       []json.RawMessage `json:"data" db:"data" validate:"nonzero"`
	Message    string            `json:"message" db:"message" validate:"nonzero"`
	RetryAfter int64             `json:"retryAfter,omitempty" db:"retry_after"`
}

type Meta struct {
	Data            []InciSummary `json:"data" db:"data" validate:"nonzero"`
	Error           Error         `json:"error" db:"error" validate:"nonzero"`
	Op              string        `json:"op" db:"op" validate:"regexp=^(create|read|update|delete)$,nonzero"`
	SoftwareRelease string        `json:"softwareRelease" db:"software_release" validate:"nonzero"`
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
		if bucket.Inci_ts.After(*lastTs) {
			*lastTs = bucket.Inci_ts
		}
		fmt.Println(bucket)
		event := mb.Event{
			Namespace: "detections",
			Timestamp: bucket.Inci_ts,
			MetricSetFields: common.MapStr{
				"deployment_id": bucket.Deployment_id,
				"title":         bucket.Itype_title,
				"significance":  bucket.Inci_significance,
				"report_url":    bucket.Inci_report_url,
				"occurrence":    common.MapStr{"count": bucket.Inci_itype_occ},
			},
		}
		serviceGroupFromGroups(event.MetricSetFields, bucket.Inci_svc_grps)
		words(event.MetricSetFields, bucket.Inci_words)
		events = append(events, event)
	}
	return events, nil
}

func serviceGroupFromGroups(base common.MapStr, inci_svc_grps string) {
	svc_grp := ""
	includes_default := false
	for _, grp := range strings.Split(inci_svc_grps, ",") {
		switch grp {
		case "default":
			includes_default = true
			if svc_grp == "" {
				svc_grp = grp
			}
		default:
			svc_grp = grp
		}
	}
	base.Put("includes_default", includes_default)
	base.Put("service_group", svc_grp)
}

func words(base common.MapStr, words []Word) {
	for _, word := range words {
		base.Put("work_cloud", common.MapStr{
			"b": word.B,
			"s": word.S,
			"w": word.W,
		})
	}
}
