package environment

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestEnsureEnvFileConcurrentCreation(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	const workers = 16

	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- ensureEnvFile(path)
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Errorf("ensureEnvFile returned an error: %v", err)
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated env file: %v", err)
	}
	if strings.Count(string(content), "JWT_KEY=") != 1 || !strings.Contains(string(content), "DB_PATH=app_data.db") {
		t.Fatalf("generated env file is incomplete: %q", content)
	}
}
