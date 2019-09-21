/*
Copyright 2019 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clients

import (
	"context"
	"net/http"

	"github.com/packethost/packngo"
)

// NewClient ... TODO
func NewClient(ctx context.Context, credentials []byte) (*packngo.Client, error) {
	return packngo.NewClientWithBaseURL("crossplane", string(credentials), nil, "")
}

// IsNotFound returns true if error is not found
func IsNotFound(err error) bool {
	if e, ok := err.(*packngo.ErrorResponse); ok {
		return e.Response.StatusCode == http.StatusNotFound
	}
	return false
}
