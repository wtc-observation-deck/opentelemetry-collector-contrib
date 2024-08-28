// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package metricsplitterprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type MetricsSplitterProcessor struct {
	config *Config
	logger *zap.Logger
}

func newMetricsSplitterProcessor(config *Config, logger *zap.Logger) *MetricsSplitterProcessor {
	return &MetricsSplitterProcessor{config: config, logger: logger}
}

func (msp *MetricsSplitterProcessor) ConsumeMetrics(_ context.Context, metrics pmetric.Metrics) (pmetric.Metrics, error) {
	resourceMetrics := metrics.ResourceMetrics()
	for i := 0; i < resourceMetrics.Len(); i++ {
		scopeMetrics := resourceMetrics.At(i).ScopeMetrics()
		for j := 0; j < scopeMetrics.Len(); j++ {
			metrics := scopeMetrics.At(j).Metrics()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)
				if metric.Type() == pmetric.MetricTypeSummary {
					msp.logger.Debug("Processing summary metric", zap.String("metric_name", metric.Name()))
					dataPoints := metric.Summary().DataPoints()
					for l := 0; l < dataPoints.Len(); l++ {
						dataPoint := dataPoints.At(l)
						sum := dataPoint.Sum()
						count := dataPoint.Count()
						if count > 0 {
							average := sum / float64(count)
							msp.logger.Debug("Summary metric details",
								zap.String("metric_name", metric.Name()),
								zap.Float64("sum", sum),
								zap.Uint64("count", count),
								zap.Float64("average", average))

							// Create a new metric for the average
							newMetric := metrics.AppendEmpty()
							newMetric.SetName(metric.Name() + "_avg")
							newMetric.SetUnit(metric.Unit())
							newMetric.SetEmptyGauge()
							newDataPoint := newMetric.Gauge().DataPoints().AppendEmpty()
							newDataPoint.SetStartTimestamp(dataPoint.StartTimestamp())
							newDataPoint.SetTimestamp(dataPoint.Timestamp())
							newDataPoint.SetDoubleValue(average)

							// Copy original labels and attributes
							dataPoint.Attributes().Range(func(k string, v pcommon.Value) bool {
								switch v.Type() {
								case pcommon.ValueTypeStr:
									newDataPoint.Attributes().PutStr(k, v.Str())
								case pcommon.ValueTypeInt:
									newDataPoint.Attributes().PutInt(k, v.Int())
								case pcommon.ValueTypeDouble:
									newDataPoint.Attributes().PutDouble(k, v.Double())
								case pcommon.ValueTypeBool:
									newDataPoint.Attributes().PutBool(k, v.Bool())
								}
								return true
							})
						}
					}
				}
			}
		}
	}
	return metrics, nil
}

func (msp *MetricsSplitterProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (msp *MetricsSplitterProcessor) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (msp *MetricsSplitterProcessor) Shutdown(_ context.Context) error {
	return nil
}
