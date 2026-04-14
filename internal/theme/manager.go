package theme

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	WaybarDir    = filepath.Join(mustHomeDir(), ".config", "waybar")
	ThemesDir    = filepath.Join(mustHomeDir(), ".config", "waybar", "themes")
	stateFile    = filepath.Join(mustHomeDir(), ".config", "waybar", ".waytui-active")
	managedFiles = []string{"config.jsonc", "style.css"}
)

func mustHomeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return h
}

func ListThemes() ([]string, error) {
	if err := os.MkdirAll(ThemesDir, 0755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(ThemesDir)
	if err != nil {
		return nil, err
	}
	var themes []string
	for _, e := range entries {
		if e.IsDir() && isValid(filepath.Join(ThemesDir, e.Name())) {
			themes = append(themes, e.Name())
		}
	}
	sort.Strings(themes)
	return themes, nil
}

func isValid(dir string) bool {
	for _, f := range managedFiles {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			return false
		}
	}
	return true
}

func GetActive() string {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func Apply(name string) error {
	themeDir := filepath.Join(ThemesDir, name)
	if !isValid(themeDir) {
		return fmt.Errorf("tema '%s' no válido (falta config.jsonc o style.css)", name)
	}
	if err := backupExisting(); err != nil {
		return fmt.Errorf("error haciendo backup: %w", err)
	}
	for _, f := range managedFiles {
		target := filepath.Join(WaybarDir, f)
		source := filepath.Join(themeDir, f)
		os.Remove(target)
		if err := os.Symlink(source, target); err != nil {
			return fmt.Errorf("error creando symlink %s: %w", f, err)
		}
	}
	if err := os.WriteFile(stateFile, []byte(name), 0644); err != nil {
		return err
	}
	return restartWaybar()
}

func Delete(name string) error {
	themeDir := filepath.Join(ThemesDir, name)
	if _, err := os.Stat(themeDir); err != nil {
		return fmt.Errorf("tema '%s' no existe", name)
	}
	if GetActive() == name {
		for _, f := range managedFiles {
			target := filepath.Join(WaybarDir, f)
			if link, err := os.Readlink(target); err == nil {
				if strings.HasPrefix(link, themeDir) {
					os.Remove(target)
				}
			}
		}
		os.WriteFile(stateFile, []byte(""), 0644)
	}
	return os.RemoveAll(themeDir)
}

func GetFiles(name string) map[string]string {
	result := make(map[string]string)
	themeDir := filepath.Join(ThemesDir, name)
	for _, f := range managedFiles {
		data, err := os.ReadFile(filepath.Join(themeDir, f))
		if err == nil {
			result[f] = string(data)
		}
	}
	return result
}

func backupExisting() error {
	hasReal := false
	for _, f := range managedFiles {
		info, err := os.Lstat(filepath.Join(WaybarDir, f))
		if err == nil && info.Mode()&os.ModeSymlink == 0 {
			hasReal = true
			break
		}
	}
	if !hasReal {
		return nil
	}
	date := time.Now().Format("2006-01-02")
	backupDir := filepath.Join(ThemesDir, "backup-"+date)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}
	for _, f := range managedFiles {
		src := filepath.Join(WaybarDir, f)
		info, err := os.Lstat(src)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		if err := copyFile(src, filepath.Join(backupDir, f)); err != nil {
			return err
		}
	}
	return nil
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

func restartWaybar() error {
	if _, err := exec.LookPath("omarchy-restart-waybar"); err == nil {
		return exec.Command("omarchy-restart-waybar").Run()
	}
	exec.Command("pkill", "-x", "waybar").Run()
	return exec.Command("waybar").Start()
}
