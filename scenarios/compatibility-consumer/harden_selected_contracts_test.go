package compatibilityconsumer_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/faustbrian/go-international/v4/locale"
	jsonapi "github.com/faustbrian/go-jsonapi/v2"
	localized "github.com/faustbrian/go-localized/v5"
	localizedquery "github.com/faustbrian/go-localized/v5/adapters/query"
)

func TestHardenSelectedJSONAPIRejectsNullStrings(t *testing.T) {
	codec, err := jsonapi.NewCodec(jsonapi.CodecOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, decoder := range []struct {
		name   string
		decode func([]byte) (jsonapi.Document, error)
	}{
		{"core", jsonapi.Unmarshal},
		{"registered", codec.Unmarshal},
	} {
		t.Run(decoder.name, func(t *testing.T) {
			for _, test := range []struct{ name, payload, path string }{
				{"resource-id", `{"data":{"type":"orders","id":null}}`, "/data/id"},
				{"relationship-type", `{"data":{"type":"orders","id":"1","relationships":{"customer":{"data":{"type":null,"id":"2"}}}}}`, "/data/relationships/customer/data/type"},
				{"version", `{"jsonapi":{"version":null},"data":null}`, "/jsonapi/version"},
				{"link-rel", `{"data":null,"links":{"self":{"href":"/orders","rel":null}}}`, "/links/self/rel"},
				{"error-title", `{"errors":[{"title":null}]}`, "/errors/0/title"},
				{"source-pointer", `{"errors":[{"source":{"pointer":null}}]}`, "/errors/0/source/pointer"},
			} {
				t.Run(test.name, func(t *testing.T) {
					_, err := decoder.decode([]byte(test.payload))
					var failure *jsonapi.DecodeError
					if !errors.As(err, &failure) || failure.Code != "type" || failure.Path != test.path {
						t.Fatalf("null string error = %v; want type error at %s", err, test.path)
					}
				})
			}
		})
	}
	_, err = jsonapi.UnmarshalAtomic([]byte(`{"atomic:operations":[{"op":"remove","ref":{"type":"orders","id":null}}]}`))
	var failure *jsonapi.DecodeError
	if !errors.As(err, &failure) || failure.Code != "type" || failure.Path != "/atomic:operations/0/ref/id" {
		t.Fatalf("Atomic null identity error = %v; want type error at operation ref/id", err)
	}
}

func TestHardenSelectedJSONAPIPreservesEmptyIdentityAndLegalNull(t *testing.T) {
	for _, payload := range []string{
		`{"data":null}`,
		`{"data":{"type":"orders","id":"","attributes":{"note":null},"relationships":{"customer":{"data":null}}},"links":{"self":null}}`,
	} {
		document, err := jsonapi.Unmarshal([]byte(payload))
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := jsonapi.Marshal(document)
		if err != nil {
			t.Fatal(err)
		}
		var before, after any
		if err := json.Unmarshal([]byte(payload), &before); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(encoded, &after); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("valid empty/null presence changed: %s", encoded)
		}
	}
}

func TestHardenSelectedLocalizedRetainsCustomBudgetSubsets(t *testing.T) {
	for _, budget := range []string{"text", "count", "total"} {
		t.Run(budget, func(t *testing.T) {
			limits := localized.DefaultLimits()
			pairs := []localized.Pair{{Locale: "sv", Text: "remove"}, {Locale: "fi", Text: ""}}
			switch budget {
			case "text":
				limits.MaxTextBytes++
				pairs = append(pairs, localized.Pair{Locale: "en", Text: strings.Repeat("x", limits.MaxTextBytes)})
			case "count":
				limits.MaxLocales += 3
				for i := 0; i < limits.MaxLocales-2; i++ {
					pairs = append(pairs, localized.Pair{Locale: fmt.Sprintf("x-test%03d", i), Text: "present"})
				}
			case "total":
				limits.MaxTotalBytes = 10 << 20
				for i := 0; i < 9; i++ {
					pairs = append(pairs, localized.Pair{Locale: fmt.Sprintf("x-test%03d", i), Text: strings.Repeat("x", 1<<20)})
				}
			}
			text, err := localized.TextFromPairsWithOptions(localized.ConstructionOptions{Limits: limits}, pairs...)
			if err != nil {
				t.Fatal(err)
			}
			original := text.Pairs()
			var want []localized.Pair
			for _, pair := range original {
				if pair.Locale != "sv" {
					want = append(want, pair)
				}
			}
			swedish, err := locale.Parse("sv")
			if err != nil {
				t.Fatal(err)
			}
			for _, operation := range []struct {
				name  string
				apply func() localized.Text
			}{
				{"remove", func() localized.Text { return text.Remove(swedish) }},
				{"filter", func() localized.Text {
					return text.Filter(func(tag locale.Tag, _ string) bool { return tag.String() != "sv" })
				}},
			} {
				t.Run(operation.name, func(t *testing.T) {
					result := operation.apply()
					if !reflect.DeepEqual(result.Pairs(), want) {
						t.Fatalf("subset lost translations/order above default %s budget: retained %d, want %d", budget, len(result.Pairs()), len(want))
					}
					for _, pair := range want {
						tag, err := locale.Parse(pair.Locale)
						if err != nil {
							t.Fatal(err)
						}
						value, present := localizedquery.ExactValue(result, tag)
						if !present || value.String() != pair.Text {
							t.Fatalf("APIQuery exact value lost retained locale %s (including present-empty)", pair.Locale)
						}
					}
					if _, present := localizedquery.ExactValue(result, swedish); present {
						t.Fatal("APIQuery still observes removed locale")
					}
					if !reflect.DeepEqual(text.Pairs(), original) {
						t.Fatal("subset operation changed its immutable receiver")
					}
				})
			}
		})
	}
}
