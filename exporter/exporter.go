package exporter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	namespace = "s3"
)

var (
	s3ListSuccess = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "list_success"),
		"If the ListObjects operation was a success",
		[]string{"bucket", "prefix", "delimiter"}, nil,
	)
	s3ListDuration = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "list_duration_seconds"),
		"The total duration of the list operation",
		[]string{"bucket", "prefix", "delimiter"}, nil,
	)
	s3LastModifiedObjectDate = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_modified_object_date"),
		"The last modified date of the object that was modified most recently",
		[]string{"bucket", "prefix"}, nil,
	)
	s3LastModifiedObjectSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "last_modified_object_size_bytes"),
		"The size of the object that was modified most recently",
		[]string{"bucket", "prefix"}, nil,
	)
	s3ObjectTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "objects"),
		"The total number of objects for the bucket/prefix combination",
		[]string{"bucket", "prefix"}, nil,
	)
	s3SumSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "objects_size_sum_bytes"),
		"The total size of all objects summed",
		[]string{"bucket", "prefix"}, nil,
	)
	s3BiggestSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "biggest_object_size_bytes"),
		"The size of the biggest object",
		[]string{"bucket", "prefix"}, nil,
	)
	s3CommonPrefixes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "common_prefixes"),
		"A count of all the keys between the prefix and the next occurrence of the string specified by the delimiter",
		[]string{"bucket", "prefix", "delimiter"}, nil,
	)
)

type (
	// Exporter is our exporter type
	Exporter struct {
		bucket    string
		prefix    string
		delimiter string
		svc       AWSClient
	}
)

// Describe all the metrics we export
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- s3ListSuccess
	ch <- s3ListDuration
	if e.delimiter == "" {
		ch <- s3LastModifiedObjectDate
		ch <- s3LastModifiedObjectSize
		ch <- s3ObjectTotal
		ch <- s3SumSize
		ch <- s3BiggestSize
	} else {
		ch <- s3CommonPrefixes
	}
}

// Collect metrics
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	var (
		lastModified      time.Time
		numberOfObjects   float64
		totalSize         int64
		biggestObjectSize int64
		lastObjectSize    int64
		commonPrefixes    int
	)

	query := &s3.ListObjectsV2Input{
		Bucket:    e.svc.ToString(e.bucket),
		Prefix:    e.svc.ToString(e.prefix),
		Delimiter: e.svc.ToString(e.delimiter),
	}

	// Continue making requests until we've listed and compared the date of every object
	startList := time.Now()
	for {
		resp, err := e.svc.s3.ListObjectsV2(query)
		log.Warnln("call to AWS")
		if err != nil {
			log.Errorln(err)
			ch <- prometheus.MustNewConstMetric(
				s3ListSuccess, prometheus.GaugeValue, 0, e.bucket, e.prefix,
			)
			return
		}
		commonPrefixes = commonPrefixes + len(resp.CommonPrefixes)
		for _, item := range resp.Contents {
			numberOfObjects++
			totalSize = totalSize + *item.Size
			if item.LastModified.After(lastModified) {
				lastModified = *item.LastModified
				lastObjectSize = *item.Size
			}
			if *item.Size > biggestObjectSize {
				biggestObjectSize = *item.Size
			}
		}
		if resp.NextContinuationToken == nil {
			break
		}
		query.ContinuationToken = resp.NextContinuationToken
	}
	listDuration := time.Now().Sub(startList).Seconds()

	ch <- prometheus.MustNewConstMetric(
		s3ListSuccess, prometheus.GaugeValue, 1, e.bucket, e.prefix, e.delimiter,
	)
	ch <- prometheus.MustNewConstMetric(
		s3ListDuration, prometheus.GaugeValue, listDuration, e.bucket, e.prefix, e.delimiter,
	)
	if e.delimiter == "" {
		ch <- prometheus.MustNewConstMetric(
			s3LastModifiedObjectDate, prometheus.GaugeValue, float64(lastModified.UnixNano()/1e9), e.bucket, e.prefix,
		)
		ch <- prometheus.MustNewConstMetric(
			s3LastModifiedObjectSize, prometheus.GaugeValue, float64(lastObjectSize), e.bucket, e.prefix,
		)
		ch <- prometheus.MustNewConstMetric(
			s3ObjectTotal, prometheus.GaugeValue, numberOfObjects, e.bucket, e.prefix,
		)
		ch <- prometheus.MustNewConstMetric(
			s3BiggestSize, prometheus.GaugeValue, float64(biggestObjectSize), e.bucket, e.prefix,
		)
		ch <- prometheus.MustNewConstMetric(
			s3SumSize, prometheus.GaugeValue, float64(totalSize), e.bucket, e.prefix,
		)
	} else {
		ch <- prometheus.MustNewConstMetric(
			s3CommonPrefixes, prometheus.GaugeValue, float64(commonPrefixes), e.bucket, e.prefix, e.delimiter,
		)
	}
}
