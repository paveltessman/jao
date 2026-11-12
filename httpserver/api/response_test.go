package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	status := 211
	payload := Response{
		Success: true,
		Data: struct {
			Opa string `json:"opa"`
		}{Opa: "hoba"},
	}

	writeJSON(w, status, payload)

	resp := w.Result()

	if got, want := resp.Header.Get("Content-Type"), "application/json"; got != want {
		t.Errorf("Content-Type header: got %q, want %q", got, want)
	}

	if resp.StatusCode != status {
		t.Errorf("statusCode: got=%d, want=%d", resp.StatusCode, status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response: %v", err)
	}

	expectedBody := `{"success":true,"data":{"opa":"hoba"}}` + "\n"
	if string(body) != expectedBody {
		t.Errorf("repsonse body not equal: got=%q, want=%q", string(body), expectedBody)
	}
}

func TestWriteJSONEncodeFailed(t *testing.T) {
	w := httptest.NewRecorder()

	payload := Response{
		Success: true,
		Data:    func() {},
	}

	writeJSON(w, http.StatusOK, payload)

	resp := w.Result()

	if got, want := resp.Header.Get("Content-Type"), "text/plain; charset=utf-8"; got != want {
		t.Errorf("Content-Type header: got %q, want %q", got, want)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("statusCode: got=%d, want=%d", resp.StatusCode, http.StatusInternalServerError)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response: %v", err)
	}

	expectedBody := `{"success": false, "error": "failed to encode response"}` + "\n"
	if string(body) != expectedBody {
		t.Errorf("repsonse body not equal: got=%q, want=%q", string(body), expectedBody)
	}
}
