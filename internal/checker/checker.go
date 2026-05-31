package checker

import "context"

type Checker interface {
	Run(ctx context.Context) Result
}
