// Copyright 2015 Light Code Labs, LLC
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

package caskethttp

import (
	// plug in the server
	_ "github.com/tmpim/casket/caskethttp/httpserver"

	// plug in the standard directives
	_ "github.com/tmpim/casket/caskethttp/basicauth"
	_ "github.com/tmpim/casket/caskethttp/bind"
	_ "github.com/tmpim/casket/caskethttp/browse"
	_ "github.com/tmpim/casket/caskethttp/chuieauth"
	_ "github.com/tmpim/casket/caskethttp/cors/casket"
	_ "github.com/tmpim/casket/caskethttp/errors"
	_ "github.com/tmpim/casket/caskethttp/expvar"
	_ "github.com/tmpim/casket/caskethttp/extensions"
	_ "github.com/tmpim/casket/caskethttp/fastcgi"
	_ "github.com/tmpim/casket/caskethttp/forwardproxy"
	_ "github.com/tmpim/casket/caskethttp/geoip"
	_ "github.com/tmpim/casket/caskethttp/gzip"
	_ "github.com/tmpim/casket/caskethttp/header"
	_ "github.com/tmpim/casket/caskethttp/index"
	_ "github.com/tmpim/casket/caskethttp/internalsrv"
	_ "github.com/tmpim/casket/caskethttp/ipfilter"
	_ "github.com/tmpim/casket/caskethttp/limits"
	_ "github.com/tmpim/casket/caskethttp/log"
	_ "github.com/tmpim/casket/caskethttp/markdown"
	_ "github.com/tmpim/casket/caskethttp/mime"
	_ "github.com/tmpim/casket/caskethttp/pprof"
	_ "github.com/tmpim/casket/caskethttp/prometheus"
	_ "github.com/tmpim/casket/caskethttp/proxy"
	_ "github.com/tmpim/casket/caskethttp/push"
	_ "github.com/tmpim/casket/caskethttp/rate-limit"
	_ "github.com/tmpim/casket/caskethttp/realip"
	_ "github.com/tmpim/casket/caskethttp/redirect"
	_ "github.com/tmpim/casket/caskethttp/requestid"
	_ "github.com/tmpim/casket/caskethttp/rewrite"
	_ "github.com/tmpim/casket/caskethttp/root"
	_ "github.com/tmpim/casket/caskethttp/status"
	_ "github.com/tmpim/casket/caskethttp/templates"
	_ "github.com/tmpim/casket/caskethttp/timeouts"
	_ "github.com/tmpim/casket/caskethttp/tmpauth"
	_ "github.com/tmpim/casket/caskethttp/tryfiles"
	_ "github.com/tmpim/casket/caskethttp/webdav"
	_ "github.com/tmpim/casket/caskethttp/websocket"
	_ "github.com/tmpim/casket/onevent"

	// plug in the dns providers
	_ "github.com/tmpim/casket/dnsproviders/acmedns"
	_ "github.com/tmpim/casket/dnsproviders/azure"
	_ "github.com/tmpim/casket/dnsproviders/cloudflare"
	_ "github.com/tmpim/casket/dnsproviders/digitalocean"
	_ "github.com/tmpim/casket/dnsproviders/dnspod"
	_ "github.com/tmpim/casket/dnsproviders/duckdns"
	_ "github.com/tmpim/casket/dnsproviders/gandi"
	_ "github.com/tmpim/casket/dnsproviders/godaddy"
	_ "github.com/tmpim/casket/dnsproviders/googlecloud"
	_ "github.com/tmpim/casket/dnsproviders/linode"
	_ "github.com/tmpim/casket/dnsproviders/namecheap"
	_ "github.com/tmpim/casket/dnsproviders/namedotcom"
	_ "github.com/tmpim/casket/dnsproviders/ovh"
	_ "github.com/tmpim/casket/dnsproviders/pdns"
	_ "github.com/tmpim/casket/dnsproviders/rfc2136"
	_ "github.com/tmpim/casket/dnsproviders/route53"
	_ "github.com/tmpim/casket/dnsproviders/transip"
	_ "github.com/tmpim/casket/dnsproviders/vultr"
)
