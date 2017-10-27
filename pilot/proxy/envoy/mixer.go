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

// Mixer filter configuration

package envoy

import (
	proxyconfig "istio.io/api/proxy/v1/config"
	"istio.io/istio/pilot/proxy"
)

const (
	// MixerCluster is the name of the mixer cluster
	MixerCluster = "mixer_server"

	// MixerFilter name and its attributes
	MixerFilter = "mixer"

	// AttrSourceIP is client source IP
	AttrSourceIP = "source.ip"

	// AttrSourceUID is platform-specific unique identifier for the client instance of the source service
	AttrSourceUID = "source.uid"

	// AttrDestinationIP is the server source IP
	AttrDestinationIP = "destination.ip"

	// AttrDestinationUID is platform-specific unique identifier for the server instance of the target service
	AttrDestinationUID = "destination.uid"

	// AttrDestinationService is name of the target service
	AttrDestinationService = "destination.service"

	// MixerRequestCount is the quota bucket name
	MixerRequestCount = "RequestCount"

	// MixerCheck switches Check call on and off
	MixerCheck = "mixer_check"

	// MixerReport switches Report call on and off
	MixerReport = "mixer_report"

	// DisableTCPCheckCalls switches Check call on and off for tcp listeners
	DisableTCPCheckCalls = "disable_tcp_check_calls"

	// MixerForward switches attribute forwarding on and off
	MixerForward = "mixer_forward"
)

// FilterMixerConfig definition
type FilterMixerConfig struct {
	// MixerAttributes specifies the static list of attributes that are sent with
	// each request to Mixer.
	MixerAttributes map[string]string `json:"mixer_attributes,omitempty"`

	// ForwardAttributes specifies the list of attribute keys and values that
	// are forwarded as an HTTP header to the server side proxy
	ForwardAttributes map[string]string `json:"forward_attributes,omitempty"`

	// QuotaName specifies the name of the quota bucket to withdraw tokens from;
	// an empty name means no quota will be charged.
	QuotaName string `json:"quota_name,omitempty"`

	// If set to true, disables mixer check calls for TCP connections
	DisableTCPCheckCalls bool `json:"disable_tcp_check_calls,omitempty"`
}

func (*FilterMixerConfig) isNetworkFilterConfig() {}

// buildMixerCluster builds an outbound mixer cluster
func buildMixerCluster(mesh *proxyconfig.MeshConfig, role proxy.Node, mixerSAN []string) *Cluster {
	mixerCluster := buildCluster(mesh.MixerAddress, MixerCluster, mesh.ConnectTimeout)
	mixerCluster.CircuitBreaker = &CircuitBreaker{
		Default: DefaultCBPriority{
			MaxPendingRequests: 10000,
			MaxRequests:        10000,
		},
	}
	mixerCluster.Features = ClusterFeatureHTTP2

	// apply auth policies
	switch mesh.DefaultConfig.ControlPlaneAuthPolicy {
	case proxyconfig.AuthenticationPolicy_NONE:
		// do nothing
	case proxyconfig.AuthenticationPolicy_MUTUAL_TLS:
		// apply SSL context to enable mutual TLS between Envoy proxies between app and mixer
		mixerCluster.SSLContext = buildClusterSSLContext(proxy.AuthCertsPath, mixerSAN)
	}

	return mixerCluster
}

func buildMixerOpaqueConfig(check, forward bool) map[string]string {
	keys := map[bool]string{true: "on", false: "off"}
	return map[string]string{
		MixerReport:  "on",
		MixerCheck:   keys[check],
		MixerForward: keys[forward],
	}
}

// Mixer filter uses outbound configuration by default (forward attributes,
// but not invoke check calls)
func mixerHTTPRouteConfig(role proxy.Node, service string) *FilterMixerConfig {
	r := &FilterMixerConfig{
		MixerAttributes: map[string]string{
			AttrDestinationIP:  role.IPAddress,
			AttrDestinationUID: "kubernetes://" + role.ID,
		},
		ForwardAttributes: map[string]string{
			AttrSourceIP:  role.IPAddress,
			AttrSourceUID: "kubernetes://" + role.ID,
		},
		QuotaName: MixerRequestCount,
	}
	if len(service) > 0 {
		r.MixerAttributes[AttrDestinationService] = service
	}

	return r
}

// Mixer TCP filter config for inbound requests
func mixerTCPConfig(role proxy.Node, check bool) *FilterMixerConfig {
	return &FilterMixerConfig{
		MixerAttributes: map[string]string{
			AttrDestinationIP:  role.IPAddress,
			AttrDestinationUID: "kubernetes://" + role.ID,
		},
		DisableTCPCheckCalls: !check,
	}
}
