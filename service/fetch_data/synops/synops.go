package synops

import "intelligentBI/pkg"

type repo interface {
	Persistdata(any) error
}
type extrnal interface {
	Request() (string, error)
	HealthCheck() int8
}

type FetchDataService struct {
	Repo     repo
	Extrnal  extrnal
	Producct pkg.ProductList
	Data     string
}

func (f FetchDataService) FetchData() error {
	dataTemp, err := f.Extrnal.Request()
	if err != nil {
		return err
	} else {
		f.Data = dataTemp
	}
	return nil
}

func (f FetchDataService) healthCheck() int8 {
	return 0
}

func (f FetchDataService) PersistData(data any) error {

	err := f.Repo.Persistdata(data)
	if err != nil {
		return err
	}
	return nil
}
