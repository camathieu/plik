package middleware

import (
	"github.com/root-gg/logger"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	data_test "github.com/root-gg/plik/server/data/testing"
	metadata_test "github.com/root-gg/plik/server/metadata/testing"
)

func newTestingContext(config *common.Configuration) (ctx *context.Context) {
	ctx = &context.Context{}
	ctx.SetConfig(config)
	ctx.SetLogger(logger.NewLogger())
	ctx.SetMetadataBackend(metadata_test.NewBackend())
	ctx.SetDataBackend(data_test.NewBackend())
	ctx.SetStreamBackend(data_test.NewBackend())
	return ctx
}