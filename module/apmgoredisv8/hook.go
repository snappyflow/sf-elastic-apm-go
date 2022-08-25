package apmgoredisv8

import (
	"bytes"
	"context"
	"strings"

	"github.com/go-redis/redis/v8"

	"go.elastic.co/apm/v2"
)

// hook is an implementation of redis.Hook that reports cmds as spans to Elastic APM.
type hook struct{}

// NewHook returns a redis.Hook that reports cmds as spans to Elastic APM.
func NewHook() redis.Hook {
	return &hook{}
}

// BeforeProcess initiates the span for the redis cmd
func (r *hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	cmdName, statement := getCmdDetails(cmd)
	span, ctx := apm.StartSpan(ctx, cmdName, "db.redis.query")

	span.Context.SetDatabase(apm.DatabaseSpanContext{
		Type:      "redis",
		Statement: statement,
	})
	return ctx, nil
}

// AfterProcess ends the initiated span from BeforeProcess
func (r *hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if span := apm.SpanFromContext(ctx); span != nil {
		span.End()
	}
	return nil
}

// BeforeProcessPipeline initiates the span for the redis cmds
func (r *hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	// Join all cmd names with ", ".
	var cmdNameBuf bytes.Buffer
	var cmdArgsBuf bytes.Buffer
	for i, cmd := range cmds {
		if i != 0 {
			cmdNameBuf.WriteString(", ")
			cmdArgsBuf.WriteString(", ")
		}
		cmdName, statement := getCmdDetails(cmd)
		cmdNameBuf.WriteString(cmdName)
		cmdArgsBuf.WriteString(statement)
	}

	span, ctx := apm.StartSpan(ctx, cmdNameBuf.String(), "db.redis.query")
	span.Context.SetDatabase(apm.DatabaseSpanContext{
		Type:      "redis",
		Statement: cmdArgsBuf.String(),
	})
	return ctx, nil
}

// AfterProcess ends the initiated span from BeforeProcessPipeline
func (r *hook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if span := apm.SpanFromContext(ctx); span != nil {
		span.End()
	}
	return nil
}

func getCmdDetails(cmd redis.Cmder) (string, string) {
	cmdName := strings.ToUpper(cmd.Name())
	if cmdName == "" {
		cmdName = "(empty command)"
	}

	var statement string
	for i, arg := range cmd.Args() {
		if i == 0 {
			arg = strings.ToUpper(arg.(string))
		}
		argString := arg.(string)
		statement = statement + " " + argString
	}
	if statement == "" && cmdName != "PING" {
		statement = "(empty args)"
	}

	return cmdName, strings.TrimSpace(statement)
}
