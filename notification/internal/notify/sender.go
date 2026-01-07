package notify

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"time"
)

type Sender struct {
	client  *http.Client
	webhook string
}

func New(webhook string, timeout time.Duration) *Sender {
	return &Sender{
		client:  &http.Client{Timeout: timeout},
		webhook: webhook,
	}
}

func (s *Sender) Send(ctx context.Context, payload []byte) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, s.webhook, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return errors.New("transient notify error")
	}
	if resp.StatusCode >= 400 {
		return errors.New("permanent notify error")
	}

	return nil
}
