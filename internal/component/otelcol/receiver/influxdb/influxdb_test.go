package influxdb_test

import (
    "context"
    "testing"
    "time"
    "fmt"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    influxdb2 "github.com/influxdata/influxdb-client-go/v2"
    influxdb1 "github.com/influxdata/influxdb1-client/v2"
    "go.opentelemetry.io/collector/component/componenttest"
    "go.opentelemetry.io/collector/config/confighttp"
    "go.opentelemetry.io/collector/consumer"
    "go.opentelemetry.io/collector/pdata/pmetric"
    "go.opentelemetry.io/collector/receiver/receivertest"
    influxdbreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/influxdbreceiver"
)

func TestWriteLineProtocol_Alloy(t *testing.T) {
    addr := getFreeAddr(t)
    config := &influxdbreceiver.Config{
        ServerConfig: confighttp.ServerConfig{
            Endpoint: addr,
        },
    }
    nextConsumer := new(mockConsumer)

    receiver, outerErr := influxdbreceiver.NewFactory().CreateMetrics(context.Background(), receivertest.NewNopSettings(), config, nextConsumer)
    require.NoError(t, outerErr)
    require.NotNil(t, receiver)

    require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
    t.Cleanup(func() { require.NoError(t, receiver.Shutdown(context.Background())) })

    // Send test data using InfluxDB client v1
    t.Run("influxdb-client-v1", func(t *testing.T) {
        nextConsumer.lastMetricsConsumed = pmetric.NewMetrics()

        client, err := influxdb1.NewHTTPClient(influxdb1.HTTPConfig{
            Addr:    "http://" + addr,
            Timeout: time.Second,
        })
        require.NoError(t, err)

        batchPoints, err := influxdb1.NewBatchPoints(influxdb1.BatchPointsConfig{Precision: "µs"})
        require.NoError(t, err)
        point, err := influxdb1.NewPoint("cpu_temp", map[string]string{"foo": "bar"}, map[string]any{"gauge": 87.332})
        require.NoError(t, err)
        batchPoints.AddPoint(point)
        err = client.Write(batchPoints)
        require.NoError(t, err)

        metrics := nextConsumer.lastMetricsConsumed
        assert.NotNil(t, metrics)
        assert.Equal(t, 1, metrics.MetricCount())
        metric := metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
        assert.Equal(t, "cpu_temp", metric.Name())
        assert.InEpsilon(t, 87.332, metric.Gauge().DataPoints().At(0).DoubleValue(), 0.001)
    })

    // Send test data using InfluxDB client v2
    t.Run("influxdb-client-v2", func(t *testing.T) {
        nextConsumer.lastMetricsConsumed = pmetric.NewMetrics()

        o := influxdb2.DefaultOptions()
        o.SetPrecision(time.Microsecond)
        client := influxdb2.NewClientWithOptions("http://"+addr, "", o)
        t.Cleanup(client.Close)

        err := client.WriteAPIBlocking("my-org", "my-bucket").WriteRecord(context.Background(), "cpu_temp,foo=bar gauge=87.332")
        require.NoError(t, err)

        metrics := nextConsumer.lastMetricsConsumed
        assert.NotNil(t, metrics)
        assert.Equal(t, 1, metrics.MetricCount())
        metric := metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
        assert.Equal(t, "cpu_temp", metric.Name())
        assert.InEpsilon(t, 87.332, metric.Gauge().DataPoints().At(0).DoubleValue(), 0.001)
    })
}

type mockConsumer struct {
    lastMetricsConsumed pmetric.Metrics
}

func (m *mockConsumer) Capabilities() consumer.Capabilities {
    return consumer.Capabilities{MutatesData: false}
}

func (m *mockConsumer) ConsumeMetrics(_ context.Context, md pmetric.Metrics) error {
    m.lastMetricsConsumed = pmetric.NewMetrics()
    md.CopyTo(m.lastMetricsConsumed)
    return nil
}

// Helper function to get a free port for the server to use
func getFreeAddr(t *testing.T) string {
    t.Helper()
    port := 8086  // Use a suitable port for testing
    return fmt.Sprintf("localhost:%d", port)
}