package runtime

import (
	"errors"
	"log"
)

var (
	ErrNotImplemented       = errors.New("evaluation for this node type not implemented")
	ErrNotFound             = errors.New("identifier not found")
	ErrInternalFuncNotFound = errors.New("internal function not found")
	ErrUnsupportedType      = errors.New("unsupported type for operation")
	ErrInvalidType          = errors.New("invalid type")
)

func ensureNoErr(err error, args ...any) error {
	if err != nil {
		if len(args) > 0 {
			msg := args[0].(string)
			args = args[1:]
			log.Printf(msg, args...)
		}
		panic(err)
	}
	return err
}
