package ipfilter

import (
	"bufio"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

const defaultPrefixFileDebounce = 500 * time.Millisecond

var (
	prefixFileReloaderCache   = make(map[string]*PrefixFileReloader)
	prefixFileReloaderCacheMu sync.Mutex
)

type PrefixFileReloader struct {
	filePath    string
	mu          sync.RWMutex
	nets        []*net.IPNet
	closed      bool
	watcher     *fsnotify.Watcher
	done        chan struct{}
	wg          sync.WaitGroup
	debounce    time.Duration
	onReload    func()
	reloadCount atomic.Int64
}

func newPrefixFileReloader(filePath string, debounce time.Duration, onReload func()) (*PrefixFileReloader, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	filePath = absPath

	nets, err := parsePrefixFile(filePath)
	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	r := &PrefixFileReloader{
		filePath: filePath,
		nets:     nets,
		watcher:  watcher,
		done:     make(chan struct{}),
		debounce: debounce,
		onReload: onReload,
	}

	// Watch the directory containing the prefix file so we also catch
	// atomic rename replacements (which appear as Create events).
	if err := watcher.Add(filepath.Dir(filePath)); err != nil {
		watcher.Close()
		return nil, err
	}

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.watch()
	}()

	return r, nil
}

func (r *PrefixFileReloader) ReloadCount() int64 {
	return r.reloadCount.Load()
}

func (r *PrefixFileReloader) Close() error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	r.nets = nil
	r.mu.Unlock()

	close(r.done)
	if r.watcher != nil {
		r.watcher.Close()
	}
	r.wg.Wait()
	return nil
}

func (r *PrefixFileReloader) Contains(ip net.IP) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed {
		return false
	}

	for _, netRange := range r.nets {
		if netRange.Contains(ip) {
			return true
		}
	}

	return false
}

func (r *PrefixFileReloader) watch() {
	var debounceTimer *time.Timer
	var debounceCh <-chan time.Time

	for {
		select {
		case <-r.done:
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return

		case event, ok := <-r.watcher.Events:
			if !ok {
				return
			}

			// Only react to events for our specific prefix file.
			absEvent, err := filepath.Abs(event.Name)
			if err != nil || absEvent != r.filePath {
				continue
			}

			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			if debounceTimer != nil {
				debounceTimer.Stop()
			}

			debounceTimer = time.NewTimer(r.debounce)
			debounceCh = debounceTimer.C

		case <-debounceCh:
			r.reload()
			debounceCh = nil

		case err, ok := <-r.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("ipfilter: watcher error: %v", err)
		}
	}
}

func (r *PrefixFileReloader) reload() {
	if _, err := os.Stat(r.filePath); err != nil {
		log.Printf("ipfilter: cannot stat prefix file for reload: %v", err)
		return
	}

	nets, err := parsePrefixFile(r.filePath)
	if err != nil {
		log.Printf("ipfilter: failed to reload database: %v", err)
		return
	}

	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.nets = nets
	r.mu.Unlock()

	if r.onReload != nil {
		r.onReload()
	}

	r.reloadCount.Add(1)
	log.Printf("ipfilter: successfully reloaded database from %s", r.filePath)
}

func skipLine(line string) bool {
	return line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "/")
}

func parsePrefixFile(path string) ([]*net.IPNet, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var nets []*net.IPNet
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		if skipLine(line) {
			continue
		}

		line = strings.Trim(line, "\r\n")
		line = strings.TrimSpace(line)
		if skipLine(line) {
			continue
		}

		ipRange, err := parseIP(line)
		if err != nil {
			return nil, err
		}

		nets = append(nets, ipRange...)
	}

	if err := scan.Err(); err != nil {
		return nil, err
	}

	return nets, nil
}
