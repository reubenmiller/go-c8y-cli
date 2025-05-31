package curly

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/textproto" // Required for multipart custom headers in tests
	"net/url"
	"strings"
	"testing"
)

// helperRequest creates a simple http.Request for testing purposes.
// It simplifies the creation of requests with various methods, URLs, bodies, and headers.
func helperRequest(method, urlStr string, body io.Reader, headers map[string]string) *http.Request {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		// Panic in tests if request creation fails, as it indicates a test setup issue.
		panic(err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req
}

// TestToCurl_GETRequest verifies the conversion of GET requests.
func TestToCurl_GETRequest(t *testing.T) {
	tests := []struct {
		name                 string
		req                  *http.Request
		expectedBaseCommand  string   // Expected part before headers
		expectedHeaderParts  []string // Expected header flags
		expectedNoHeaderPart bool     // True if no headers are expected
	}{
		{
			name:                 "Simple GET request",
			req:                  helperRequest(http.MethodGet, "http://example.com/api/data", nil, nil),
			expectedBaseCommand:  "curl -X GET 'http://example.com/api/data'",
			expectedNoHeaderPart: true,
		},
		{
			name:                 "GET request with query parameters",
			req:                  helperRequest(http.MethodGet, "http://example.com/api/data?param1=value1&param2=value%20two", nil, nil),
			expectedBaseCommand:  "curl -X GET 'http://example.com/api/data?param1=value1&param2=value%20two'",
			expectedNoHeaderPart: true,
		},
		{
			name: "GET request with headers",
			req: helperRequest(http.MethodGet, "http://example.com/api/user", nil, map[string]string{
				"Authorization": "Bearer token123",
				"Accept":        "application/json",
			}),
			expectedBaseCommand: "curl -X GET 'http://example.com/api/user'",
			expectedHeaderParts: []string{
				"-H 'Authorization: Bearer token123'",
				"-H 'Accept: application/json'",
			},
		},
		{
			name: "GET request with query params and headers",
			req: helperRequest(http.MethodGet, "http://example.com/api/search?q=test&page=1", nil, map[string]string{
				"User-Agent": "Golang-Test-Client",
			}),
			expectedBaseCommand: "curl -X GET 'http://example.com/api/search?q=test&page=1'",
			expectedHeaderParts: []string{
				"-H 'User-Agent: Golang-Test-Client'",
			},
		},
		{
			name: "GET request with special chars in URL",
			req: func() *http.Request {
				// The input URL string to helperRequest
				urlInput := "http://example.com/path with/spaces?key=value's"
				return helperRequest(http.MethodGet, urlInput, nil, nil)
			}(),
			expectedBaseCommand: func() string {
				// Dynamically construct the expected URL string based on Go's net/url behavior
				u, _ := url.Parse("http://example.com/path with/spaces?key=value's")
				return fmt.Sprintf("curl -X GET '%s'", u.String())
			}(),
			expectedNoHeaderPart: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := ToCurl(tt.req)
			if err != nil {
				t.Fatalf("ToCurl returned an unexpected error: %v", err)
			}

			// Check base command
			if !strings.HasPrefix(got, tt.expectedBaseCommand) {
				t.Errorf("ToCurl() got command prefix = %q, want %q", got, tt.expectedBaseCommand)
			}

			// Check headers independently of order
			for _, headerPart := range tt.expectedHeaderParts {
				if !strings.Contains(got, headerPart) {
					t.Errorf("ToCurl() missing expected header part: %q in %q", headerPart, got)
				}
			}

			// Verify no unexpected headers if expectedNoHeaderPart is true
			if tt.expectedNoHeaderPart && strings.Contains(got, "-H '") {
				// This is a simplified check. A more rigorous one would parse all headers.
				// For most GET tests without explicit headers, this is sufficient.
				if len(strings.Split(got, " -H '")) > 1 { // Check if any -H flags are present
					t.Errorf("ToCurl() got unexpected headers: %q", got)
				}
			}
		})
	}
}

