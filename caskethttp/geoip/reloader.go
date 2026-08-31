package geoip

import (
	"errors"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/oschwald/maxminddb-golang"
)

const defaultDebounce = 500 * time.Millisecond

// DBReloader manages a maxminddb.Reader that automatically reloads
// when the underlying database file changes on disk. It watches the
// directory containing the database file for changes (to handle both
// in-place writes and atomic rename replacements) and debounces
// rapid successive events before performing a reload.
//
// All lookups are protected by an RWMutex to ensure that in-flight
// lookups complete before an old reader is closed during a reload.
type DBReloader struct {
	dbPath      string
	mu          sync.RWMutex
	reader      *maxminddb.Reader
	closed      bool
	watcher     *fsnotify.Watcher
	done        chan struct{}
	wg          sync.WaitGroup
	debounce    time.Duration
	onReload    func() // optional callback, primarily for testing
	reloadCount atomic.Int64
}

// NewDBReloader opens the database at dbPath and starts watching for file changes.
// The caller must call Close when the reloader is no longer needed.
func NewDBReloader(dbPath string) (*DBReloader, error) {
	return newDBReloader(dbPath, defaultDebounce, nil)
}

func newDBReloader(dbPath string, debounce time.Duration, onReload func()) (*DBReloader, error) {
	reader, err := maxminddb.Open(dbPath)
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		reader.Close()
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		reader.Close()
		return nil, err
	}

	r := &DBReloader{
		dbPath:   absPath,
		reader:   reader,
		watcher:  watcher,
		done:     make(chan struct{}),
		debounce: debounce,
		onReload: onReload,
	}

	// Watch the directory containing the DB file so we also catch
	// atomic rename replacements (which appear as Create events).
	dir := filepath.Dir(absPath)
	if err := watcher.Add(dir); err != nil {
		reader.Close()
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

// Lookup performs a geo-IP lookup using the current database reader.
// It is safe for concurrent use and will not conflict with database reloads.
func (r *DBReloader) Lookup(ip net.IP, result interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.reader == nil {
		return errors.New("geoip: database is closed")
	}
	return r.reader.Lookup(ip, result)
}

// ReloadCount returns the number of times the database has been successfully reloaded.
func (r *DBReloader) ReloadCount() int64 {
	return r.reloadCount.Load()
}

// Close stops the file watcher, waits for the watch goroutine to exit,
// and closes the current database reader.
func (r *DBReloader) Close() error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	reader := r.reader
	r.reader = nil
	r.mu.Unlock()

	close(r.done)
	r.watcher.Close()
	r.wg.Wait()

	if reader != nil {
		return reader.Close()
	}
	return nil
}

func (r *DBReloader) watch() {
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

			// Only react to events for our specific database file.
			absEvent, err := filepath.Abs(event.Name)
			if err != nil || absEvent != r.dbPath {
				continue
			}

			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			// Debounce: reset the timer on each qualifying event so that
			// rapid successive writes (common during file copy) coalesce
			// into a single reload.
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
			log.Printf("geoip: watcher error: %v", err)
		}
	}
}

func (r *DBReloader) reload() {
	if _, err := os.Stat(r.dbPath); err != nil {
		log.Printf("geoip: cannot stat database file for reload: %v", err)
		return
	}

	newReader, err := maxminddb.Open(r.dbPath)
	if err != nil {
		log.Printf("geoip: failed to reload database: %v", err)
		return
	}

	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		newReader.Close()
		return
	}
	oldReader := r.reader
	r.reader = newReader
	r.mu.Unlock()

	if oldReader != nil {
		if err := oldReader.Close(); err != nil {
			log.Printf("geoip: error closing old database: %v", err)
		}
	}

	r.reloadCount.Add(1)
	log.Printf("geoip: successfully reloaded database from %s", r.dbPath)

	if r.onReload != nil {
		r.onReload()
	}
}
