package db

type User struct {
	Uid    string `json:"uid"`
	Name   string `json:"name"`
	Rub    int64  `json:"rub"`
	Pen    int    `json:"pen"`
	RubRes int64  `json:"res_rub"`
	PenRes int    `json:"res_pen"`
}

type UserInfo struct {
	Uid  string `json:"uid"`
	Name string `json:"name"`
}

type Credition struct {
	Uid     string `json:"uid"`
	Rubles  int64  `json:"rub"`
	Pennies int    `json:"pen"`
}

type Transfer struct {
	UidFrom string `json:"uid_sender"`
	UidTo   string `json:"uid_reciever"`
	Rubles  int64  `json:"rub"`
	Pennies int    `json:"pen"`
}

type Buyer struct {
	BuyerUid    string   `json:"uid_buyer"`
	ServicesUid []string `json:"uid_services"`
}
