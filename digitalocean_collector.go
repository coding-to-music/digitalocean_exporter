package digitaloceanexporter

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

// A DigitalOceanSource is an interface which can retrieve information about a
// resources in a DigitalOcean account. It is implemented by
// *digitaloceanexporter.DigitalOceanService.
type DigitalOceanSource interface {
	Droplets() (map[DropletCounter]int, error)
	FloatingIPs() (map[FlipCounter]int, error)
	LoadBalancers() (map[LoadBalancerCounter]int, error)
	Tags() (map[TagCounter]int, error)
	Volumes() (map[VolumeCounter]int, error)
}

// A DigitalOceanCollector is a Prometheus collector for metrics regarding
// DigitalOcean.
type DigitalOceanCollector struct {
	Droplets      *prometheus.Desc
	FloatingIPs   *prometheus.Desc
	LoadBalancers *prometheus.Desc
	Tags          *prometheus.Desc
	Volumes       *prometheus.Desc

	dos DigitalOceanSource
}

// Verify that DigitalOceanCollector implements the prometheus.Collector interface.
var _ prometheus.Collector = &DigitalOceanCollector{}

// NewDigitalOceanCollector creates a new DigitalOceanCollector which collects
// metrics about resources in a DigitalOcean account.
func NewDigitalOceanCollector(dos DigitalOceanSource) *DigitalOceanCollector {
	return &DigitalOceanCollector{
		Droplets: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "droplets", "count"),
			"Number of Droplets by region, size, and status.",
			[]string{"region", "size", "status"},
			nil,
		),
		FloatingIPs: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "floating_ips", "count"),
			"Number of Floating IPs by region and status.",
			[]string{"region", "status"},
			nil,
		),
		LoadBalancers: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "load_balancers", "count"),
			"Number of Load Balancers by region and status.",
			[]string{"region", "status"},
			nil,
		),
		Tags: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "tags", "count"),
			"Count of tagged resources by name and resource type.",
			[]string{"name", "resource_type"},
			nil,
		),
		Volumes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "volumes", "count"),
			"Number of Volumes by region, size in GiB, and status.",
			[]string{"region", "size", "status"},
			nil,
		),

		dos: dos,
	}
}

// collect begins a metrics collection task for all metrics related to
// resources in a DigitalOcean account.
func (c *DigitalOceanCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	if count, err := c.collectDropletCounts(ch); err != nil {
		return count, err
	}
	if count, err := c.collectFipsCounts(ch); err != nil {
		return count, err
	}
	if count, err := c.collectLoadBalancerCounts(ch); err != nil {
		return count, err
	}
	if count, err := c.collectTagCounts(ch); err != nil {
		return count, err
	}
	if count, err := c.collectVolumeCounts(ch); err != nil {
		return count, err
	}

	return nil, nil
}

func (c *DigitalOceanCollector) collectDropletCounts(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	droplets, err := c.dos.Droplets()
	if err != nil {
		return c.Droplets, err
	}

	for d, count := range droplets {
		ch <- prometheus.MustNewConstMetric(
			c.Droplets,
			prometheus.GaugeValue,
			float64(count),
			d.region,
			d.size,
			d.status,
		)
	}

	return nil, nil
}

func (c *DigitalOceanCollector) collectFipsCounts(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	fips, err := c.dos.FloatingIPs()
	if err != nil {
		return c.FloatingIPs, err
	}

	for fip, count := range fips {
		ch <- prometheus.MustNewConstMetric(
			c.FloatingIPs,
			prometheus.GaugeValue,
			float64(count),
			fip.region,
			fip.status,
		)
	}

	return nil, nil
}

func (c *DigitalOceanCollector) collectLoadBalancerCounts(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	fips, err := c.dos.LoadBalancers()
	if err != nil {
		return c.FloatingIPs, err
	}

	for fip, count := range fips {
		ch <- prometheus.MustNewConstMetric(
			c.LoadBalancers,
			prometheus.GaugeValue,
			float64(count),
			fip.region,
			fip.status,
		)
	}

	return nil, nil
}

func (c *DigitalOceanCollector) collectTagCounts(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	tags, err := c.dos.Tags()
	if err != nil {
		return c.Tags, err
	}

	for t, count := range tags {
		ch <- prometheus.MustNewConstMetric(
			c.Tags,
			prometheus.GaugeValue,
			float64(count),
			t.name,
			t.resourceType,
		)
	}

	return nil, nil
}

func (c *DigitalOceanCollector) collectVolumeCounts(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	volumes, err := c.dos.Volumes()
	if err != nil {
		return c.Volumes, err
	}

	for v, count := range volumes {
		ch <- prometheus.MustNewConstMetric(
			c.Volumes,
			prometheus.GaugeValue,
			float64(count),
			v.region,
			v.size,
			v.status,
		)
	}

	return nil, nil
}

// Describe sends the descriptors of each metric over to the provided channel.
// The corresponding metric values are sent separately.
func (c *DigitalOceanCollector) Describe(ch chan<- *prometheus.Desc) {
	ds := []*prometheus.Desc{
		c.Droplets,
	}

	for _, d := range ds {
		ch <- d
	}
}

// Collect sends the metric values for each metric pertaining to the DigitalOcean
// resources to the provided prometheus Metric channel.
func (c *DigitalOceanCollector) Collect(ch chan<- prometheus.Metric) {
	if desc, err := c.collect(ch); err != nil {
		log.Printf("[ERROR] failed collecting DigitalOcean metric %v: %v", desc, err)
		ch <- prometheus.NewInvalidMetric(desc, err)
		return
	}
}
