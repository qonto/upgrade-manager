package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qonto/upgrade-manager/config"
	"github.com/qonto/upgrade-manager/internal/app/calculators"
	soft "github.com/qonto/upgrade-manager/internal/app/core/software"
	"github.com/qonto/upgrade-manager/internal/app/sources/argohelm"
	eksSource "github.com/qonto/upgrade-manager/internal/app/sources/aws/eks"
	elasticacheSource "github.com/qonto/upgrade-manager/internal/app/sources/aws/elasticache"
	lambdaSource "github.com/qonto/upgrade-manager/internal/app/sources/aws/lambda"
	mskSource "github.com/qonto/upgrade-manager/internal/app/sources/aws/msk"
	rdsSource "github.com/qonto/upgrade-manager/internal/app/sources/aws/rds"
	"github.com/qonto/upgrade-manager/internal/app/sources/deployments"
	"github.com/qonto/upgrade-manager/internal/app/sources/filesystemhelm"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"github.com/qonto/upgrade-manager/internal/infra/kubernetes"
)

type App struct {
	log       *slog.Logger
	sources   []soft.Source
	softwares []*soft.Software
	Config    config.Config
	registry  *prometheus.Registry
	metrics   appMetrics
	k8sClient kubernetes.KubernetesClient
	s3Api     aws.S3Api
	escApi    aws.ElasticacheApi
	eksApi    aws.EKSApi
	rdsApi    aws.RDSApi
	mskApi    aws.MSKApi
	lambdaApi aws.LambdaApi
	done      chan bool
	wg        sync.WaitGroup
}
type appMetrics struct {
	scores              *prometheus.GaugeVec
	successLoads        prometheus.Gauge
	foundSoftwares      prometheus.Gauge
	successComputeScore prometheus.Gauge
	processError        *prometheus.CounterVec
	loopExecTime        prometheus.Gauge
}

func New(l *slog.Logger, registry *prometheus.Registry, k8sClient kubernetes.KubernetesClient, config config.Config) (*App, error) {
	app := &App{
		log:       l,
		registry:  registry,
		done:      make(chan bool, 1),
		k8sClient: k8sClient,
		Config:    config,
	}
	// TODO: make region mandatory ?
	// helm sources cannot work withou
	if app.Config.Global.AwsConfig.Region != "" {
		awscfg, err := awsConfig.LoadDefaultConfig(context.TODO())
		awscfg.Region = app.Config.Global.AwsConfig.Region
		app.log.Info(fmt.Sprintf("Initializing AWS configuration in region %s", awscfg.Region))
		if err != nil {
			return app, err
		}
		app.s3Api = s3.NewFromConfig(awscfg)
		app.escApi = elasticache.NewFromConfig(awscfg)
		app.eksApi = eks.NewFromConfig(awscfg)
		app.rdsApi = rds.NewFromConfig(awscfg)
		app.lambdaApi = lambda.NewFromConfig(awscfg)
		app.mskApi = kafka.NewFromConfig(awscfg)
	}
	if err := app.InitSources(); err != nil {
		return app, err
	}
	if err := app.InitPrometheusMetrics(); err != nil {
		return app, err
	}
	return app, nil
}

