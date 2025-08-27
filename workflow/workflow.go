package workflow

import (
	"io"

	"github.com/nektos/act/pkg/model"
)

type Workflow = model.Workflow

func ReadWorkflow(in io.Reader) (*Workflow, error) {
	return model.ReadWorkflow(in, false)
}
