package middleware

import (
	"github.com/root-gg/juliet"
	"github.com/root-gg/logger"
	"github.com/root-gg/plik/server/common"
	"github.com/root-gg/plik/server/context"
	data_test "github.com/root-gg/plik/server/data/testing"
	metadata_test "github.com/root-gg/plik/server/metadata/testing"
)

func newTestingContext(config *common.Configuration) (ctx *juliet.Context) {
	ctx = juliet.NewContext()
	context.SetConfig(ctx, config)
	context.SetLogger(ctx, logger.NewLogger())
	context.SetMetadataBackend(ctx, metadata_test.NewBackend())
	context.SetDataBackend(ctx, data_test.NewBackend())
	context.SetStreamBackend(ctx, data_test.NewBackend())
	return ctx
}