// TestToCurl_POSTRequest verifies the conversion of POST requests with various body types.
func TestToCurl_POSTRequest(t *testing.T) {
	tests := []struct {
		name                 string
		req                  *http.Request
		expectedBaseCommand  string   // Expected part before headers
		expectedHeaderParts  []string // Expected header flags
		expectedDataParts    []string // Expected --data or --data-raw flags
		expectedNoHeaderPart bool
	}{
		{
			name: "POST request with JSON body",
			req: helperRequest(http.MethodPost, "http://example.com/api/create",
				bytes.NewBufferString(`{"name":"test", "value":123}`),
				map[string]string{"Content-Type": "application/json"}),
			expectedBaseCommand: "curl -X POST 'http://example.com/api/create'",
			expectedHeaderParts: []string{
				"-H 'Content-Type: application/json'",
			},
			expectedDataParts: []string{
				`--data-raw '{"name":"test", "value":123}'`,
			},
		},
		{
			name: "POST request with JSON body and special characters",
			req: helperRequest(http.MethodPost, "http://example.com/api/create",
				bytes.NewBufferString(`{"message":"It's a test!"}`),
				map[string]string{"Content-Type": "application/json"}),
			expectedBaseCommand: "curl -X POST 'http://example.com/api/create'",
			expectedHeaderParts: []string{
				"-H 'Content-Type: application/json'",
			},
			expectedDataParts: []string{
				`--data-raw '{"message":"It'\''s a test!"}'`,
			},
		},
		{
			name: "POST request with form-urlencoded body",
			req: func() *http.Request {
				form := url.Values{}
				form.Add("field1", "value one")
				form.Add("field2", "value&two") // Contains special char
				return helperRequest(http.MethodPost, "http://example.com/api/form",
					strings.NewReader(form.Encode()),
					map[string]string{"Content-Type": "application/x-www-form-urlencoded"})
			}(),
			expectedBaseCommand: "curl -X POST 'http://example.com/api/form'",
			expectedHeaderParts: []string{
				"-H 'Content-Type: application/x-www-form-urlencoded'",
			},
			expectedDataParts: []string{
				`--data 'field1=value%20one'`,
				`--data 'field2=value%26two'`,
			},
		},
		{
			name:                 "POST request with empty body (http.NoBody)",
			req:                  helperRequest(http.MethodPost, "http://example.com/api/empty", http.NoBody, nil),
			expectedBaseCommand:  "curl -X POST 'http://example.com/api/empty'",
			expectedNoHeaderPart: true, // No Content-Type header if no body
		},
		{
			name:                 "POST request with empty buffer body",
			req:                  helperRequest(http.MethodPost, "http://example.com/api/empty", bytes.NewBuffer(nil), nil),
			expectedBaseCommand:  "curl -X POST 'http://example.com/api/empty'",
			expectedNoHeaderPart: true, // No Content-Type header if no body
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := ToCurl(tt.req)
			if err != nil {
				t.Fatalf("ToCurl returned an unexpected error: %v", err)
			}

			if !strings.HasPrefix(got, tt.expectedBaseCommand) {
				t.Errorf("ToCurl() got command prefix = %q, want %q", got, tt.expectedBaseCommand)
			}

			for _, headerPart := range tt.expectedHeaderParts {
				if !strings.Contains(got, headerPart) {
					t.Errorf("ToCurl() missing expected header part: %q in %q", headerPart, got)
				}
			}
			for _, dataPart := range tt.expectedDataParts {
				if !strings.Contains(got, dataPart) {
					t.Errorf("ToCurl() missing expected data part: %q in %q", dataPart, got)
				}
			}

			// Verify no unexpected headers if expectedNoHeaderPart is true
			if tt.expectedNoHeaderPart && strings.Contains(got, "-H '") {
				t.Errorf("ToCurl() got unexpected headers: %q", got)
			}
			// Verify no unexpected data if no data expected
			if len(tt.expectedDataParts) == 0 && (strings.Contains(got, "--data ") || strings.Contains(got, "--data-raw ") || strings.Contains(got, "--data-binary ")) {
				t.Errorf("ToCurl() got unexpected data flags: %q", got)
			}
		})
	}
}

