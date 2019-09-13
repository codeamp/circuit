package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromContext(t *testing.T) {
	loggerFromEmptyContext := FromContext(context.Background())
	assert.NotNil(t, loggerFromEmptyContext, "FromContext should return a non-nil logger, even if the context does not contain a logger")

	logger := New("logger-from-context")
	ctx := context.Background()
	ctx = NewContext(ctx, logger)
	loggerFromContext := FromContext(ctx)
	assert.Equal(t, logger, loggerFromContext, "Logger retrieved from context should be the same one we placed in the context")
}
