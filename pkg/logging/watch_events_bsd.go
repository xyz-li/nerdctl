//go:build freebsd || openbsd || netbsd

/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package logging

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/containerd/log"
	"github.com/fsnotify/fsnotify"
)

// startTail wait for the next log write.
// the boolean value indicates if the log file was recreated;
// the error is error happens during waiting new logs.
func startTail(ctx context.Context, logName string, w *fsnotify.Watcher) (bool, error) {
	errRetry := 5
	for {
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("context cancelled")
		case e := <-w.Events:
			switch {
			case e.Has(fsnotify.Write):
				return false, nil
			case e.Op.Has(fsnotify.Rename):
				return filepath.Base(e.Name) == logName, nil
			default:
				log.L.Debugf("Received unexpected fsnotify event: %v, retrying", e)
			}
		case err := <-w.Errors:
			log.L.Debugf("Received fsnotify watch error, retrying unless no more retries left, retries: %d, error: %s", errRetry, err)
			if errRetry == 0 {
				return false, err
			}
			errRetry--
		case <-time.After(logForceCheckPeriod):
			return false, nil
		}
	}
}
