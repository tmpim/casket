package ipfilter

import (
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func writePrefixFile(t *testing.T, name, contents string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatalf("failed to create directory for %s: %v", name, err)
	}
	if err := os.WriteFile(name, []byte(contents), 0o600); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
	return name
}

func setupPrefixFileDir(t *testing.T, name string) (dir, activeFile string) {
	t.Helper()
	dir = filepath.Join("testdata", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create test directory %s: %v", dir, err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	activeFile = filepath.Join(dir, "active.prefixes")
	writePrefixFile(t, activeFile, "192.168.1.2\n10.0.0.0/8\n# comment\n; another\n/ ignored\n")
	return dir, activeFile
}

func TestParsePrefixFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "prefixes")
	path := filepath.Join(dir, "rules.txt")
	writePrefixFile(t, path, "192.168.1.2\r\n\n# ignored\r\n10.0.0.0/8\r\n; skip\r\n/ skip\r\n")

	nets, err := parsePrefixFile(path)
	if err != nil {
		t.Fatalf("parsePrefixFile returned error: %v", err)
	}
	if len(nets) != 2 {
		t.Fatalf("expected 2 networks, got %d", len(nets))
	}

	if !nets[0].Contains(net.ParseIP("192.168.1.2")) {
		t.Fatalf("expected 192.168.1.2 to match first network")
	}
	if !nets[1].Contains(net.ParseIP("10.0.0.1")) {
		t.Fatalf("expected 10.0.0.1 to match second network")
	}
}

func TestPrefixFileReloaderBasicReload(t *testing.T) {
	_, activeFile := setupPrefixFileDir(t, "reload-basic")

	reloaded := make(chan struct{}, 1)
	reloader, err := newPrefixFileReloader(activeFile, 100*time.Millisecond, func() {
		select {
		case reloaded <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("failed to create reloader: %v", err)
	}
	defer reloader.Close()

	if !reloader.Contains(net.ParseIP("192.168.1.2")) {
		t.Fatal("expected initial prefix to match")
	}
	initialReloads := reloader.ReloadCount()

	writePrefixFile(t, activeFile, "243.1.3.15\r\n# changed\n")

	select {
	case <-reloaded:
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for reload")
	}

	if got := reloader.ReloadCount(); got != initialReloads+1 {
		t.Fatalf("expected reload count %d, got %d", initialReloads+1, got)
	}
	if reloader.Contains(net.ParseIP("243.1.3.15")) == false {
		t.Fatal("expected reloaded prefix to match")
	}
	if reloader.Contains(net.ParseIP("192.168.1.2")) {
		t.Fatal("expected old prefix to be removed after reload")
	}
}

func TestPrefixFileReloaderMultipleReloads(t *testing.T) {
	_, activeFile := setupPrefixFileDir(t, "reload-multi")

	reloaded := make(chan struct{}, 10)
	reloader, err := newPrefixFileReloader(activeFile, 100*time.Millisecond, func() {
		select {
		case reloaded <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("failed to create reloader: %v", err)
	}
	defer reloader.Close()

	sources := []string{
		"8.8.8.8\n",
		"8.8.4.4\n",
		"8.8.8.8\n",
	}
	for i, src := range sources {
		writePrefixFile(t, activeFile, src)
		select {
		case <-reloaded:
		case <-time.After(10 * time.Second):
			t.Fatalf("timed out waiting for reload #%d", i+1)
		}
	}

	if got := reloader.ReloadCount(); got != 3 {
		t.Fatalf("expected 3 reloads, got %d", got)
	}
	if !reloader.Contains(net.ParseIP("8.8.8.8")) {
		t.Fatal("expected 8.8.8.8 to match after multiple reloads")
	}
}

func TestPrefixFileReloaderConcurrentAccess(t *testing.T) {
	_, activeFile := setupPrefixFileDir(t, "reload-concurrent")

	reloaded := make(chan struct{}, 10)
	reloader, err := newPrefixFileReloader(activeFile, 100*time.Millisecond, func() {
		select {
		case reloaded <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("failed to create reloader: %v", err)
	}
	defer reloader.Close()

	const numReaders = 10
	var wg sync.WaitGroup
	done := make(chan struct{})
	errCh := make(chan error, numReaders)

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					if !reloader.Contains(net.ParseIP("192.168.1.2")) && !reloader.Contains(net.ParseIP("10.0.0.1")) && !reloader.Contains(net.ParseIP("8.8.8.8")) {
						errCh <- nil
						return
					}
				}
			}
		}()
	}

	for i := 0; i < 3; i++ {
		writePrefixFile(t, activeFile, "8.8.8.8\n")
		select {
		case <-reloaded:
		case <-time.After(10 * time.Second):
			close(done)
			wg.Wait()
			t.Fatalf("timed out waiting for reload #%d", i+1)
		}
		time.Sleep(50 * time.Millisecond)
	}

	close(done)
	wg.Wait()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("reader error: %v", err)
		}
	default:
	}
}

func TestPrefixFileReloaderClose(t *testing.T) {
	_, activeFile := setupPrefixFileDir(t, "reload-close")

	reloader, err := newPrefixFileReloader(activeFile, 100*time.Millisecond, nil)
	if err != nil {
		t.Fatalf("failed to create reloader: %v", err)
	}

	if err := reloader.Close(); err != nil {
		t.Fatalf("close returned error: %v", err)
	}
	if err := reloader.Close(); err != nil {
		t.Fatalf("double close returned error: %v", err)
	}

	if reloader.Contains(net.ParseIP("192.168.1.2")) {
		t.Fatal("expected no matching prefix after close")
	}
}

func TestPrefixFileReloaderAtomicRename(t *testing.T) {
	dir, activeFile := setupPrefixFileDir(t, "reload-rename")

	reloaded := make(chan struct{}, 1)
	reloader, err := newPrefixFileReloader(activeFile, 100*time.Millisecond, func() {
		select {
		case reloaded <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("failed to create reloader: %v", err)
	}
	defer reloader.Close()

	tmpFile := filepath.Join(dir, "temp.prefixes")
	writePrefixFile(t, tmpFile, "8.8.4.4\n")
	if err := os.Rename(tmpFile, activeFile); err != nil {
		t.Fatalf("rename failed: %v", err)
	}

	select {
	case <-reloaded:
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for reload after atomic rename")
	}

	if !reloader.Contains(net.ParseIP("8.8.4.4")) {
		t.Fatal("expected renamed prefix to match")
	}
	if reloader.Contains(net.ParseIP("192.168.1.2")) {
		t.Fatal("expected old prefix to be removed after rename")
	}
}
