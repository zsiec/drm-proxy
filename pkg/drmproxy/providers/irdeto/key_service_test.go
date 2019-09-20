package irdeto

import (
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cbsinteractive/drm-proxy/pkg/drmproxy"
	"gotest.tools/assert"
)

type fakeIrdetoKeyService struct {
	shouldErr   bool
	respondWith drmproxy.ContentKeyResponse
}

func (s fakeIrdetoKeyService) ContentKeyFrom(cfg drmproxy.ContentCfg) (drmproxy.ContentKeyResponse, error) {
	if s.shouldErr {
		return drmproxy.ContentKeyResponse{}, errors.New("forced by test")
	}

	return s.respondWith, nil
}

func TestContentKeyFrom(t *testing.T) {
	tests := []struct {
		name                    string
		token, contentID, keyID string
		returnKey, returnIV     string
		wantKeyID               string
		handler                 func(t *testing.T, w http.ResponseWriter, r *http.Request)
		wantErr                 bool
	}{
		{
			name:      "a valid request returns the expected response with key, keyID and IV",
			token:     "someToken",
			contentID: "someContentID",
			keyID:     "86f13e28-eded-4e58-946f-67c51ddd09e4",
			returnKey: "someKey",
			returnIV:  "someIV",
			wantKeyID: "hvE+KO3tTliUb2fFHd0J5A==",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				if g, e := r.Header.Get(headerKeyAuthorization), "Basic someToken"; g != e {
					t.Errorf("got auth header with value %q, expected %q", g, e)
				}

				if g, e := r.URL.Path, "/tkm/v1/cbsi/contents/someContentID/copyProtectionData"; g != e {
					t.Errorf("got req path %q, expected %q", g, e)
				}

				resp := GetContentKeyResponse{
					ContentKeyList: ContentKeyList{
						ContentKey: ContentKey{
							ExplicitIV: "someIV",
							Data: ContentKeyData{
								Secret: ContentKeySecret{
									PlainValue: "someKey",
								},
							},
						},
					},
				}

				b, err := xml.Marshal(resp)
				if err != nil {
					t.Errorf("marshalling response to xml: %v", err)
				}

				w.Write(b)
			},
		},
		{
			name:    "an erroring api results in a useful error",
			wantErr: true,
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
				if tt.handler != nil {
					tt.handler(t, w, r)
				}
			}
			server := httptest.NewServer(handler)

			svc := IrdetoKeySvc{
				endpoint: server.URL,
				token:    tt.token,
				client:   *server.Client(),
			}

			resp, err := svc.ContentKeyFrom(drmproxy.ContentCfg{
				ContentID: tt.contentID,
				KeyID:     tt.keyID,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("ContentKeyFrom() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err != nil {
				return
			}

			assert.Assert(t, resp[0].Key == tt.returnKey)
			if g, e := resp[0].KeyID, tt.wantKeyID; g != e {
				t.Errorf("got keyID with value %q, expected %q", g, e)
			}
			assert.Assert(t, resp[0].IV == tt.returnIV)
		})
	}

}
