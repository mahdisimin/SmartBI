package synops

import "intelligentBI/pkg"

type repo interface {
	Persistdata(any) error
}

type extrnal interface {
	Request() (any, error)
	HealthCheck() int8
}

type FetchDataService struct {
	Repo     repo
	Extrnal  extrnal
	Producct pkg.ProductList
	Data     any
}

func (f *FetchDataService) FetchData() error {
	dataTemp, err := f.Extrnal.Request()
	if err != nil {
		return err
	} else {
		f.Data = dataTemp
	}
	if err := f.Repo.Persistdata(f.Data); err != nil {
		return err
	}

	return nil
}

func (f *FetchDataService) healthCheck() int8 {
	return 0
}
