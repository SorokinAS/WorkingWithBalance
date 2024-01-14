package db

type User struct {
	Uid    string `json:"uid"`
	Name   string `json:"name"`
	Rub    int64  `json:"rubles"`
	Pen    int    `json:"pennies"`
	RubRes int64  `json:"resRubles"`
	PenRes int    `json:"resPennies"`
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
