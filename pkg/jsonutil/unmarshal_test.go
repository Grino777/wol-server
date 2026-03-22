package jsonutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

type dataModel struct {
	Keys                []dataModelKey `json:"temporaryExposureKeys"`
	Regions             []string       `json:"regions"`
	AppPackageName      string         `json:"appPackageName"`
	VerificationPayload string         `json:"verificationPayload"`
}

type dataModelKey struct {
	Key              string `json:"key"`
	IntervalNumber   int32  `json:"rollingStartNumber"`
	IntervalCount    int32  `json:"rollingPeriod"`
	TransmissionRisk int    `json:"transmissionRisk"`
}

func TestInvalidHeader(t *testing.T) {
	t.Parallel()

	body := io.NopCloser(bytes.NewReader([]byte("")))
	r := httptest.NewRequest("POST", "/", body)
	r.Header.Set("content-type", "application/text")

	w := httptest.NewRecorder()
	data := &dataModel{}
	code, err := UnmarshalRequest(w, r, data)

	expCode := http.StatusUnsupportedMediaType
	expErr := "content-type is not application/json"
	if code != expCode {
		t.Errorf("unmarshal wanted %v response code, got %v", expCode, code)
	}

	if err == nil || err.Error() != expErr {
		t.Errorf("expected error '%v', got: %v", expErr, err)
	}
}

func TestEmptyBody(t *testing.T) {
	t.Parallel()

	invalidJSON := []string{
		``,
	}
	errors := []string{
		`body must not be empty`,
	}
	unmarshalTestHelper(t, invalidJSON, errors, http.StatusBadRequest)
}

func TestMultipleJSON(t *testing.T) {
	t.Parallel()

	invalidJSON := []string{
		`{"temporaryExposureKeys":
			[{"key": "ABC"},
			 {"key": "DEF"},
			 {"key": "123"}],
		"appPackageName": "com.google.android.awesome",
		"regions": ["us"]}
		{"temporaryExposureKeys":
			[{"key": "ABC"},
			 {"key": "DEF"},
			 {"key": "123"}],
		"appPackageName": "com.google.android.awesome",
		"regions": ["us"]}`,
	}
	errors := []string{
		"body must contain only one JSON object",
	}
	unmarshalTestHelper(t, invalidJSON, errors, http.StatusBadRequest)
}

func TestInvalidJSON(t *testing.T) {
	t.Parallel()

	invalidJSON := []string{
		`totally not json`,
		`{"key": "value", badKey: 6`,
		`{"exposureKeys": ["ABC", "DEF", "123"],`,
	}
	errors := []string{
		`malformed json at position 2`,
		`malformed json at position 18`,
		`malformed json`,
	}
	unmarshalTestHelper(t, invalidJSON, errors, http.StatusBadRequest)
}

func TestInvalidStructure(t *testing.T) {
	t.Parallel()

	invalidJSON := []string{
		`{"temporaryExposureKeys": 42}`,
		`{"temporaryExposureKeys": ["41", 42]}`,
		`{"appPackageName": 4.5}`,
		`{"regions": "us"}`,
		`{"badField": "doesn't exist"}`,
	}
	errors := []string{
		`invalid value temporaryExposureKeys at position 28`,
		`invalid value temporaryExposureKeys at position 31`,
		`invalid value appPackageName at position 22`,
		`invalid value regions at position 16`,
		`unknown field "badField"`,
	}
	unmarshalTestHelper(t, invalidJSON, errors, http.StatusBadRequest)
}

func TestValidPublishMessage(t *testing.T) {
	t.Parallel()

	intervalNumber := int32(time.Date(2020, 4, 17, 20, 4, 1, 1, time.UTC).Unix() / 600)
	json := `{"temporaryExposureKeys": [
		  {"key": "ABC", "rollingStartNumber": %v, "rollingPeriod": 144, "TransmissionRisk": 2},
		  {"key": "DEF", "rollingStartNumber": %v, "rollingPeriod": 122, "TransmissionRisk": 2},
			{"key": "123", "rollingStartNumber": %v, "rollingPeriod": 1, "TransmissionRisk": 2}],
    "appPackageName": "com.google.android.awesome",
    "regions": ["CA", "US"],
    "VerificationPayload": "1234-ABCD-EFGH-5678"}`
	json = fmt.Sprintf(json, intervalNumber, intervalNumber, intervalNumber)

	body := io.NopCloser(bytes.NewReader([]byte(json)))
	r := httptest.NewRequest("POST", "/", body)
	r.Header.Set("content-type", "application/json")

	w := httptest.NewRecorder()

	got := &dataModel{}
	code, err := UnmarshalRequest(w, r, got)
	if err != nil {
		t.Fatalf("unexpected err, %v", err)
	}
	if code != http.StatusOK {
		t.Errorf("unmarshal wanted %v response code, got %v", http.StatusOK, code)
	}

	want := &dataModel{
		Keys: []dataModelKey{
			{Key: "ABC", IntervalNumber: intervalNumber, IntervalCount: 144, TransmissionRisk: 2},
			{Key: "DEF", IntervalNumber: intervalNumber, IntervalCount: 122, TransmissionRisk: 2},
			{Key: "123", IntervalNumber: intervalNumber, IntervalCount: 1, TransmissionRisk: 2},
		},
		Regions:             []string{"CA", "US"},
		AppPackageName:      "com.google.android.awesome",
		VerificationPayload: "1234-ABCD-EFGH-5678",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unmarshal mismatch (-want +got):\n%v", diff)
	}
}

func unmarshalTestHelper(t *testing.T, payloads []string, errors []string, expCode int) {
	t.Helper()

	for i, testStr := range payloads {
		body := io.NopCloser(bytes.NewReader([]byte(testStr)))
		r := httptest.NewRequest("POST", "/", body)
		r.Header.Set("content-type", "application/json; charset=utf-8")

		w := httptest.NewRecorder()
		data := &dataModel{}
		code, err := UnmarshalRequest(w, r, data)
		if code != expCode {
			t.Errorf("unmarshal wanted %v response code, got %v", expCode, code)
		}
		if errors[i] == "" {
			// No error expected for this test, bad if we got one.
			if err != nil {
				t.Errorf("expected no error for `%v`, got: %v", testStr, err)
			}
		} else {
			if err == nil {
				t.Errorf("wanted error '%v', got nil", errors[i])
			} else if err.Error() != errors[i] {
				t.Errorf("expected error '%v', got: %v", errors[i], err)
			}
		}
	}
}
