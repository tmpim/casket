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

//go:build !windows && !plan9 && !nacl && !js
// +build !windows,!plan9,!nacl,!js

package casket

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// trapSignalsPosix captures POSIX-only signals.
func trapSignalsPosix() {
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)

		for sig := range sigchan {
			// ignore SIGHUP; this signal is sometimes sent outside of the user's control
			switch sig {
			case syscall.SIGQUIT:
				log.Println("[INFO] SIGQUIT: Quitting process immediately")
				for _, f := range OnProcessExit {
					f() // only perform important cleanup actions
				}
				os.Exit(0)

			case syscall.SIGTERM:
				log.Println("[INFO] SIGTERM: Shutting down servers then terminating")
				exitCode := executeShutdownCallbacks("SIGTERM")
				for _, f := range OnProcessExit {
					f() // only perform important cleanup actions
				}
				err := Stop()
				if err != nil {
					log.Printf("[ERROR] SIGTERM stop: %v", err)
					exitCode = 3
				}

				os.Exit(exitCode)

			case syscall.SIGUSR1:
				log.Println("[INFO] SIGUSR1: Reloading")

				// Start with the existing Casketfile
				casketfileToUse, inst, err := getCurrentCasketfile()
				if err != nil {
					log.Printf("[ERROR] SIGUSR1: %v", err)
					continue
				}
				if loaderUsed.loader == nil {
					// This also should never happen
					log.Println("[ERROR] SIGUSR1: no Casketfile loader with which to reload Casketfile")
					continue
				}

				// Load the updated Casketfile
				newCasketfile, err := loaderUsed.loader.Load(inst.serverType)
				if err != nil {
					log.Printf("[ERROR] SIGUSR1: loading updated Casketfile: %v", err)
					continue
				}
				if newCasketfile != nil {
					casketfileToUse = newCasketfile
				}

				// Backup old event hooks
				oldEventHooks := cloneEventHooks()

				// Purge the old event hooks
				purgeEventHooks()

				// Kick off the restart; our work is done
				EmitEvent(InstanceRestartEvent, nil)
				_, err = inst.Restart(casketfileToUse)
				if err != nil {
					restoreEventHooks(oldEventHooks)

					log.Printf("[ERROR] SIGUSR1: %v", err)
				}

			case syscall.SIGUSR2:
				log.Println("[INFO] SIGUSR2: Upgrading")
				if err := Upgrade(); err != nil {
					log.Printf("[ERROR] SIGUSR2: upgrading: %v", err)
				}

			}
		}
	}()
}
