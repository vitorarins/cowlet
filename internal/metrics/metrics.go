package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	once   sync.Once
	metric *metrics
)

type Metrics interface {
	OxygenSaturationSet(int)
	HeartRateSet(int)
	BatteryPercentageSet(int)
	BatteryMinutesSet(int)
	SignalStrengthSet(int)
	OxygenTenAVSet(int)
	SockConnectionSet(int)
	SleepStateSet(int)
	SkinTemperatureSet(int)
	MovementSet(int)
	AlertPausedStatusSet(int)
	ChargingSet(int)
	MovementBucketSet(int)
	WellnessAlertSet(int)
	MonitoringStartTimeSet(int)
	BaseBatteryStatusSet(int)
	BaseStationOnSet(int)
}

type metrics struct {
	ox   prometheus.Gauge
	hr   prometheus.Gauge
	bat  prometheus.Gauge
	btt  prometheus.Gauge
	rsi  prometheus.Gauge
	oxta prometheus.Gauge
	sc   prometheus.Gauge
	ss   prometheus.Gauge
	st   prometheus.Gauge
	mv   prometheus.Gauge
	aps  prometheus.Gauge
	chg  prometheus.Gauge
	mvb  prometheus.Gauge
	onm  prometheus.Gauge
	mst  prometheus.Gauge
	bsb  prometheus.Gauge
	bso  prometheus.Gauge
}

func New(reg prometheus.Registerer) Metrics {
	// Prometheus library panics when try to register the same metrics twice
	// for preventing that to happen we made metrics a singleton using once
	once.Do(func() {
		metric = &metrics{
			ox: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "oxygen_saturation_percent",
				Help: "Current Reading Oxygen Saturation.",
			}),
			hr: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "heart_rate_bpm",
				Help: "Current Reading Heart Rate.",
			}),
			bat: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "battery_percent",
				Help: "Sock battery percentage.",
			}),
			btt: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "battery_minutes",
				Help: "Minutes until sock battery runs out.",
			}),
			rsi: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "signal_strength_rssi",
				Help: "Strenght of signal from sock.",
			}),
			oxta: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "oxygen_10_av_percent",
				Help: "Oxygen Saturation Average over 10 something.",
			}),
			sc: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "sock_connected_bool",
				Help: "If sock is connected.",
			}),
			ss: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "sleep_state",
				Help: "Current sleep state.",
			}),
			st: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "skin_temperature_celsius",
				Help: "Current skin temperature.",
			}),
			mv: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "movement_intensity",
				Help: "Intensity of movement/wiggling.",
			}),
			aps: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "alert_paused_status",
				Help: "If alert is paused.",
			}),
			chg: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "charging",
				Help: "If sock is charging.",
			}),
			mvb: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "movement_bucket",
				Help: "Movement bucket.",
			}),
			onm: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "wellness_alert",
				Help: "Wellness alert.",
			}),
			mst: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "monitoring_start_time_unix_seconds",
				Help: "Monitoring start time in UNIX seconds.",
			}),
			bsb: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "base_battery_status",
				Help: "Status of base battery.",
			}),
			bso: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "base_station_on_bool",
				Help: "If base station is on.",
			}),
		}

		reg.MustRegister(metric.ox)
		reg.MustRegister(metric.hr)
		reg.MustRegister(metric.bat)
		reg.MustRegister(metric.btt)
		reg.MustRegister(metric.rsi)
		reg.MustRegister(metric.oxta)
		reg.MustRegister(metric.sc)
		reg.MustRegister(metric.ss)
		reg.MustRegister(metric.st)
		reg.MustRegister(metric.mv)
		reg.MustRegister(metric.aps)
		reg.MustRegister(metric.chg)
		reg.MustRegister(metric.mvb)
		reg.MustRegister(metric.onm)
		reg.MustRegister(metric.mst)
		reg.MustRegister(metric.bsb)
		reg.MustRegister(metric.bso)
	})

	return metric
}

func (m *metrics) OxygenSaturationSet(ox int) {
	m.ox.Set(float64(ox))
}

func (m *metrics) HeartRateSet(hr int) {
	m.hr.Set(float64(hr))
}

func (m *metrics) BatteryPercentageSet(bat int) {
	m.bat.Set(float64(bat))
}

func (m *metrics) BatteryMinutesSet(btt int) {
	m.btt.Set(float64(btt))
}

func (m *metrics) SignalStrengthSet(rsi int) {
	m.rsi.Set(float64(rsi))
}

func (m *metrics) OxygenTenAVSet(oxta int) {
	m.oxta.Set(float64(oxta))
}

func (m *metrics) SockConnectionSet(sc int) {
	m.sc.Set(float64(sc))
}

func (m *metrics) SleepStateSet(ss int) {
	m.ss.Set(float64(ss))
}

func (m *metrics) SkinTemperatureSet(st int) {
	m.st.Set(float64(st))
}

func (m *metrics) MovementSet(mv int) {
	m.mv.Set(float64(mv))
}

func (m *metrics) AlertPausedStatusSet(aps int) {
	m.aps.Set(float64(aps))
}

func (m *metrics) ChargingSet(chg int) {
	m.chg.Set(float64(chg))
}

func (m *metrics) MovementBucketSet(mvb int) {
	m.mvb.Set(float64(mvb))
}

func (m *metrics) WellnessAlertSet(onm int) {
	m.onm.Set(float64(onm))
}

func (m *metrics) MonitoringStartTimeSet(mst int) {
	m.mst.Set(float64(mst))
}

func (m *metrics) BaseBatteryStatusSet(bsb int) {
	m.bsb.Set(float64(bsb))
}

func (m *metrics) BaseStationOnSet(bso int) {
	m.bso.Set(float64(bso))
}
