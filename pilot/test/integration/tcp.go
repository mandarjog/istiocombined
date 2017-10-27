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

package main

import (
	"fmt"

	proxyconfig "istio.io/api/proxy/v1/config"
	"istio.io/istio/pilot/platform"
)

type tcp struct {
	*infra
}

func (t *tcp) String() string {
	return "tcp-reachability"
}

func (t *tcp) setup() error {
	return nil
}

func (t *tcp) teardown() {
}

func (t *tcp) run() error {
	// TCP in Eureka is tested by the headless service test.
	if platform.ServiceRegistry(t.Registry) == platform.EurekaRegistry {
		return nil
	}
	srcPods := []string{"a", "b", "t"}
	dstPods := []string{"a", "b"}
	if t.Auth == proxyconfig.MeshConfig_NONE {
		// t is not behind proxy, so it cannot talk in Istio auth.
		dstPods = append(dstPods, "t")
	}
	funcs := make(map[string]func() status)
	for _, src := range srcPods {
		for _, dst := range dstPods {
			if src == "t" && dst == "t" {
				// this is flaky in minikube
				continue
			}
			for _, port := range []string{":90", ":9090"} {
				for _, domain := range []string{"", "." + t.Namespace} {
					name := fmt.Sprintf("TCP connection from %s to %s%s%s", src, dst, domain, port)
					funcs[name] = (func(src, dst, port, domain string) func() status {
						url := fmt.Sprintf("http://%s%s%s/%s", dst, domain, port, src)
						return func() status {
							resp := t.clientRequest(src, url, 1, "")
							if t.Auth == proxyconfig.MeshConfig_MUTUAL_TLS && src == "t" {
								// t cannot talk to envoy (a or b) with mTLS enabled.
								if len(resp.code) == 0 || resp.code[0] != httpOk {
									return nil
								}
							} else if len(resp.code) > 0 && resp.code[0] == httpOk {
								return nil
							}
							return errAgain
						}
					})(src, dst, port, domain)
				}
			}
		}
	}
	return parallel(funcs)
}
