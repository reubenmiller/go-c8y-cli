// Package curly provides functionality to convert an http.Request into its equivalent curl command string
package curly

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// DummyFile represents a file that is created in memory when converting
// a multipart/form-data request to a cURL command, especially when a form part
// is identified as a file but its content is provided directly rather than as a filename.
type DummyFile struct {
	Name     string
	Contents []byte
}

// NewDummyFileCollection creates and initializes a new DummyFileCollection.
func NewDummyFileCollection() *DummyFileCollection {
	return &DummyFileCollection{
		Files: make([]DummyFile, 0),
	}
}

// DummyFileCollection holds a collection of DummyFile instances.
type DummyFileCollection struct {
	Files []DummyFile
}

// Append adds a new DummyFile to the collection.
// If the provided name is empty, a default name is generated (e.g., "input1.txt").
// It returns the name of the appended file (either the provided name or the generated one).
func (f *DummyFileCollection) Append(name string, contents []byte) string {
	if name == "" {
		name = fmt.Sprintf("%s%d.txt", "input", len(f.Files)+1)
	}
	f.Files = append(f.Files, DummyFile{
		Name:     name,
		Contents: contents,
	})
	return name
}

// ToCurl converts an http.Request into a string representing the equivalent curl command.
// It handles various HTTP methods (GET, POST, PUT, DELETE) and request body types,
// including JSON, form-urlencoded, and multipart/form-data.
func ToCurl(req *http.Request) (string, *DummyFileCollection, error) {
	var curlBuilder strings.Builder
	curlBuilder.WriteString("curl")

	// Add HTTP method
	curlBuilder.WriteString(fmt.Sprintf(" -X %s", req.Method))

	// Add URL
	curlBuilder.WriteString(fmt.Sprintf(" '%s'", req.URL.String()))

	isMultiPartFormData := true

	// Add main request headers
	for _, name := range getSortedHeaders(req.Header) {
		for _, value := range req.Header[name] {
			// Skip Content-Length as curl calculates it automatically when using --data or --form
			if strings.EqualFold(name, "Content-Length") {
				continue
			}
			// Skip multipart/form-data
			if strings.HasPrefix(value, "multipart/form-data") {
				isMultiPartFormData = true
				continue
			}
			// Escape single quotes in header values for shell compatibility
			curlBuilder.WriteString(fmt.Sprintf(" -H '%s: %s'", name, escapeSingleQuotes(value)))
		}
	}

	_ = isMultiPartFormData
	dummyFiles := NewDummyFileCollection()

	// Handle request body
	if req.Body != nil && req.Body != http.NoBody {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return "", nil, fmt.Errorf("failed to read request body: %w", err)
		}
		// Restore the body for the original http.Request if it's intended for further processing.
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		contentType := req.Header.Get("Content-Type")

		if strings.HasPrefix(contentType, "application/json") ||
			strings.HasPrefix(contentType, "application/xml") ||
			strings.HasPrefix(contentType, "text/") {
			curlBuilder.WriteString(fmt.Sprintf(" --data-raw '%s'", escapeSingleQuotes(string(bodyBytes))))
		} else if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
			form, err := url.ParseQuery(string(bodyBytes))
			if err != nil {
				return "", nil, fmt.Errorf("failed to parse form-urlencoded body: %w", err)
			}
			for key, values := range form {
				for _, value := range values {
					formData := url.Values{}
					formData.Add(key, value)
					// encode values, and then replace + with %20
					curlBuilder.WriteString(fmt.Sprintf(" --data '%s'", strings.Replace(formData.Encode(), "+", "%20", -1)))
				}
			}
		} else if strings.HasPrefix(contentType, "multipart/form-data") {
			// Handling multipart/form-data: each part translates to a -F argument.
			// File parts use @filename, regular fields use value.
			// Part-specific Content-Type headers are added as ';type=mimetype'.

			boundary, err := getBoundary(contentType)
			if err != nil {
				return "", nil, fmt.Errorf("failed to get boundary for multipart/form-data: %w", err)
			}

			reader := multipart.NewReader(bytes.NewReader(bodyBytes), boundary)

			for {
				part, err := reader.NextPart()
				if err == io.EOF {
					break // No more parts
				}
				if err != nil {
					return "", nil, fmt.Errorf("error reading multipart part: %w", err)
				}

				partBytes, err := io.ReadAll(part)
				if err != nil {
					return "", nil, fmt.Errorf("error reading multipart part data: %w", err)
				}

				formValue := ""
				// Determine if it's a file field or a regular form field based on FileName()
				if part.FileName() != "" {
					// This is a file field: use 'name=@filename' format for curl -F
					formValue = fmt.Sprintf("%s=@%s", escapeSingleQuotes(part.FormName()), escapeSingleQuotes(part.FileName()))
					dummyFiles.Append(part.FileName(), partBytes)
				} else {
					// Check if it is plain text or a file reference
					if hasFileParameter(part) {
						// the file is too large, create a dummy file instead and reference it
						dummyFileName := dummyFiles.Append("", partBytes)
						formValue = fmt.Sprintf("%s=@%s", escapeSingleQuotes(part.FormName()), escapeSingleQuotes(dummyFileName))
					} else {
						// This is a regular text field: use 'name=value' format for curl -F
						formValue = fmt.Sprintf("%s=%s", escapeSingleQuotes(part.FormName()), escapeSingleQuotes(string(partBytes)))
					}
				}

				// If the part has its own Content-Type header, append it using ';type=mimetype'
				if part.Header.Get("Content-Type") != "" {
					formValue += fmt.Sprintf(";type=%s", escapeSingleQuotes(part.Header.Get("Content-Type")))
				}

				// Append the fully constructed -F argument
				curlBuilder.WriteString(fmt.Sprintf(" -F '%s'", formValue))
			}
		} else if len(bodyBytes) > 0 {
			curlBuilder.WriteString(fmt.Sprintf(" --data-binary '%s'", escapeSingleQuotes(string(bodyBytes))))
		}
	}

	return curlBuilder.String(), dummyFiles, nil
}

func getSortedHeaders(header http.Header) []string {
	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// escapeSingleQuotes escapes single quotes in a string for shell compatibility.
// Example: 'foo'bar' becomes 'foo'\"bar'
func escapeSingleQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "'\\''")
}

// getBoundary extracts the boundary string from a Content-Type header
// for multipart/form-data requests.
func getBoundary(contentType string) (string, error) {
	// Changed to mime.ParseMediaType
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", fmt.Errorf("failed to parse media type: %w", err)
	}
	if boundary, ok := params["boundary"]; ok {
		return boundary, nil
	}
	return "", fmt.Errorf("no boundary found in Content-Type header")
}

// hasFileParameter checks if a multipart.Part is likely a file upload
// by inspecting its Content-Disposition header for a "name" parameter equal to "file".
func hasFileParameter(part *multipart.Part) bool {
	if v := part.Header.Get("Content-Disposition"); v != "" {
		if _, params, err := mime.ParseMediaType(v); err == nil {
			if objectName, ok := params["name"]; ok && objectName == "file" {
				return true
			}
		}
	}
	return false
}
