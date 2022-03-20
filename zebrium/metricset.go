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

package zebrium

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"github.com/elastic/beats/v7/metricbeat/helper"
	"github.com/elastic/beats/v7/metricbeat/mb"
	"github.com/pkg/errors"
)

// MetricSet can be used to build other metric sets that query Zebrium
// management plugin
type MetricSet struct {
	mb.BaseMetricSet
	*helper.HTTP
	AccessTokens []string
}

// NewMetricSet creates an metric set that can be used to build other metric
// sets that query Zebrium management plugin
func NewMetricSet(base mb.BaseMetricSet, subPath string) (*MetricSet, error) {
	// Unpack additional configuration options.
	config := struct {
		Access_tokens_file string `config:"access_tokens_file"`
	}{}
	err := base.Module().UnpackConfig(&config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read configuration")
	}
	accessTokensMap, err := parseAccessTokensFile(config.Access_tokens_file)

	http, err := helper.NewHTTP(base)
	if err != nil {
		return nil, err
	}
	http.SetURI(http.GetURI() + subPath)
	http.SetHeader("Accept", "application/json")
	http.SetMethod("POST")

	accessTokens, ok := accessTokensMap[base.HostData().SanitizedURI]
	if !ok {
		return nil, fmt.Errorf("No access tokens for host %s in access tokens file", base.HostData().SanitizedURI)
	}

	return &MetricSet{
		base,
		http,
		accessTokens,
	}, nil
}

func parseAccessTokensFile(fileName string) (map[string][]string, error) {
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var yamlConfig map[string][]string
	err = yaml.Unmarshal(yamlFile, &yamlConfig)
	if err != nil {
		return nil, err
	}
	return yamlConfig, nil
}
