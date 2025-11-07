package service

type Cache interface{}
type Database interface{}
type Logger interface{}

type Service struct {
	CacheImpl    Cache
	DatabaseImpl Database
	LoggerImpl   Logger
}

func NewService(cache Cache, database Database, logger Logger) *Service {
	return &Service{
		CacheImpl:    cache,
		DatabaseImpl: database,
		LoggerImpl:   logger,
	}
}
