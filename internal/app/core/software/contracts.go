package software

type Source interface {
	Load() ([]*Software, error)
	Name() string // return source name
}

type Calculator interface {
	CalculateObsolescenceScore(software *Software) error
}
