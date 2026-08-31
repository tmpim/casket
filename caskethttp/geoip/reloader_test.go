package geoip

import (
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// copyFile copies src to dst, creating or truncating dst.
func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	in, err := os.Open(src)
	if err != nil {
		t.Fatalf("Failed to open source file %s: %v", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		t.Fatalf("Failed to create destination file %s: %v", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		t.Fatalf("Failed to copy %s -> %s: %v", src, dst, err)
	}
	if err := out.Sync(); err != nil {
		t.Fatalf("Failed to sync %s: %v", dst, err)
	}
}

// swapFile replaces dst with the contents of src using an atomic rename.
// This is the correct update pattern: on Windows a memory-mapped file
// cannot be overwritten directly, but can be replaced via rename.
func swapFile(t *testing.T, src, dst string) {
	t.Helper()
	tmp := dst + ".tmp"
	copyFile(t, src, tmp)
	if err := os.Rename(tmp, dst); err != nil {
		t.Fatalf("Failed to rename %s -> %s: %v", tmp, dst, err)
	}
}

// setupTestDir creates a subdirectory inside the gitignored db/ directory
// and copies the initial database into it. Returns the directory path and
// the path to the active database file.
func setupTestDir(t *testing.T, name string) (dir, activeDB string) {
	t.Helper()
	dir = filepath.Join("db", name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create test directory %s: %v", dir, err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	activeDB = filepath.Join(dir, "active.mmdb")
	copyFile(t, "./test-data/GeoLite2-City.mmdb", activeDB)
	return dir, activeDB
}

func TestDBReloaderBasicReload(t *testing.T) {
	_, activeDB := setupTestDir(t, "reload-basic")

	reloaded := make(chan struct{}, 1)
	reloader, err := newDBReloader(activeDB, 100*time.Millisecond, func() {
		select {
		case reloaded <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("Failed to create reloader: %v", err)
	}
	defer reloader.Close()

	// Verify initial lookup works (original test DB → Cyprus).
	ip := net.ParseIP("212.50.99.193")
	var record GeoIPRecord
	if err := reloader.Lookup(ip, &record); err != nil {
		t.Fatalf("Initial lookup failed: %v", err)
	}
	if record.Country.ISOCode != "CY" {
		t.Fatalf("Expected initial country code CY, got %s", record.Country.ISOCode)
	}
	initialReloads := reloader.ReloadCount()

	// Replace the database with the live copy to trigger a reload.
	swapFile(t, "./test-data/GeoLite2-City-Live.mmdb", activeDB)

	select {
	case <-reloaded:
		// OK
	case <-time.After(10 * time.Second):
		t.Fatal("Timed out waiting for database reload")
	}

	if got := reloader.ReloadCount(); got != initialReloads+1 {
		t.Fatalf("Expected reload count %d, got %d", initialReloads+1, got)
	}

	// Verify lookups still work after the reload.
	var newRecord GeoIPRecord
	if err := reloader.Lookup(ip, &newRecord); err != nil {
		t.Fatalf("Post-reload lookup failed: %v", err)
	}
	if newRecord.Country.ISOCode != "CY" {
		t.Fatalf("Expected country code CY after reload, got %s", newRecord.Country.ISOCode)
	}
}

func TestDBReloaderMultipleReloads(t *testing.T) {
	_, activeDB := setupTestDir(t, "reload-multi")

	reloaded := make(chan struct{}, 10)
	reloader, err := newDBReloader(activeDB, 100*time.Millisecond, func() {
		select {
		case reloaded <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("Failed to create reloader: %v", err)
	}
	defer reloader.Close()

	sources := []string{
		"./test-data/GeoLite2-City-Live.mmdb",
		"./test-data/GeoLite2-City.mmdb",
		"./test-data/GeoLite2-City-Live.mmdb",
	}

	for i, src := range sources {
		swapFile(t, src, activeDB)

		select {
		case <-reloaded:
			// OK
		case <-time.After(10 * time.Second):
			t.Fatalf("Timed out waiting for reload #%d", i+1)
		}
	}

	if got := reloader.ReloadCount(); got != 3 {
		t.Fatalf("Expected 3 reloads, got %d", got)
	}

	// Lookups should still work after multiple reloads.
	ip := net.ParseIP("212.50.99.193")
	var record GeoIPRecord
	if err := reloader.Lookup(ip, &record); err != nil {
		t.Fatalf("Lookup after multiple reloads failed: %v", err)
	}
	if record.Country.ISOCode != "CY" {
		t.Fatalf("Expected CY, got %s", record.Country.ISOCode)
	}
}

func TestDBReloaderConcurrentAccess(t *testing.T) {
	_, activeDB := setupTestDir(t, "reload-concurrent")

	reloaded := make(chan struct{}, 10)
	reloader, err := newDBReloader(activeDB, 100*time.Millisecond, func() {
		select {
		case reloaded <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("Failed to create reloader: %v", err)
	}
	defer reloader.Close()

	ip := net.ParseIP("212.50.99.193")

	// Start concurrent readers that continuously perform lookups.
	const numReaders = 10
	var wg sync.WaitGroup
	done := make(chan struct{})
	errs := make(chan error, numReaders)

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					var record GeoIPRecord
					if err := reloader.Lookup(ip, &record); err != nil {
						errs <- err
						return
					}
					// The IP is always in Cyprus regardless of which DB version.
					if record.Country.ISOCode != "CY" && record.Country.ISOCode != "" {
						t.Errorf("Unexpected country code during concurrent read: %s", record.Country.ISOCode)
						return
					}
				}
			}
		}()
	}

	// Trigger multiple reloads while concurrent readers are active.
	for i := 0; i < 3; i++ {
		swapFile(t, "./test-data/GeoLite2-City-Live.mmdb", activeDB)

		select {
		case <-reloaded:
			// OK
		case <-time.After(10 * time.Second):
			close(done)
			wg.Wait()
			t.Fatalf("Timed out waiting for reload #%d", i+1)
		}

		time.Sleep(50 * time.Millisecond)
	}

	close(done)
	wg.Wait()

	select {
	case err := <-errs:
		t.Fatalf("Concurrent reader encountered error: %v", err)
	default:
		// No errors.
	}
}

func TestDBReloaderClose(t *testing.T) {
	_, activeDB := setupTestDir(t, "reload-close")

	reloader, err := newDBReloader(activeDB, 100*time.Millisecond, nil)
	if err != nil {
		t.Fatalf("Failed to create reloader: %v", err)
	}

	// Close should succeed.
	if err := reloader.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	// Double close should be safe.
	if err := reloader.Close(); err != nil {
		t.Fatalf("Double close returned error: %v", err)
	}

	// Lookup after close should return an error.
	ip := net.ParseIP("212.50.99.193")
	var record GeoIPRecord
	if err := reloader.Lookup(ip, &record); err == nil {
		t.Fatal("Expected error on Lookup after Close, got nil")
	}
}

func TestDBReloaderAtomicRename(t *testing.T) {
	dir, activeDB := setupTestDir(t, "reload-rename")

	reloaded := make(chan struct{}, 1)
	reloader, err := newDBReloader(activeDB, 100*time.Millisecond, func() {
		select {
		case reloaded <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("Failed to create reloader: %v", err)
	}
	defer reloader.Close()

	// Simulate an atomic rename replacement:
	// 1. Write to a temp file in the same directory
	// 2. Rename over the active database
	tmpDB := filepath.Join(dir, "temp.mmdb")
	copyFile(t, "./test-data/GeoLite2-City-Live.mmdb", tmpDB)

	if err := os.Rename(tmpDB, activeDB); err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	select {
	case <-reloaded:
		// OK
	case <-time.After(10 * time.Second):
		t.Fatal("Timed out waiting for reload after atomic rename")
	}

	// Verify the reloaded database works.
	ip := net.ParseIP("212.50.99.193")
	var record GeoIPRecord
	if err := reloader.Lookup(ip, &record); err != nil {
		t.Fatalf("Lookup after atomic rename failed: %v", err)
	}
	if record.Country.ISOCode != "CY" {
		t.Fatalf("Expected CY, got %s", record.Country.ISOCode)
	}
}
