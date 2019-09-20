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

	widevineSystemUUID = "edef8ba9-79d6-4ace-a3c8-27dcd51d21ed"
	widevinePSSHBoxOffset = 32
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
	defer resp.Body.Close()

	respObj := GetContentKeyResponse{}
	err = xml.Unmarshal(respBody, &respObj)
	if err != nil {
		return drmproxy.ContentKeyResponse{}, fmt.Errorf("unmarshalling GetContentKeyResponse=%v into xml, %w", respObj, err)
	}

	base64KeyID, err := base64EncodedUUID(cfg.KeyID)
	if err != nil {
		return drmproxy.ContentKeyResponse{}, err
	}

	var psshCfgs []drmproxy.Pssh
	for _, drmSystem := range respObj.DRMSystemList.DRMSystems {
		if drmSystem.ContentProtectionData == "" {
			continue
		}

		data, err := psshDataFor(drmSystem, base64KeyID, cfg.ContentID)
		if err != nil {
			return drmproxy.ContentKeyResponse{}, err
		}

		psshCfgs = append(psshCfgs, drmproxy.Pssh{
			Data: data,
			UUID: drmSystem.SystemId,
		})
	}

	return drmproxy.ContentKeyResponse{{
		Pssh:  psshCfgs,
		Key:   respObj.ContentKeyList.ContentKey.Data.Secret.PlainValue,
		KeyID: base64KeyID,
		IV:    respObj.ContentKeyList.ContentKey.ExplicitIV,
	}}, nil
}

func reqBodyWith(cfg drmproxy.ContentCfg) (io.Reader, error) {
	filled := bytes.Buffer{}
	t := template.Must(template.New("req").Parse(reqTempl))
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

func psshDataFor(system DRMSystem, keyID, contentID string) (string, error) {
	if system.SystemId == widevineSystemUUID {
		pssh, err := base64.StdEncoding.DecodeString(system.PSSH)
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(pssh[widevinePSSHBoxOffset:]), nil
	}

	return system.PSSH, nil
}
