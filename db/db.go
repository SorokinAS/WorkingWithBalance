package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DataBaseConnection struct {
	// Dbconn *pgx.Conn
	Pool *pgxpool.Pool
}

func NewDbConnection() *DataBaseConnection {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("Failed connection %v", err)
	}
	return &DataBaseConnection{pool}
}

func (db *DataBaseConnection) GetUsers() ([]byte, error) {
	users := make([]UserInfo, 0)
	rows, err := db.Pool.Query(context.Background(), "SELECT uuid, name FROM users")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var man UserInfo
		if err := rows.Scan(&man.Uuid, &man.Name); err != nil {
			return nil, err
		}
		users = append(users, man)
	}
	res, err := json.Marshal(users)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (db *DataBaseConnection) GetUserById(uid string) ([]byte, error) {
	var user User
	row := db.Pool.QueryRow(context.Background(), "SELECT name, rubles, pennies, rubles_res, pennies_res FROM users WHERE uuid=$1", uid)
	if err := row.Scan(&user.Name, &user.Rub, &user.Pen, &user.RubRes, &user.PenRes); err != nil {
		return nil, err
	}
	user.Uuid = uid
	res, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (db *DataBaseConnection) CreateUser(user *User) ([]byte, error) {
	user.Uuid = uuid.New().String()
	_, err := db.Pool.Exec(context.Background(), "INSERT INTO users VALUES ($1, $2, $3, $4)", user.Uuid, user.Name, user.Rub, user.Pen)
	if err != nil {
		return nil, err
	}
	res, _ := json.Marshal(UserInfo{
		Uuid: user.Uuid,
		Name: user.Name,
	})
	return res, nil
}

func (db *DataBaseConnection) AddMoney(cash *Credition) ([]byte, error) {
	_, err := db.Pool.Exec(context.Background(), "UPDATE users SET rubles=rubles+$1, pennies=pennies+$2 WHERE uuid=$3", cash.Rubles, cash.Pennies, cash.Uuid)
	if err != nil {
		return nil, err
	}
	res, _ := json.Marshal(fmt.Sprintf("balanse %v was credited", *cash))
	return res, nil
}

func (db *DataBaseConnection) ReserveMoneyFromBalance(cash *Credition) ([]byte, error) {
	var res []byte
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer func(cause error) {
		if err != nil {
			tx.Rollback(context.Background())
		} else {
			tx.Commit(context.Background())
		}
	}(err)
	_, err = tx.Exec(context.Background(), "UPDATE users SET pennies=pennies+100, rubles=rubles-1 WHERE uuid=$1 AND pennies<$2 AND rubles>=1", cash.Uuid, cash.Pennies)
	_, err = tx.Exec(context.Background(), "UPDATE users SET rubles=rubles-$1, pennies=pennies-$2 WHERE uuid=$3 AND rubles>=$1 AND pennies>=$2", cash.Rubles, cash.Pennies, cash.Uuid)

	if err == nil {
		res, _ = json.Marshal(fmt.Sprintf("reserve balance %v was credited", *cash))
	}
	return res, err
}
