package nmcslog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	nmcslog "github.com/notmycloud/slog"
)

func TestValidateSchema(t *testing.T) {
	exampleConfigsDir := filepath.Join("example", "configs")
	fileInfos, err := os.ReadDir(exampleConfigsDir)
	if err != nil {
		t.Fatalf("reading example configs directory: %v", err)
	}
	if len(fileInfos) == 0 {
		t.Fatalf("no test files found")
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}

		fileName := fileInfo.Name()
		filePath := filepath.Join(exampleConfigsDir, fileName)

		expectedErr := false
		if strings.HasPrefix(fileName, "fail") {
			expectedErr = true
		} else if strings.HasPrefix(fileName, "pass") {
			expectedErr = false
		} else {
			t.Errorf("test config has an invalid prefix [%s]", fileName)
			continue
		}

		err := nmcslog.ValidateSchema(filePath)
		if expectedErr && err == nil {
			t.Errorf("expected error for config file %s, but no error occurred", fileName)
		} else if !expectedErr && err != nil {
			t.Errorf("unexpected error for config file %s: %v", fileName, err)
		}
	}
}
