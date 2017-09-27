package stats

type Service interface {
	PublishMetric(metric string, status int, time float64) error
}

func Open(location, service string) (Service, error) {
	return OpenStatsdService(location, service)
}
