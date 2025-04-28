package transfer

import (
	"strconv"
	"time"
	. "utils/dsp"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	usePrometheus    bool
	qpsCounterVec    *prometheus.CounterVec
	costHistogramVec *prometheus.HistogramVec
)

func InitPrometheus() {
	usePrometheus = true

	qpsCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "qps",
			Help: "The QPS of each method.",
		},
		[]string{TagHost, TagInst, TagApp, TagMethod, TagStatus},
	)

	costHistogramVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cost",
			Help:    "The COST for each request.",
			Buckets: []float64{0.1, 0.5, 1, 2, 5},
		},
		[]string{TagHost, TagInst, TagApp, TagMethod},
	)

	prometheus.MustRegister(
		qpsCounterVec,
		costHistogramVec,
	)
}

func recordRequest(req *Req, res *Res, cost time.Duration) {
	if !usePrometheus {
		return
	}

	hostStr := HOST()
	instStr := INST()
	app := req.GetApp()
	method := req.GetMethod()

	code := res.GetCode()

	qpsCounterVec.WithLabelValues(
		hostStr,
		instStr,
		app,
		method,
		strconv.Itoa(int(code)),
	).Inc()

	costHistogramVec.WithLabelValues(
		hostStr,
		instStr,
		app,
		method,
	).Observe(cost.Seconds())
}
