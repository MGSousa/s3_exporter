package main

import (
	"net/http"
	"os"
	"time"

	"github.com/MGSousa/s3_exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	prometheus.MustRegister(version.NewCollector("s3_exporter"))
}

func main() {
	var (
		app            = kingpin.New("s3_exporter", "Export metrics for S3 certificates").DefaultEnvars()
		listenAddress  = app.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9340").String()
		metricsPath    = app.Flag("web.metrics-path", "Path under which to expose metrics").Default("/metrics").String()
		probePath      = app.Flag("web.probe-path", "Path under which to expose the probe endpoint").Default("/probe").String()
		discoveryPath  = app.Flag("web.discovery-path", "Path under which to expose service discovery").Default("/discovery").String()
		endpointURL    = app.Flag("s3.endpoint-url", "Custom endpoint URL").Default("").String()
		disableSSL     = app.Flag("s3.disable-ssl", "Custom disable SSL").Bool()
		forcePathStyle = app.Flag("s3.force-path-style", "Custom force path style").Bool()
		// useCaching     = app.Flag("s3.use-cache", "Use k:v for caching S3 results").Bool()
	)

	log.AddFlags(app)
	app.Version(version.Print("s3_exporter"))
	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// TODO: Use with a awscli

	// ------ aws --------

	// use with localstack
	// awsRegion := "us-east-1"
	// customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
	// 	if *endpointURL != "" {
	// 		return aws.Endpoint{
	// 			PartitionID:   "aws",
	// 			URL:           *endpointURL,
	// 			SigningRegion: awsRegion,
	// 		}, nil
	// 	}

	// 	// returning EndpointNotFoundError will allow the service to fallback to its default resolution
	// 	return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	// })

	// awsCfg, err := config.LoadDefaultConfig(context.TODO(),
	// 	config.WithRegion(awsRegion),
	// 	config.WithEndpointResolverWithOptions(customResolver),
	// 	// config.WithHTTPClient(
	// 	// 	&http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}),
	// )
	// if err != nil {
	// 	log.Fatalf("Cannot load the AWS configs: %s", err)
	// }

	// Create the resource client
	// svc := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
	// 	o.UsePathStyle = true
	// 	// o.UsePathStyle = *forcePathStyle
	// 	// o.EndpointOptions.DisableHTTPS = *disableSSL
	// })

	// log.Infoln(svc, *endpointURL)
	// ---- localstack -----

	aws := exporter.NewAwsSession(*endpointURL, *disableSSL, *forcePathStyle)
	if aws == nil {
		log.Fatalln("error setting up AWS Session")
	}

	log.Infoln("Starting s3_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc(*probePath, func(w http.ResponseWriter, r *http.Request) {
		exporter.ProbeHandler(w, r, aws)
	})
	http.HandleFunc(*discoveryPath, func(w http.ResponseWriter, r *http.Request) {
		exporter.DiscoveryHandler(w, r, aws)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`<html>
						 <head><title>AWS S3 Exporter</title></head>
						 <body>
						 <h1>AWS S3 Exporter</h1>
						 <p><a href="` + *probePath + `?bucket=BUCKET&prefix=PREFIX">Query metrics for objects in BUCKET that match PREFIX</a></p>
						 <p><a href='` + *metricsPath + `'>Metrics</a></p>
						 <p><a href='` + *discoveryPath + `'>Service Discovery</a></p>
						 </body>
						 </html>`)); err != nil {
			log.Fatalln(err)
		}
	})

	log.Infoln("Listening on", *listenAddress)

	srv := &http.Server{
		Addr:              *listenAddress,
		ReadHeaderTimeout: 30 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Errorf("http server quit with error: %v", err)
	}
}
