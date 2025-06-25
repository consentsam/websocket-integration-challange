package bugs

import (
	"os"
	"reflect"
	"testing"

	"github.com/consentsam/websocket-integration-challange/internal/config"
)

func TestBug06_Repro(t *testing.T) {
	os.Setenv("ENVIRONMENT", "local")
	// run from temporary directory to mimic installed binary location
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	if err := os.Chdir(temp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)

	cfg, err := config.LoadConfig("websocket-service")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	want := []string{"BTC_USDT", "ETH_USDT", "SOL_USDT"}
	if !reflect.DeepEqual(cfg.Delta.ProductIDs, want) {
		t.Fatalf("expected %v, got %v", want, cfg.Delta.ProductIDs)
	}
}
