package processor

import (
	"errors"
	"net/http"
	"qp"
	"strings"
	"time"
	"utils"
)

type HttpProxy struct {
	configuration httpProxyConfiguration
	client        *http.Client
}

type httpProxyConfiguration struct {
	Timeout int
	Url     string
}

func (h *HttpProxy) Process(job qp.IJob) error {
	serializedMessage, err := job.GetMessage().Serialize()
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", h.configuration.Url, strings.NewReader(serializedMessage))
	if err != nil {
		return err
	}

	resp, err := h.client.Do(request)
	if err != nil || resp.StatusCode != 200 {
		if rejectError := job.RejectMessage(); rejectError != nil {
			panic(rejectError.Error()) //TODO what to do?
		}
	} else {
		resp.Body.Close()
		if ackError := job.AckMessage(); ackError != nil {
			panic(ackError.Error()) //TODO what to do?
		}
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

	return nil
}
