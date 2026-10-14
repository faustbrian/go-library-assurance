// These finite HTTP scenarios qualify the selected public discovery and query
// seams. The application, not OpenRPC metadata, owns parameter validation.
package compatibilityconsumer_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	apiquery "github.com/faustbrian/go-api-query/v4"
	jsonschema "github.com/faustbrian/go-json-schema/v2"
	jsonrpc "github.com/faustbrian/go-jsonrpc"
	openrpc "github.com/faustbrian/go-openrpc/v2"
	"github.com/faustbrian/go-openrpc/v2/builder"
	"github.com/faustbrian/go-openrpc/v2/discovery"
	rpcadapter "github.com/faustbrian/go-openrpc/v2/jsonrpc"
	rpcschema "github.com/faustbrian/go-openrpc/v2/jsonschema"
	"github.com/faustbrian/go-openrpc/v2/jsonvalue"
	"github.com/faustbrian/go-wire/v3/jsonwire"
)

func TestHardenPublishedRPCDiscoverySchemaQueryHTTP(t *testing.T) {
	const policy = `{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","properties":{"status":{"enum":["paid","pending"]},"size":{"type":"integer","minimum":1,"maximum":2}},"required":["status","size"],"additionalProperties":false}`
	compiler, err := jsonschema.NewCompiler(jsonschema.WithDialect(jsonschema.Draft7))
	if err != nil {
		t.Fatal(err)
	}
	compiled, err := compiler.Compile(t.Context(), []byte(policy))
	if err != nil {
		t.Fatal(err)
	}
	schemaValue, err := rpcschema.Parse([]byte(policy), jsonvalue.DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	descriptor, err := openrpc.NewContentDescriptor(openrpc.ContentDescriptorInput{Name: "request", Schema: &schemaValue})
	if err != nil {
		t.Fatal(err)
	}
	method, err := openrpc.NewMethod(openrpc.MethodInput{Name: "orders", Params: []openrpc.ContentDescriptorOrReference{openrpc.ContentDescriptorValue(descriptor)}})
	if err != nil {
		t.Fatal(err)
	}
	version, err := openrpc.ParseVersion("1.4.1")
	if err != nil {
		t.Fatal(err)
	}
	info, err := openrpc.NewInfo(openrpc.InfoInput{Title: "Orders", Version: "1"})
	if err != nil {
		t.Fatal(err)
	}
	documentBuilder, err := builder.NewDocument(version, info)
	if err != nil {
		t.Fatal(err)
	}
	documentBuilder, err = documentBuilder.WithMethod(method)
	if err != nil {
		t.Fatal(err)
	}
	document, err := documentBuilder.Build()
	if err != nil {
		t.Fatal(err)
	}
	discoverer, err := discovery.NewService(discovery.Static(document), nil)
	if err != nil {
		t.Fatal(err)
	}
	querySchema, err := apiquery.NewSchema(apiquery.SchemaConfig{
		Resource: "orders", Revision: "orders-1",
		Fields:     []apiquery.FieldDefinition{{Name: "status", Type: apiquery.TypeString, Default: true}},
		Filters:    []apiquery.FilterDefinition{{Name: "status", Type: apiquery.TypeString, Operators: []apiquery.Operator{apiquery.OpEqual}}},
		Pagination: apiquery.PaginationDefinition{Offset: true, DefaultPageSize: 1},
		Bounds:     apiquery.Bounds{MaxPageSize: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	registry := jsonrpc.NewRegistry()
	var consumed atomic.Int32
	if err := registry.Register("orders", func(ctx context.Context, raw json.RawMessage) (any, error) {
		result, err := compiled.Validate(ctx, raw)
		if err != nil || !result.Valid {
			return nil, jsonrpc.InvalidParams()
		}
		var params struct {
			Status string `json:"status"`
			Size   int    `json:"size"`
		}
		if err := json.Unmarshal(raw, &params); err != nil {
			return nil, jsonrpc.InvalidParams()
		}
		plan, err := apiquery.Compile(ctx, querySchema, apiquery.Request{
			Filter: &apiquery.FilterExpr{Predicate: &apiquery.Predicate{Name: "status", Operator: apiquery.OpEqual, Values: []apiquery.Value{apiquery.StringValue(params.Status)}}},
			Page:   apiquery.PageRequest{Mode: apiquery.PageOffset, Size: params.Size},
		}, apiquery.CompileOptions{})
		if err != nil {
			return nil, jsonrpc.InvalidParams()
		}
		consumed.Add(1)
		return map[string]any{"status": plan.Filter().Predicate.Values[0].String(), "size": plan.Page().Size}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := rpcadapter.RegisterDiscovery[jsonrpc.Handler](registry, discoverer); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(jsonrpc.NewHTTPHandler(jsonrpc.NewDispatcher(registry)))
	t.Cleanup(server.Close)
	transport, err := jsonrpc.NewHTTPTransport(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := jsonrpc.NewClient(transport)
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	// The selected ordinary Client deliberately reserves rpc.* names. Preserve
	// that contract and use the public raw transport for this system extension.
	if _, err := jsonrpc.Call[json.RawMessage](ctx, client, "rpc.discover", []any{}); !errors.Is(err, jsonrpc.ErrInvalidMethodName) {
		t.Fatal("selected ordinary client lost its reserved-system-method boundary")
	}
	payload, err := transport.RoundTrip(ctx, []byte(`{"jsonrpc":"2.0","method":"rpc.discover","params":[],"id":7}`))
	if err != nil {
		t.Fatal(err)
	}
	var discoveryResponse jsonrpc.Response
	if err := jsonwire.Decode(payload, &discoveryResponse, jsonwire.DecodeOptions{MaxBytes: 4096}); err != nil {
		t.Fatal(err)
	}
	if err := discoveryResponse.Validate(); err != nil || discoveryResponse.Error != nil || !discoveryResponse.ID.Equal(jsonrpc.NumberID("7")) {
		t.Fatal("raw discovery HTTP response lost envelope/ID correlation")
	}
	// RawMessage requires the result to remain an embedded document object,
	// not an opaque struct or a JSON string containing another document.
	raw := discoveryResponse.Result
	var discovered struct {
		OpenRPC string `json:"openrpc"`
		Methods []struct {
			Name   string `json:"name"`
			Params []struct {
				Schema json.RawMessage `json:"schema"`
			} `json:"params"`
		} `json:"methods"`
	}
	if err := jsonwire.Decode(raw, &discovered, jsonwire.DecodeOptions{MaxBytes: 4096}); err != nil {
		t.Fatal(err)
	}
	if discovered.OpenRPC != "1.4.1" || len(discovered.Methods) != 1 || discovered.Methods[0].Name != "orders" || len(discovered.Methods[0].Params) != 1 {
		t.Fatal("HTTP discovery lost the registered public method/schema document")
	}
	var originalSchema, transportedSchema any
	if err := json.Unmarshal([]byte(policy), &originalSchema); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(discovered.Methods[0].Params[0].Schema, &transportedSchema); err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(originalSchema)
	b, _ := json.Marshal(transportedSchema)
	if string(a) != string(b) {
		t.Fatal("discovery changed the application's Draft 7 policy")
	}
	var response struct {
		Status string `json:"status"`
		Size   int    `json:"size"`
	}
	response, err = jsonrpc.Call[struct {
		Status string `json:"status"`
		Size   int    `json:"size"`
	}](ctx, client, "orders", map[string]any{"status": "paid", "size": 2})
	if err != nil || response.Status != "paid" || response.Size != 2 {
		t.Fatalf("public query HTTP result = %#v, %v", response, err)
	}
	for _, bad := range []map[string]any{{"status": "paid", "size": 3}, {"status": "unknown", "size": 1}} {
		_, err := jsonrpc.Call[json.RawMessage](ctx, client, "orders", bad)
		var protocolError *jsonrpc.Error
		if !errors.As(err, &protocolError) || protocolError.Code != jsonrpc.CodeInvalidParams {
			t.Fatalf("invalid schema input did not retain RPC error classification: %v", err)
		}
	}
	if consumed.Load() != 1 {
		t.Fatal("invalid schema inputs reached query consumption")
	}
	_, err = jsonrpc.Call[json.RawMessage](ctx, client, "missing", nil)
	var missing *jsonrpc.Error
	if !errors.As(err, &missing) || missing.Code != jsonrpc.CodeMethodNotFound {
		t.Fatalf("HTTP method-not-found = %v", err)
	}
}
