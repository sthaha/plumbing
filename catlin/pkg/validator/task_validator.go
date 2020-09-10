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

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"go.uber.org/zap"

	"github.com/tektoncd/plumbing/catlin/pkg/app"
	"github.com/tektoncd/plumbing/catlin/pkg/parser"
)

type TaskValidator struct {
	log *zap.SugaredLogger
	res *parser.Resource
}

var _ Validator = (*TaskValidator)(nil)

func NewTaskValidator(cli app.CLI, r *parser.Resource) *TaskValidator {
	return &TaskValidator{res: r, log: cli.Logger().Sugar()}
}

func (t *TaskValidator) Validate() Result {

	result := Result{}

	res, err := t.res.ToType()
	if err != nil {
		result.Error("failed to decode task - %s", err)
		return result
	}

	for _, step := range res.(*v1beta1.Task).Spec.Steps {

		res, err := name.ParseReference(step.Image, name.StrictValidation)
		if err != nil {
			result.Error("Invalid Image Reference: %s", err)
			continue
		}

		if strings.Contains(res.String(), "latest") {
			result.Error("Task image (%s) must be tagged with a specific version", step.Image)
		}
	}

	return result
}