// TestToCurl_PUTRequest verifies the conversion of PUT requests.
func TestToCurl_PUTRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      *http.Request
		expected string
	}{
		{
			name: "PUT request with XML body",
			req: helperRequest(http.MethodPut, "http://example.com/api/resource/1",
				bytes.NewBufferString(`<data><status>active</status></data>`),
				map[string]string{"Content-Type": "application/xml"}),
			expected: `curl -X PUT 'http://example.com/api/resource/1' -H 'Content-Type: application/xml' --data-raw '<data><status>active</status></data>'`,
		},
		{
			name: "PUT request with plain text body",
			req: helperRequest(http.MethodPut, "http://example.com/api/update",
				bytes.NewBufferString("new status: completed"),
				map[string]string{"Content-Type": "text/plain"}),
			expected: `curl -X PUT 'http://example.com/api/update' -H 'Content-Type: text/plain' --data-raw 'new status: completed'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := ToCurl(tt.req)
			if err != nil {
				t.Fatalf("ToCurl returned an unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("ToCurl() got = %q\nwant %q", got, tt.expected)
			}
		})
	}
}

// TestToCurl_DELETERequest verifies the conversion of DELETE requests.
func TestToCurl_DELETERequest(t *testing.T) {
	tests := []struct {
		name     string
		req      *http.Request
		expected string
	}{
		{
			name:     "DELETE request with no body",
			req:      helperRequest(http.MethodDelete, "http://example.com/api/items/123", nil, nil),
			expected: "curl -X DELETE 'http://example.com/api/items/123'",
		},
		{
			name: "DELETE request with headers",
			req: helperRequest(http.MethodDelete, "http://example.com/api/items/456", nil, map[string]string{
				"X-Confirm": "true",
			}),
			expected: "curl -X DELETE 'http://example.com/api/items/456' -H 'X-Confirm: true'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := ToCurl(tt.req)
			if err != nil {
				t.Fatalf("ToCurl returned an unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("ToCurl() got = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestToCurl_MultipartFormData verifies the conversion of multipart/form-data requests.
func TestToCurl_MultipartFormData(t *testing.T) {
	// Create a multipart form with text fields and a simulated file
	var b bytes.Buffer

	w := newCustomMultiPartWriter(&b)

	// Add a simple text field
	w.WriteField("username", "testuser")

	// Add another text field with special characters that need escaping
	w.WriteField("description", "This is a test description with 'quotes' and spaces.")

	// Add a simulated file field.
	// In a real scenario, this would come from an actual file.
	// Here, we simulate it with dummy content.
	fileFieldName := "profile_image"
	fileName := "avatar.jpg"
	fileContentType := "image/jpeg"
	fileContent := "dummy image content for testing"
	fw, err := w.CreateFormFileWithContentType(fileFieldName, fileName, fileContentType)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = io.Copy(fw, strings.NewReader(fileContent)) // Direct use of string
	if err != nil {
		t.Fatalf("Failed to write file content: %v", err)
	}

	// Add a part with custom headers (e.g., a JSON blob as a part)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="data_json"`)
	header.Set("Content-Type", "application/json")
	jsonPartWriter, err := w.CreatePart(header)
	if err != nil {
		t.Fatalf("Failed to create json part: %v", err)
	}
	_, err = jsonPartWriter.Write([]byte(`{"key":"value"}`))
	if err != nil {
		t.Fatalf("Failed to write json part content: %v", err)
	}

	w.Close() // IMPORTANT: Close the writer to finalize the multipart body and write the closing boundary.

	req := helperRequest(http.MethodPost, "http://example.com/api/upload",
		&b, // Use the buffer containing the multipart body
		map[string]string{"Content-Type": w.FormDataContentType()}) // Set the correct Content-Type with boundary

	got, _, err := ToCurl(req)
	if err != nil {
		t.Fatalf("ToCurl returned an unexpected error: %v", err)
	}

	// For multipart, the order of -F flags might not be strictly guaranteed due to map iteration
	// within the multipart reader. Therefore, we check for the presence of expected substrings
	// rather than an exact match for the entire string.
	if !strings.HasPrefix(got, "curl -X POST 'http://example.com/api/upload'") {
		t.Errorf("Curl command does not start with expected prefix:\nGot: %q\nWant prefix: %q", got, "curl -X POST 'http://example.com/api/upload'")
	}
	// TODO: Check this
	// if !strings.Contains(got, "-H 'Content-Type: "+w.FormDataContentType()+"'") {
	// 	t.Errorf("Curl command missing Content-Type header:\nGot: %q\nWant header part: %q", got, "-H 'Content-Type: "+w.FormDataContentType()+"'")
	// }

	// Define all expected form parts. Note the escaping for single quotes.
	expectedFormParts := []string{
		"-F 'username=testuser'",
		// "-F 'description=This is a test description with '\\''quotes'\\'' and spaces.'",
		"-F 'profile_image=@avatar.jpg;type=image/jpeg'", // File part placeholder
		"-F 'data_json={\"key\":\"value\"};type=application/json'",
	}

	for _, part := range expectedFormParts {
		if !strings.Contains(got, part) {
			t.Errorf("Curl command missing expected form part:\nGot: %q\nWant part: %q", got, part)
		}
	}
	t.Logf("Generated multipart curl: %q", got) // Log the generated command for debugging
}

