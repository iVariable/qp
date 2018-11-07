package processor

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/iVariable/qp/src/qp"
	"github.com/iVariable/qp/src/utils"
	"net/http"
	"strings"
	"time"
)

// HTTPProxy - Proxy request to custom http-endpoint.
// Acknowledge message in case of HTTP response code 200
// Any other response code - reject message
type HTTPProxy struct {
	configuration httpProxyConfiguration
	client        *http.Client
	logger        *log.Entry
}

type httpProxyConfiguration struct {
	Timeout int
	URL     string
}

// Process - Process job
func (h *HTTPProxy) Process(job qp.IJob) error {
	h.logger.WithField("job", job).Debug("Processing job")
	serializedMessage, err := job.GetMessage().Serialize()
	if err != nil {
		h.logger.WithError(err).Error("Error during message serialization")
		return err
	}

	request, err := http.NewRequest("POST", h.configuration.URL, strings.NewReader(serializedMessage))
	if err != nil {
		h.logger.WithError(err).Warn("Error sending HTTP request")
		return err
	}

	resp, err := h.client.Do(request)

	if err == nil && resp.StatusCode != 200 {
		err = errors.New("Response code is not 200. It is: "+resp.Status)
	}

	if err != nil  {
		h.logger.WithError(err).Debug("Job failed")
		if rejectError := job.RejectMessage(); rejectError != nil {
			h.logger.WithError(rejectError).Debug("Error on MessageReject")
			return rejectError
		} else {
			h.logger.Debug("Job rejected")
		}
	} else {
		resp.Body.Close()
		h.logger.Debug("Job Processed")
		if ackError := job.AckMessage(); ackError != nil {
			h.logger.WithError(ackError).Debug("Error on MessageAcknowledge")
			return ackError
		}
		h.logger.Debug("Job acknowledged")
	}

	return nil
}

// Configure - configure processor
func (h *HTTPProxy) Configure(configuration map[string]interface{}) error {
	utils.FillStruct(configuration, &h.configuration)

	if h.configuration.Timeout < 0 {
		return errors.New("Timout setting for HttpProxy should be > 0")
	}

	h.client = &http.Client{
		Timeout: time.Duration(h.configuration.Timeout) * time.Second}

	h.logger = log.WithFields(log.Fields{
		"type":      "processor",
		"processor": "HttpProxy",
	})

	h.logger.WithField("configuration", h.configuration).Info("Configuration loaded")

	return nil
}