// Initizalize the different software sources which have a config section specified
func (a *App) InitSources() error {
	a.sources = nil

	if a.Config.Sources.FsHelm != nil {
		for _, item := range a.Config.Sources.FsHelm {
			softSource, err := filesystemhelm.NewSource(item, a.log, a.s3Api)
			if err != nil {
				a.log.Error(fmt.Sprint(err))
				os.Exit(1)
			}
			a.sources = append(a.sources, softSource)
		}
	}
	if a.Config.Sources.ArgocdHelm != nil {
		for _, item := range a.Config.Sources.ArgocdHelm {
			softSource, err := argohelm.NewSource(item, a.log, a.k8sClient, true, a.s3Api)
			if err != nil {
				a.log.Error(fmt.Sprint(err))
				os.Exit(1)
			}
			a.sources = append(a.sources, softSource)
		}
	}
	if a.Config.Sources.Deployments != nil {
		for _, item := range a.Config.Sources.Deployments {
			softSource, err := deployments.NewSource(a.log, a.k8sClient, item)
			if err != nil {
				a.log.Error(fmt.Sprint(err))
				os.Exit(1)
			}
			a.sources = append(a.sources, softSource)
		}
	}
	if a.Config.Sources.Aws.Elasticache.Enabled {
		escSource, err := elasticacheSource.NewSource(a.escApi, a.log, &a.Config.Sources.Aws.Elasticache)
		if err != nil {
			a.log.Error(fmt.Sprint(err))
			os.Exit(1)
		}
		a.sources = append(a.sources, escSource)
	}
	if a.Config.Sources.Aws.Eks.Enabled {
		eksSource, err := eksSource.NewSource(a.eksApi, a.log, &a.Config.Sources.Aws.Eks)
		if err != nil {
			a.log.Error(fmt.Sprint(err))
			os.Exit(1)
		}
		a.sources = append(a.sources, eksSource)
	}
	if a.Config.Sources.Aws.Msk.Enabled {
		mskSource, err := mskSource.NewSource(a.mskApi, a.log, &a.Config.Sources.Aws.Msk)
		if err != nil {
			a.log.Error(fmt.Sprint(err))
			os.Exit(1)
		}
		a.sources = append(a.sources, mskSource)
	}
	if a.Config.Sources.Aws.Rds.Enabled {
		rdsSource, err := rdsSource.NewSource(a.rdsApi, a.log, &a.Config.Sources.Aws.Rds)
		if err != nil {
			a.log.Error(fmt.Sprint(err))
			os.Exit(1)
		}
		a.sources = append(a.sources, rdsSource)
	}
	if a.Config.Sources.Aws.Lambda.Enabled {
		lambdaSource, err := lambdaSource.NewSource(a.lambdaApi, a.log, &a.Config.Sources.Aws.Lambda)
		if err != nil {
			a.log.Error(fmt.Sprint(err))
			os.Exit(1)
		}
		a.sources = append(a.sources, lambdaSource)
	}
	return nil
}

func (a *App) InitPrometheusMetrics() error {
	labels := []string{"app", "app_type", "current_version", "target_version", "isparent", "parent", "action"}
	scores := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "upgrade_manager_software_obsolescence_score",
		Help: "obsolescence score for softwares discovered by upgrade-manager app",
	}, labels)
	peLabels := []string{"app", "app_type", "isparent", "parent", "error_type"}
	processError := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "upgrade_manager_software_process_error",
		Help: "errors while processing softwares",
	}, peLabels)
	foundSoftwares := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "upgrade_manager_total_software_found",
		Help: "Total number of softwares found in the auto-discovery process",
	})
	successLoads := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "upgrade_manager_total_software_load_success",
		Help: "Total amount of software with successfully loaded candidates",
	})
	successComputeScore := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "upgrade_manager_total_software_obsolescence_score_compute_success",
		Help: "Total amount of software with successfully computed obsolescence score",
	})
	loopExecTime := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "upgrade_manager_main_loop_execution_time",
		Help: "Time taken by the last main loop execution (find software, check versions and compute score)",
	})
	a.metrics.scores = scores
	a.metrics.processError = processError
	a.metrics.foundSoftwares = foundSoftwares
	a.metrics.successLoads = successLoads
	a.metrics.successComputeScore = successComputeScore
	a.metrics.loopExecTime = loopExecTime
	err := a.registry.Register(scores)
	if err != nil {
		return err
	}
	err = a.registry.Register(processError)
	if err != nil {
		return err
	}
	err = a.registry.Register(successComputeScore)
	if err != nil {
		return err
	}
	err = a.registry.Register(successLoads)
	if err != nil {
		return err
	}
	err = a.registry.Register(foundSoftwares)
	if err != nil {
		return err
	}
	err = a.registry.Register(loopExecTime)
	if err != nil {
		return err
	}
	return nil
}

// Run the app's main process
func (a *App) Start() {
	interval, err := time.ParseDuration(a.Config.Global.Interval)
	if err != nil {
		a.log.Error(fmt.Sprintf("Failed parsing time interval. Using a default interval of 1h, %v", err))
		interval = time.Hour
	}
	ticker := time.NewTicker(interval)

	go func() {
		a.mainLoop()
		defer a.wg.Done()
		for {
			select {
			case <-ticker.C:
				a.reset()
				a.mainLoop()

			case <-a.done:
				return
			}
		}
	}()
	a.wg.Add(1)
}

func (a *App) Stop() {
	close(a.done)
	a.wg.Wait()
}

