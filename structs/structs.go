package structs

type QtInfo struct {
	Code   string `json:"f12"`
	Name   string `json:"f14"`
	Sector string `json:"f100"`
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
	ID     int    `db:"id"`
	Data   string `db:"data"`
	Name   string `db:"name"`
	Code   string `db:"code"`
	Sector string `db:"sector"`
}