// TestToCurl_BinaryBody verifies the conversion of requests with generic binary bodies.
func TestToCurl_BinaryBody(t *testing.T) {
	binaryData := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02, 0x03, 0x04} // Example binary data
	req := helperRequest(http.MethodPost, "http://example.com/api/binary",
		bytes.NewReader(binaryData),
		map[string]string{"Content-Type": "application/octet-stream"})

	// When using --data-binary, curl treats the input as raw bytes.
	// Go's string(binaryData) might result in unprintable characters, but curl handles it.
	expectedPrefix := "curl -X POST 'http://example.com/api/binary' -H 'Content-Type: application/octet-stream' --data-binary '"
	expectedSuffix := "'"

	got, _, err := ToCurl(req)
	if err != nil {
		t.Fatalf("ToCurl returned an unexpected error: %v", err)
	}

	if !strings.HasPrefix(got, expectedPrefix) {
		t.Errorf("Curl command does not start with expected binary prefix:\nGot: %q\nWant prefix: %q", got, expectedPrefix)
	}
	if !strings.HasSuffix(got, expectedSuffix) {
		t.Errorf("Curl command does not end correctly:\nGot: %q\nWant suffix: %q", got, expectedSuffix)
	}
	// A more thorough check for binary data might involve inspecting the raw bytes
	// but for typical curl usage, this string representation is what's expected.
}

func newMockBody(b string) *mockBody {
	return &mockBody{
		reader: bytes.NewBuffer([]byte(b)),
	}
}

type mockBody struct {
	reader io.Reader
	closed bool
}

func (mb *mockBody) Read(p []byte) (n int, err error) {
	if mb.closed {
		return 0, fmt.Errorf("already closed")
	}
	return mb.reader.Read(p)
}

func (mb *mockBody) Close() error {
	if mb.closed {
		return fmt.Errorf("already closed")
	}
	mb.closed = true
	return nil
}

// TestToCurl_ErrorHandling verifies that ToCurl handles errors gracefully,
// such as when the request body has already been read/consumed.
func TestToCurl_ErrorHandling(t *testing.T) {
	// Create a request with a body
	requestBody := newMockBody("data")
	req := helperRequest(http.MethodPost, "http://example.com/api/fail", requestBody, nil)

	// Simulate the body being consumed before ToCurl is called
	_, _ = io.ReadAll(req.Body)
	req.Body.Close() // Close the body after reading

	_, _, err := ToCurl(req)
	if err == nil {
		t.Errorf("ToCurl did not return an error for a consumed body")
	}
	if !strings.Contains(err.Error(), "failed to read request body") {
		t.Errorf("Unexpected error message for consumed body: %v", err)
	}

	// Test error handling for invalid multipart boundary
	invalidMultipartReq := helperRequest(http.MethodPost, "http://example.com/invalid-multipart",
		bytes.NewBufferString("some content"),
		map[string]string{"Content-Type": "multipart/form-data; noboundary"})
	_, _, err = ToCurl(invalidMultipartReq)
	if err == nil {
		t.Errorf("ToCurl did not return an error for invalid multipart Content-Type")
	}
	if !strings.Contains(err.Error(), "failed to get boundary for multipart/form-data") {
		t.Errorf("Unexpected error message for invalid multipart Content-Type: %v", err)
	}
}

// TestEscapeSingleQuotes verifies the `escapeSingleQuotes` helper function.
func TestEscapeSingleQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No quotes", "hello world", "hello world"},
		{"Single quote", "it's a test", "it'\\''s a test"},
		{"Quotes at start and end", "'quoted'", "'\\''quoted'\\''"},
		{"Multiple quotes", "one'two'three", "one'\\''two'\\''three"},
		{"Empty string", "", ""},
		{"String with backslashes", `path\to\file`, `path\to\file`}, // Backslashes are not escaped by this function
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeSingleQuotes(tt.input)
			if got != tt.expected {
				t.Errorf("escapeSingleQuotes(%q) got = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestGetBoundary verifies the `getBoundary` helper function.
func TestGetBoundary(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expected    string
		expectError bool
	}{
		{"Valid boundary", "multipart/form-data; boundary=----WebKitFormBoundary", "----WebKitFormBoundary", false},
		{"Valid boundary with charset", "multipart/form-data; charset=UTF-8; boundary=\"--boundary123\"", "--boundary123", false},
		{"No boundary parameter", "multipart/form-data", "", true},
		{"Invalid media type string", "invalid-type", "", true},
		{"Empty Content-Type string", "", "", true},
		{"Boundary with spaces (should be quoted)", `multipart/form-data; boundary="a b c"`, "a b c", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getBoundary(tt.contentType)
			if (err != nil) != tt.expectError {
				t.Errorf("getBoundary() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if got != tt.expected {
				t.Errorf("getBoundary() got = %q, want %q", got, tt.expected)
			}
		})
	}
}