func (a *App) reset() {
	a.softwares = nil
	a.metrics.processError.MetricVec.Reset()
	a.metrics.successLoads.Set(0)
	a.metrics.successComputeScore.Set(0)
}

func (a *App) mainLoop() {
	startTime := time.Now()

	// Load here all apps as softwares
	a.loadSoftwares()
	a.metrics.successLoads.Add(float64(len(a.softwares)))
	a.log.Info(fmt.Sprintf("Found %d software(s) in total", len(a.softwares)))
	a.metrics.foundSoftwares.Set(float64(len(a.softwares)))
	// Process each software
	for _, software := range a.softwares {
		a.log.Debug("computing obsolescence score", slog.String("software", software.Name))
		if err := a.scoreSoftware(software); err != nil {
			a.log.Error("failed to compute obsolescence score", slog.String("software", software.Name), slog.String("software_type", string(software.Type)))
			a.metrics.processError.WithLabelValues(software.Name, string(software.Type), "1", software.Name, "compute score").Add(1)
			continue
		}
		a.metrics.successComputeScore.Add(1)
		a.log.Debug("obsolescence score computed", slog.String("software", software.Name), slog.String("software_type", string(software.Type)), slog.Int("score", software.CalculatedScore))
	}

	// not in reset function because we want to wait as much as possible to avoid empty metrics
	a.metrics.scores.Reset()

	// print report and update metrics
	a.report()
	duration := time.Since(startTime)
	a.metrics.loopExecTime.Set(duration.Seconds())
	a.log.Info(fmt.Sprintf("Main loop execution time: %f seconds", duration.Seconds()))
}

// Load Softwares from all sources
func (a *App) loadSoftwares() {
	for _, source := range a.sources {
		found, err := source.Load()
		if err != nil {
			a.log.Error(fmt.Sprintf("failed to load softwares %v", err), slog.String("source", source.Name()))
			a.metrics.processError.WithLabelValues("", "", "", "", "load software").Add(1)
		}
		a.softwares = append(a.softwares, found...)
		a.log.Info(fmt.Sprintf("Found %d software(s)", len(found)), slog.String("source", source.Name()))
	}
}

// Calculate obsolescence score for the software
func (a *App) scoreSoftware(software *soft.Software) error {
	c := calculators.New(a.log, software.Calculator, true)
	err := c.CalculateObsolescenceScore(software)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) report() {
	// Recap
	a.log.Debug("")
	a.log.Debug("")
	a.log.Debug("************************************* RECAP **********************************************")
	for idx, software := range a.softwares {
		a.log.Debug(fmt.Sprintf("%d) Application '%s' of type '%s'-> Score: %d", idx+1, software.Name, software.Type, software.CalculatedScore))
		if software.CalculatedScore > 0 {
			if len(software.VersionCandidates) > 0 {
				a.metrics.scores.WithLabelValues(software.Name, string(software.Type), software.Version.Version, software.VersionCandidates[0].Version, "1", software.Name, "update").Set(float64(software.CalculatedScore))
				a.log.Debug("--> update: ", slog.String("software", software.Name), slog.String("software_type", string(software.Type)), slog.String("parent", "self"), slog.String("version", software.Version.Version), slog.String("target_version", software.VersionCandidates[0].Version))
			} else {
				for i, dep := range software.Dependencies {
					if len(dep.VersionCandidates) > 0 {
						a.metrics.scores.WithLabelValues(software.Name, string(software.Type), software.Version.Version, "", "1", software.Name, "update_dependencies").Set(float64(software.CalculatedScore))
						a.metrics.scores.WithLabelValues(dep.Name, string(dep.Type), dep.Version.Version, dep.VersionCandidates[0].Version, "0", software.Name, "update").Set(float64(dep.CalculatedScore))
						a.log.Debug(fmt.Sprintf("--> update %d: ", i+1), slog.String("software", dep.Name), slog.String("software_type", string(dep.Type)), slog.String("parent", software.Name), slog.String("version", dep.Version.Version), slog.String("target_version", dep.VersionCandidates[0].Version))
					}
				}
			}
		}
		if software.CalculatedScore == 0 {
			a.metrics.scores.WithLabelValues(software.Name, string(software.Type), software.Version.Version, "", "1", software.Name, "").Set(float64(software.CalculatedScore))
		}
		a.log.Debug("------------------------------------------------------------------------")
	}
}
