// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package metricsplitterprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

var (
	typeStr   = component.MustNewType("metricsplitter")
	stability = component.StabilityLevelDevelopment
)

// Config defines the configuration for the processor
type Config struct {
	// Add configuration fields here
}

// NewFactory creates a factory for the processor
func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.Settings,
	config component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	metricsProcessor := newMetricsSplitterProcessor(config.(*Config), set.Logger)

	return processorhelper.NewMetricsProcessor(
		ctx,
		set,
		config,
		nextConsumer,
		metricsProcessor.ConsumeMetrics,
		processorhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}),
	)
}
