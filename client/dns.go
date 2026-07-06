package client

import (
	"context"
	"net"
)

func init() {
	var dialer net.Dialer

	net.DefaultResolver = &net.Resolver{
		PreferGo: false,
		Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, "udp", "1.1.1.1:53")
			if err != nil {
				return dialer.DialContext(ctx, "udp", "8.8.8.8:53")
			}
			return conn, nil
		},
	}
}
