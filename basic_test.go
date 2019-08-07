// Copyright 2019 Jacques Supcik
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gelakelevel

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	var client = &http.Client{
		Timeout: time.Second * 30,
	}
	l, err := GetLevel(client)
	assert.Nil(t, err)
	assert.Equal(t, l["La Gruyère"].Name, "La Gruyère")
	assert.Greater(t, l["La Gruyère"].MaxLevel, 600.0)
	assert.Less(t, l["La Gruyère"].MaxLevel, 800.0)

	for k, v := range l["La Gruyère"].Measures {
		_, err := time.Parse("2006-01-02", k)
		assert.Nil(t, err)
		assert.LessOrEqual(t, v.Min, v.Max)
	}
}
