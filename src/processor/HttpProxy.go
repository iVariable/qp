package processor

import (
	"errors"
	"net/http"
	"qp"
	"strings"
	"time"
	"utils"
	log "github.com/Sirupsen/logrus"
)

type HttpProxy struct {
	configuration httpProxyConfiguration
	client        *http.Client
	logger 		  *log.Entry
}

type httpProxyConfiguration struct {
	Timeout int
	Url     string
}

func (h *HttpProxy) Process(job qp.IJob) error {
	h.logger.WithField("job", job).Debug("Processing job")
	serializedMessage, err := job.GetMessage().Serialize()
	if err != nil {
		h.logger.WithField("error", err).Error("Error during message serialization")
		return err
	}

	request, err := http.NewRequest("POST", h.configuration.Url, strings.NewReader(serializedMessage))
	if err != nil {
		return err
	}

	resp, err := h.client.Do(request)
	if err != nil || resp.StatusCode != 200 {
		if rejectError := job.RejectMessage(); rejectError != nil {
			h.logger.WithField("error", rejectError).Debug("Error on MessageReject")
			return rejectError
		}
		h.logger.Debug("job rejected")
	} else {
		resp.Body.Close()
		if ackError := job.AckMessage(); ackError != nil {
			h.logger.WithField("error", ackError).Debug("Error on MessageAcknowledge")
			return ackError
		}
		h.logger.Debug("job acknowledged")
	}

	return nil
}

func (h *HttpProxy) Configure(configuration map[string]interface{}) error {
	utils.FillStruct(configuration, &h.configuration)

	if h.configuration.Timeout < 0 {
		return errors.New("Timout setting for HttpProxy should be > 0")
	}

	h.client = &http.Client{
		Timeout: time.Duration(h.configuration.Timeout) * time.Second}

	h.logger = log.WithFields(log.Fields{
		"type": "processor",
		"processor": "HttpProxy",
	})

	h.logger.WithField("configuration", h.configuration).Info("Configuration loaded")

	return nil
}
