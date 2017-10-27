// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"

	proxyconfig "istio.io/api/proxy/v1/config"
	"istio.io/istio/pilot/model"
)

func TestProtoSchemaConversions(t *testing.T) {
	routeRuleSchema := &model.ProtoSchema{MessageName: model.RouteRule.MessageName}

	msg := &proxyconfig.RouteRule{
		Destination: &proxyconfig.IstioService{
			Name: "foo",
		},
		Precedence: 5,
		Route: []*proxyconfig.DestinationWeight{
			{Destination: &proxyconfig.IstioService{Name: "bar"}, Weight: 75},
			{Destination: &proxyconfig.IstioService{Name: "baz"}, Weight: 25},
		},
	}

	wantJSON := `
	{
		"destination": {
			"name": "foo"
		},
		"precedence": 5,
		"route": [
		{
			"destination": {
				"name" : "bar"
			},
			"weight": 75
		},
		{
			"destination": {
				"name" : "baz"
			},
			"weight": 25
		}
		]
	}
	`

	wantYAML := "destination:\n" +
		"  name: foo\n" +
		"precedence: 5\n" +
		"route:\n" +
		"- destination:\n" +
		"    name: bar\n" +
		"  weight: 75\n" +
		"- destination:\n" +
		"    name: baz\n" +
		"  weight: 25\n"

	wantJSONMap := map[string]interface{}{
		"destination": map[string]interface{}{
			"name": "foo",
		},
		"precedence": 5.0,
		"route": []interface{}{
			map[string]interface{}{
				"destination": map[string]interface{}{
					"name": "bar",
				},
				"weight": 75.0,
			},
			map[string]interface{}{
				"destination": map[string]interface{}{
					"name": "baz",
				},
				"weight": 25.0,
			},
		},
	}

	badSchema := &model.ProtoSchema{MessageName: "bad-name"}
	if _, err := badSchema.FromYAML(wantYAML); err == nil {
		t.Errorf("FromYAML should have failed using ProtoSchema with bad MessageName")
	}

	gotJSON, err := model.ToJSON(msg)

	if err != nil {
		t.Errorf("ToJSON failed: %v", err)
	}

	if gotJSON != strings.Join(strings.Fields(wantJSON), "") {
		t.Errorf("ToJSON failed: got %s, want %s", gotJSON, wantJSON)
	}

	if _, err = model.ToJSON(nil); err == nil {
		t.Error("should produce an error")
	}

	gotFromJSON, err := routeRuleSchema.FromJSON(wantJSON)

	if err != nil {
		t.Errorf("FromJSON failed: %v", err)
	}

	if !reflect.DeepEqual(gotFromJSON, msg) {
		t.Errorf("FromYAML failed: got %+v want %+v", spew.Sdump(gotFromJSON), spew.Sdump(msg))
	}

	gotYAML, err := model.ToYAML(msg)

	if err != nil {
		t.Errorf("ToYAML failed: %v", err)
	}

	if !reflect.DeepEqual(gotYAML, wantYAML) {
		t.Errorf("ToYAML failed: got %+v want %+v", spew.Sdump(gotYAML), spew.Sdump(wantYAML))
	}

	if _, err = model.ToYAML(nil); err == nil {
		t.Error("should produce an error")
	}

	gotFromYAML, err := routeRuleSchema.FromYAML(wantYAML)

	if err != nil {
		t.Errorf("FromYAML failed: %v", err)
	}

	if !reflect.DeepEqual(gotFromYAML, msg) {
		t.Errorf("FromYAML failed: got %+v want %+v", spew.Sdump(gotFromYAML), spew.Sdump(msg))
	}

	if _, err = routeRuleSchema.FromYAML(":"); err == nil {
		t.Errorf("should produce an error")
	}

	gotJSONMap, err := model.ToJSONMap(msg)

	if err != nil {
		t.Errorf("ToJSONMap failed: %v", err)
	}

	if !reflect.DeepEqual(gotJSONMap, wantJSONMap) {
		t.Errorf("ToJSONMap failed: \ngot %vwant %v", spew.Sdump(gotJSONMap), spew.Sdump(wantJSONMap))
	}

	if _, err = model.ToJSONMap(nil); err == nil {
		t.Error("should produce an error")
	}

	gotFromJSONMap, err := routeRuleSchema.FromJSONMap(wantJSONMap)

	if err != nil {
		t.Errorf("FromJSONMap failed: %v", err)
	}

	if !reflect.DeepEqual(gotFromJSONMap, msg) {
		t.Errorf("FromJSONMap failed: got %+v want %+v", spew.Sdump(gotFromJSONMap), spew.Sdump(msg))
	}

	if _, err = routeRuleSchema.FromJSONMap(1); err == nil {
		t.Error("should produce an error")
	}

	if _, err = routeRuleSchema.FromJSON(":"); err == nil {
		t.Errorf("should produce an error")
	}
}
