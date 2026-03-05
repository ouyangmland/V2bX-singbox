package conf

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConf_LoadFromPath(t *testing.T) {
	c := New()
	t.Log(c.LoadFromPath("../example/config.json"), c.NodeConfig)
}

func TestConf_Watch(t *testing.T) {
	c := New()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	reloaded := make(chan struct{}, 1)
	err := c.Watch(configPath, "", "", func() {
		select {
		case reloaded <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte(`{"changed":true}`), 0644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-reloaded:
	case <-time.After(8 * time.Second):
		t.Fatal("watch callback timeout")
	}
}
