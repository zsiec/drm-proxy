package irdeto

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cbsinteractive/drm-proxy/pkg/drmproxy"
	"github.com/google/uuid"
)

const (
	headerKeyAuthorization = "Authorization"

	pathLicenseKeyRequestTmpl = "/tkm/v1/cbsi/contents/%s/copyProtectionData"
)

type KeyService interface {
	ContentKeyFrom(cfg drmproxy.ContentCfg) (drmproxy.ContentKeyResponse, error)
}

type IrdetoKeySvc struct {
	endpoint, token string
	client          http.Client
}

func NewKeyService(endpoint, token string, client http.Client) *IrdetoKeySvc {
	return &IrdetoKeySvc{
		endpoint: endpoint,
		token:    token,
		client:   client,
	}
}

func (s *IrdetoKeySvc) ContentKeyFrom(cfg drmproxy.ContentCfg) (drmproxy.ContentKeyResponse, error) {
	reqBody, err := reqBodyWith(cfg)
	if err != nil {
		return drmproxy.ContentKeyResponse{}, fmt.Errorf("creating request body from %v, %w", cfg, err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(
		"%s"+pathLicenseKeyRequestTmpl, s.endpoint, cfg.ContentID,
	), reqBody)

	req.Header.Add(headerKeyAuthorization, fmt.Sprintf("Basic %s", s.token))

	resp, err := s.client.Do(req)
	if err != nil {
		return drmproxy.ContentKeyResponse{}, fmt.Errorf("making request for key: %w", err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return drmproxy.ContentKeyResponse{}, fmt.Errorf("reading resp body from io.Reader %v, %w", resp.Body, err)
	}

	respObj := GetContentKeyResponse{}
	err = xml.Unmarshal(respBody, &respObj)
	if err != nil {
		return drmproxy.ContentKeyResponse{}, fmt.Errorf("unmarshalling GetContentKeyResponse=%v into xml, %w", respObj, err)
	}

	base64UUID, err := base64EncodedUUID(cfg.KeyID)
	if err != nil {
		return drmproxy.ContentKeyResponse{}, err
	}

	return drmproxy.ContentKeyResponse{{
		Key:   respObj.ContentKeyList.ContentKey.Data.Secret.PlainValue,
		KeyID: base64UUID,
		IV:    respObj.ContentKeyList.ContentKey.ExplicitIV,
	}}, nil
}

func reqBodyWith(cfg drmproxy.ContentCfg) (io.Reader, error) {
	filled := bytes.Buffer{}
	t := template.Must(template.New("fpReq").Parse(fairplayReqTempl))
	err := t.Execute(&filled, cfg)
	if err != nil {
		return nil, err
	}

	return &filled, nil
}

func base64EncodedUUID(rawUUID string) (string, error) {
	uuid, err := uuid.Parse(rawUUID)
	if err != nil {
		return "", fmt.Errorf("parsing %q into a uuid: %w", rawUUID, err)
	}

	paddedBase64 := fmt.Sprintf("%-24s", base64.RawStdEncoding.EncodeToString(uuid[:]))

	return strings.Replace(paddedBase64, " ", "=", -1), nil
}
