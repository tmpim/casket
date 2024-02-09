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

// By moving the application's package main logic into
// a package other than main, it becomes much easier to
// wrap casket for custom builds that are go-gettable.
// https://casket.community/t/my-wish-for-0-9-go-gettable-custom-builds/59?u=matt

package main

import (
	"github.com/tmpim/casket"
	"log"

	"github.com/tmpim/casket/casket/casketmain"
)

var (
	// Populated by ldflags (Docker builds) and goreleaser - https://goreleaser.com/cookbooks/using-main.version/
	version = "devel"
)

var run = casketmain.Run // replaced for tests

func main() {
	casket.AppVersion = version

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	run()
}
