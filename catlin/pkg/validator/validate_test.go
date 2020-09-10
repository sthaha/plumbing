// Copyright © 2020 The Tekton Authors.
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

package validator

import (
	"strings"
	"testing"

	"github.com/tektoncd/plumbing/catlin/pkg/app"
	"github.com/tektoncd/plumbing/catlin/pkg/parser"
	"go.uber.org/zap"
	"gotest.tools/v3/assert"
)

var validTask = `
---
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: valid
  labels:
    app.kubernetes.io/version: a,b,c
  annotations:
    tekton.dev/tags: a,b,c
    tekton.dev/pipelines.minVersion: "0.12"
    tekton.dev/displayName: My Example Task
spec:
  description: |-
    A summary of the resource

    A para about this valid task

  steps:
  - name: hello
    image: abc.io/ubuntu:1.0
    command: [sleep, infinity]
  - name: foo-bar
    image: abc.io/fedora:1.0@sha256:deadb33fdeadb33fdeadb33fdeadb33fdeadb33fdeadb33fdeadb33fdeadb33f
`

var taskWithInvalidImageRef = `
---
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: valid
  labels:
    app.kubernetes.io/version: a,b,c
  annotations:
    tekton.dev/tags: a,b,c
    tekton.dev/pipelines.minVersion: "0.12"
    tekton.dev/displayName: My Example Task
spec:
  description: |-
    A summary of the resource

    A para about this valid task

  steps:
  - name: hello
    image: ubuntu
    command: [sleep, infinity]
  - name: foo
    image: abc.io/fedora:latest
  - name: bar
    image: abc.io/fedora:1.0@sha256:deadb33fdeadb33fdeadb33f
  - name: valid
    image: abc.io/ubuntu:1.0
`

var validPipeline = `
---
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: valid
  labels:
    app.kubernetes.io/version: a,b,c
  annotations:
    tekton.dev/tags: a,b,c
    tekton.dev/pipelines.minVersion: "0.12"
    tekton.dev/displayName: My Example Task
spec:
  description: |-
    A summary of the resource

    A para about this valid task

  tasks:
  - name: hello
    taskRef:
      name: hello
`

type TestConfig struct {
	log *zap.Logger
}

var _ app.CLI = (*TestConfig)(nil)

func (t *TestConfig) Logger() *zap.Logger {
	return t.log
}

func (t *TestConfig) Stream() *app.Stream {
	return nil
}

func TestContentValidator_Task(t *testing.T) {
	log, _ := zap.NewDevelopment()

	tc := &TestConfig{log: log}

	r := strings.NewReader(validTask)
	parser := parser.ForReader(r)

	res, err := parser.Parse()
	assert.NilError(t, err)

	v := NewContentValidator(tc, res)
	result := v.Validate()

	//t.Logf("%v", result.Lints)
	assert.Equal(t, 0, result.Errors)
	assert.Equal(t, 0, len(result.Lints))
}

func TestContentValidator_Pipeline(t *testing.T) {
	log, _ := zap.NewDevelopment()

	tc := &TestConfig{log: log}

	r := strings.NewReader(validPipeline)
	parser := parser.ForReader(r)

	res, err := parser.Parse()
	assert.NilError(t, err)

	v := NewContentValidator(tc, res)
	result := v.Validate()

	assert.Equal(t, 0, result.Errors)
	assert.Equal(t, 0, len(result.Lints))
}

func TestValidatorForKind_Task(t *testing.T) {
	log, _ := zap.NewDevelopment()

	tc := &TestConfig{log: log}

	r := strings.NewReader(validTask)
	parser := parser.ForReader(r)

	res, err := parser.Parse()
	assert.NilError(t, err)

	v := ForKind(tc, res)
	result := v.Validate()

	assert.Equal(t, 0, result.Errors)
	assert.Equal(t, 0, len(result.Lints))
}

func TestValidatorForKind_Task_InvalidImageTag(t *testing.T) {
	log, _ := zap.NewDevelopment()

	tc := &TestConfig{log: log}

	r := strings.NewReader(taskWithInvalidImageRef)
	parser := parser.ForReader(r)

	res, err := parser.Parse()
	assert.NilError(t, err)

	v := ForKind(tc, res)
	result := v.Validate()

	assert.Equal(t, 3, result.Errors)
	assert.Equal(t, 3, len(result.Lints))
	assert.Equal(t, "Invalid Image Reference: could not parse reference: ubuntu", result.Lints[0].Message)
	assert.Equal(t, "Task image (abc.io/fedora:latest) must be tagged with a specific version", result.Lints[1].Message)
	assert.Equal(t, "Invalid Image Reference: could not parse reference: abc.io/fedora:1.0@sha256:deadb33fdeadb33fdeadb33f", result.Lints[2].Message)
}
