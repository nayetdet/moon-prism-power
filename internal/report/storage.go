package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func (r Report) Write(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create report directory: %w", err)
	}

	temporary, err := os.CreateTemp(filepath.Dir(path), ".report-*.json")
	if err != nil {
		return fmt.Errorf("create temporary report: %w", err)
	}

	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)

	encoder := json.NewEncoder(temporary)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(r); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("encode report: %w", err)
	}

	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close report: %w", err)
	}

	if err := os.Rename(temporaryName, path); err != nil {
		return fmt.Errorf("store report: %w", err)
	}

	return nil
}
