package execution

import (
	"errors"

	"github.com/platonoff-dev/coredb/pkg/corekv/engines"
)

type VMContext struct {
}

type VM struct {
	context *VMContext

	storage engines.Engine
}

func (vm *VM) Execute(program []Instruction) error {
	return errors.New("not implemented")
}
