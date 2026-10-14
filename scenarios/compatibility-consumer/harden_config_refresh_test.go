// Refresh publication belongs to the caller: Config returns a validated
// candidate and never replaces application state or installs a watcher.
package compatibilityconsumer_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	config "github.com/faustbrian/go-config/v2"
	configfilesystem "github.com/faustbrian/go-config/v2/filesystem"
	configvalidation "github.com/faustbrian/go-config/v2/validation"
	filesystem "github.com/faustbrian/go-filesystem/v2"
	"github.com/faustbrian/go-filesystem/v2/local"
	jsonschema "github.com/faustbrian/go-json-schema/v2"
	"github.com/faustbrian/go-wire/v3/jsonwire"
)

// hardenRefreshSettings is the application's consumed configuration, not a
// new library adapter. Secret remains explicit in both config and JSON views.
type hardenRefreshSettings struct {
	// Status is the admitted order state consumed by the application.
	Status string `config:"status" json:"status"`
	// Size is the schema-bounded page size consumed by the application.
	Size int `config:"size" json:"size"`
	// Token is explicitly revealed only at the trusted consumption boundary.
	Token config.Secret `config:"token,secret" json:"token"`
}

func TestHardenPublishedConfigWireFilesystemRefresh(t *testing.T) {
	store, err := local.Open(t.Context(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	path := filesystem.MustParsePath("settings.json")
	// The sibling Filesystem owns storage. Config's filesystem subpackage
	// consumes its public io/fs bridge and reopens the source for each load.
	source, err := configfilesystem.FromFS(filesystem.NewIOFS(store, store, store), path.String(), configfilesystem.Options{Name: "settings", Sensitive: true})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := config.NewPlan(source)
	if err != nil {
		t.Fatal(err)
	}
	compiler, err := jsonschema.NewCompiler(jsonschema.WithDialect(jsonschema.Draft7))
	if err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(t.Context(), []byte(`{"type":"object","properties":{"status":{"enum":["paid","pending"]},"size":{"type":"integer","minimum":1,"maximum":2},"token":{"type":"string"}},"required":["status","size","token"],"additionalProperties":false}`))
	if err != nil {
		t.Fatal(err)
	}
	rejected := errors.New("private-validator-fixture")
	validator := func(ctx context.Context, candidate hardenRefreshSettings) error {
		encoded, err := json.Marshal(candidate)
		if err != nil {
			return err
		}
		result, err := schema.Validate(ctx, encoded)
		if err != nil {
			return err
		}
		if !result.Valid {
			return configvalidation.At("orders", rejected)
		}
		return nil
	}
	var published *config.Snapshot[hardenRefreshSettings]
	refresh := func() error {
		candidate, err := config.LoadWithValidators(t.Context(), plan, validator)
		if err != nil {
			return err
		}
		published = candidate
		return nil
	}
	write := func(status string, size int, token string) {
		t.Helper()
		payload, err := jsonwire.Encode(map[string]any{"status": status, "size": size, "token": token}, jsonwire.EncodeOptions{MaxBytes: 1024})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.Write(t.Context(), path, bytes.NewReader(payload), filesystem.WriteOptions{}); err != nil {
			t.Fatal(err)
		}
	}
	const firstSecret = "first-private-fixture"
	write("paid", 2, firstSecret)
	if err := refresh(); err != nil {
		t.Fatal(err)
	}
	first := published
	if value := published.Value(); value.Status != "paid" || value.Size != 2 || value.Token.Reveal() != firstSecret {
		t.Fatal("Wire/filesystem/config publication lost the admitted values")
	}
	origin, ok := published.Origin("token")
	if !ok || !origin.Sensitive || origin.Source != "settings" {
		t.Fatal("sensitive provenance was lost at the config filesystem seam")
	}
	const rejectedSecret = "rejected-private-fixture"
	write("unknown", 3, rejectedSecret)
	err = refresh()
	if err == nil || !errors.Is(err, rejected) || published != first {
		t.Fatal("invalid candidate replaced the last good snapshot or lost protected error identity")
	}
	failure := err
	encoded, err := json.Marshal(published.Value())
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := fmt.Sprintf("%v %+v %#v %s %v %+v %#v", published.Value(), published.Value(), published.Value(), encoded, failure, failure, failure)
	for _, secret := range []string{firstSecret, rejectedSecret, rejected.Error()} {
		if strings.Contains(diagnostics, secret) {
			t.Fatal("default configuration diagnostics disclosed a fixture secret")
		}
	}
	if !strings.Contains(diagnostics, config.Redacted) {
		t.Fatal("default diagnostic encoding lost explicit secret redaction")
	}
	// A later valid reload publishes anew; rejection must not poison the plan.
	write("pending", 1, "later-private-fixture")
	if err := refresh(); err != nil {
		t.Fatal(err)
	}
	if published == first || published.Value().Status != "pending" || first.Value().Status != "paid" || first.Value().Token.Reveal() != firstSecret {
		t.Fatal("valid refresh or retained snapshot identity was lost")
	}
}
