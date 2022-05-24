package structs

type QtInfo struct {
	Code    string `json:"f12"`
	Name    string `json:"f14"`
	Sector  string `json:"f100"`
	IncRate string `json:"f3"`
}

type QtProcess1 struct {
	Diff []QtInfo `json:"diff"`
}

type QtData struct {
	Data QtProcess1 `json:"data"`
}

type DbTest struct {
	ID   int    `db:"id"`
	Data string `db:"data"`
}

type DbSt struct {
	ID      int    `db:"id"`
	Data    string `db:"data"`
	Name    string `db:"name"`
	Code    string `db:"code"`
	Sector  string `db:"sector"`
	IncRate string `db:"inc_rate"`
}

type StData struct {
	Hq [][]string `json:"hq"`
}

//0=日期，1=开盘，2=收盘，3=涨跌额，4=涨跌幅，5=最低，6=最高，7=成交量，8=成交额，9=换手率
type K struct {
	Date    string  //0
	Open    float64 //1
	Close   float64 //2
	High    float64 //6
	Low     float64 //5
	Vol     float64 //7
	TRate   float64 //9	TurnoverRate
	IncRate float64 //4 increase
}

type Res struct {
	Date    string
	FindNum int
	UpNum   int
	DownNum int
}

type Conf struct {
	Db_username string
	Db_pwd      string
	Db_server   string
	Db_database string
	Db_port     int
	Db_network  string
}

type DaItems struct {
	Total     int
	Low_up    int
	High_up   int
	Close_up  int
	All_well  int
	Sz1_close float64 //大盘收盘1
	Sz2_close float64 //大盘收盘2
}
