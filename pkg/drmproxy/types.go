package drmproxy

type KeyService interface {
	ContentKeyFrom(cfg ContentCfg) (ContentKeyResponse, error)
}

// ContentKeyResponse is a structured representation the proxy's response
// to a request for an asset's key information following a structure recognized
// by the nginx-vod-module
type ContentKeyResponse []struct {
	Pssh []struct {
		Data string `json:"data"`
		UUID string `json:"uuid"`
	} `json:"pssh"`
	Key   string `json:"key"`
	KeyID string `json:"key_id"`
	IV    string `json:"iv"`
}

// ContentCfg holds information about an asset
type ContentCfg struct {
	ContentID, KeyID string
}
