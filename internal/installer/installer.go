package installer

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"waybar-tui/internal/theme"
)

func Clone(url string) (string, error) {
	tmp, err := os.MkdirTemp("", "waytui-*")
	if err != nil {
		return "", err
	}
	cmd := exec.Command("git", "clone", "--depth=1", url, tmp)
	out, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tmp)
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return tmp, nil
}

func ScanCandidates(repoPath string) ([]string, error) {
	var candidates []string
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if info.Name() == ".git" {
			return filepath.SkipDir
		}
		if isValidTheme(path) {
			candidates = append(candidates, path)
		}
		return nil
	})
	return candidates, err
}

func isValidTheme(dir string) bool {
	for _, f := range []string{"config.jsonc", "style.css"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			return false
		}
	}
	return true
}

func Install(sourceDir, name string) error {
	dest := filepath.Join(theme.ThemesDir, name)
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("ya existe un tema llamado '%s'", name)
	}
	if err := os.MkdirAll(theme.ThemesDir, 0755); err != nil {
		return err
	}
	return copyDir(sourceDir, dest)
}

func Cleanup(tmpPath string) {
	if tmpPath != "" {
		os.RemoveAll(tmpPath)
	}
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
