package ai

import "context"

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Ping(ctx context.Context) error {
	_ = ctx
	return nil
}
