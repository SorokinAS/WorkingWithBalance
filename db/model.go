package db

type User struct {
	Uuid   string `json:"uuid"`
	Name   string `json:"name"`
	Rub    int64  `json:"rubles"`
	Pen    int    `json:"pennies"`
	RubRes int64  `json:"resRubles"`
	PenRes int    `json:"resPennies"`
}

type UserInfo struct {
	Uuid string `json:"uuid"`
	Name string `json:"name"`
}

type Credition struct {
	Uuid    string `json:"uuid"`
	Rubles  int64  `json:"rub"`
	Pennies int    `json:"pen"`
}

type Transfer struct {
	UuidFrom string `json:"uuid_sender"`
	UuidTo   string `json:"uuid_reciever"`
	Rubles   int64  `json:"rub"`
	Pennies  int    `json:"pen"`
}
