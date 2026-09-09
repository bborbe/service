// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("sentryHTTPTransport", func() {
	DescribeTable(
		"selects the transport based on the proxy value",
		func(proxy *string, expectProxy bool) {
			transport := sentryHTTPTransport(proxy)
			if expectProxy {
				Expect(transport).NotTo(BeIdenticalTo(http.DefaultTransport))
			} else {
				Expect(transport).To(BeIdenticalTo(http.DefaultTransport))
			}
		},
		Entry("nil proxy uses the default transport", nil, false),
		Entry("empty proxy uses the default transport", stringPointer(""), false),
		Entry(
			"populated proxy uses the proxy round tripper",
			stringPointer("http://localhost:8080"),
			true,
		),
	)

	Context("with a populated proxy", func() {
		It("rewrites the request URL to the proxy host before sending", func() {
			transport := sentryHTTPTransport(stringPointer("http://localhost:8080"))

			request, err := http.NewRequest(
				"POST",
				"https://o1.ingest.sentry.io/api/123/envelope/",
				strings.NewReader("{}"),
			)
			Expect(err).To(BeNil())

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			request = request.WithContext(ctx)

			_, _ = transport.RoundTrip(request)

			Expect(request.URL.Host).To(Equal("localhost:8080"))
			Expect(request.URL.Scheme).To(Equal("http"))
		})
	})
})

func stringPointer(s string) *string {
	return &s
}
