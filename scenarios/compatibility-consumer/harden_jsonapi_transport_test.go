// JSON:API owns document, media and cursor semantics, not a server framework.
// This finite application seam carries their actual outputs over loopback HTTP.
package compatibilityconsumer_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	jsonapi "github.com/faustbrian/go-jsonapi/v2"
	"github.com/faustbrian/go-wire/v3/jsonwire"
)

func TestHardenPublishedJSONAPICursorMediaErrorHTTP(t *testing.T) {
	negotiator, err := jsonapi.NewNegotiator(nil, []string{jsonapi.CursorPaginationProfileURI})
	if err != nil {
		t.Fatal(err)
	}
	pagination, err := jsonapi.NewCursorPagination(jsonapi.CursorPaginationConfig{DefaultSize: 2, MaxSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		write := func(status int, contentType string, document jsonapi.Document) {
			payload, err := jsonapi.Marshal(document)
			if err != nil {
				t.Error("fixture produced invalid JSON:API output")
				w.WriteHeader(500)
				return
			}
			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(status)
			if _, err := w.Write(payload); err != nil {
				t.Error("loopback response write failed")
			}
		}
		media, err := negotiator.NegotiateAccept(r.Header.Get("Accept"))
		if err == nil && r.Method == http.MethodPost {
			_, err = negotiator.CheckContentType(r.Header.Get("Content-Type"))
		}
		if err != nil {
			var failure *jsonapi.NegotiationError
			if !errors.As(err, &failure) {
				t.Error("negotiation lost typed classification")
				w.WriteHeader(500)
				return
			}
			write(failure.Status, jsonapi.MediaTypeJSONAPI, jsonapi.Document{Errors: []jsonapi.ErrorObject{{Status: strconv.Itoa(failure.Status), Code: failure.Code, Title: "Media negotiation failed"}}})
			return
		}
		query, err := jsonapi.ParseQuery(r.URL.Query())
		if err != nil {
			t.Error("fixture query parse failed")
			w.WriteHeader(500)
			return
		}
		request, err := pagination.ParseQuery(query)
		if err != nil {
			var failure *jsonapi.CursorPaginationError
			if !errors.As(err, &failure) {
				t.Error("cursor lost typed classification")
				w.WriteHeader(500)
				return
			}
			write(failure.Status, media.ContentType, jsonapi.Document{Errors: []jsonapi.ErrorObject{failure.ErrorObject("Invalid page", "Review the page size")}})
			return
		}
		if request.After != "one" || !request.AfterPresent || request.Size != 2 {
			t.Error("HTTP query lost the supplied cursor/page request")
			w.WriteHeader(500)
			return
		}
		itemMeta := jsonapi.CursorItemMeta("two")
		page := jsonapi.CursorPage{
			Request: request, Links: jsonapi.Links{"prev": jsonapi.URI("/orders?page%5Bbefore%5D=two"), "next": jsonapi.NullLink()},
			Items: []jsonapi.Meta{itemMeta}, HasPrevious: true,
		}
		if err := page.Validate(); err != nil {
			t.Errorf("fixture cursor page contract failed: %v", err)
			w.WriteHeader(500)
			return
		}
		if media.VaryAccept {
			w.Header().Set("Vary", "Accept")
		}
		write(200, media.ContentType, jsonapi.Document{
			JSONAPI: &jsonapi.JSONAPI{Version: "1.1", Profile: []string{jsonapi.CursorPaginationProfileURI}},
			Links:   page.Links, Data: jsonapi.ResourceCollection(jsonapi.ResourceObject{Type: "orders", ID: "two", Meta: itemMeta}),
		})
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	client.Timeout = 3 * time.Second
	profile := jsonapi.MediaTypeJSONAPI + `;profile="` + jsonapi.CursorPaginationProfileURI + `"`
	for _, test := range []struct {
		name, method, query, accept, contentType string
		status                                   int
	}{
		{"cursor-success", "GET", "page%5Bafter%5D=one&page%5Bsize%5D=2", profile, "", 200},
		{"cursor-limit", "GET", "page%5Bsize%5D=3", profile, "", 400},
		{"not-acceptable", "GET", "", "application/xml", "", 406},
		{"unsupported-content-type", "POST", "", profile, "application/json", 415},
	} {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequestWithContext(t.Context(), test.method, server.URL+"/orders?"+test.query, strings.NewReader("{}"))
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Accept", test.accept)
			if test.contentType != "" {
				request.Header.Set("Content-Type", test.contentType)
			}
			response, err := client.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != test.status || !strings.HasPrefix(response.Header.Get("Content-Type"), jsonapi.MediaTypeJSONAPI) {
				t.Fatalf("HTTP status/media = %d / %q", response.StatusCode, response.Header.Get("Content-Type"))
			}
			var raw json.RawMessage
			if err := jsonwire.DecodeReader(response.Body, &raw, jsonwire.DecodeOptions{MaxBytes: 4096, DisallowDuplicateNames: true}); err != nil {
				t.Fatal(err)
			}
			document, err := jsonapi.Unmarshal(raw)
			if err != nil {
				t.Fatal(err)
			}
			if test.status != 200 {
				if document.Data != nil || len(document.Errors) != 1 || document.Errors[0].Status != strconv.Itoa(test.status) {
					t.Fatal("transport mixed data/errors or lost protocol error status")
				}
				if test.status == 400 {
					failure := document.Errors[0]
					if failure.Code != "max-size-exceeded" || failure.Source == nil || failure.Source.Parameter != "page[size]" {
						t.Fatal("cursor error lost profile code/source")
					}
					encoded, err := json.Marshal(failure)
					if err != nil || !strings.Contains(string(encoded), jsonapi.CursorMaxSizeExceededTypeURI) || !strings.Contains(string(encoded), `"maxSize":2`) {
						t.Fatal("cursor error lost required type link or max-size metadata")
					}
				}
				return
			}
			if document.Data == nil || len(document.Errors) != 0 || !strings.Contains(response.Header.Get("Content-Type"), jsonapi.CursorPaginationProfileURI) {
				t.Fatal("success lost its cursor representation")
			}
			var transported struct {
				Data []struct {
					ID   string       `json:"id"`
					Meta jsonapi.Meta `json:"meta"`
				} `json:"data"`
				Links map[string]json.RawMessage `json:"links"`
			}
			if err := json.Unmarshal(raw, &transported); err != nil {
				t.Fatal(err)
			}
			if len(transported.Data) != 1 || transported.Data[0].ID != "two" || string(transported.Links["next"]) != "null" {
				t.Fatal("HTTP cursor response lost selected item or terminal next link")
			}
			cursor, present, err := jsonapi.ParseCursorItemMeta(transported.Data[0].Meta)
			if err != nil || !present || cursor != "two" {
				t.Fatal("Wire transport changed the opaque item cursor")
			}
		})
	}
}
