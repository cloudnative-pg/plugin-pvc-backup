package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/cloudnative-pg/cloudnative-pg/pkg/management/url"
	"github.com/cloudnative-pg/cloudnative-pg/pkg/utils"
	"github.com/cloudnative-pg/plugin-pvc-backup/pkg/logging"
)

// getPgControlData obtains the pg_controldata from the instance HTTP endpoint
func getPgControlData(
	ctx context.Context,
) (map[string]string, error) {
	contextLogger := logging.FromContext(ctx)

	const connectionTimeout = 2 * time.Second
	const requestTimeout = 30 * time.Second

	// We want a connection timeout to prevent waiting for the default
	// TCP connection timeout (30 seconds) on lost SYN packets
	timeoutClient := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: connectionTimeout,
			}).DialContext,
		},
		Timeout: requestTimeout,
	}

	httpURL := url.Build(podIP, url.PathPGControlData, url.StatusPort)
	req, err := http.NewRequestWithContext(ctx, "GET", httpURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := timeoutClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			contextLogger.Error(err, "while closing body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		contextLogger.Info("Error while querying the pg_controldata endpoint",
			"statusCode", resp.StatusCode,
			"body", string(body))
		return nil, fmt.Errorf("error while querying the pg_controldata endpoint: %d", resp.StatusCode)
	}

	type pgControldataResponse struct {
		Data  string `json:"data,omitempty"`
		Error error  `json:"error,omitempty"`
	}

	var result pgControldataResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		result.Error = err
		return nil, err
	}

	return utils.ParsePgControldataOutput(result.Data), result.Error
}
