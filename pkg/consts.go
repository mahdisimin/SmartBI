package pkg

type ProductList uint8

const (
	SynOps ProductList = iota + 1
)

type SynOpsAPIList uint8

const (
	LoginHistory SynOpsAPIList = iota + 1
	UserData
)

type SynOpsAPIURL string

const (
	Req_LoginHostory SynOpsAPIURL = "https://login.synops.io"
	Req_UserData     SynOpsAPIURL = "https://userdata.synops.io"
)

const (
	UserName = "sa"
	Password = "123456"
	URL      = "."
	Database = "SMARTBI"
)
