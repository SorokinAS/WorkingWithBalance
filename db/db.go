package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DataBaseConnection struct {
	Pool *pgxpool.Pool
}

func NewDbConnection() *DataBaseConnection {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("Failed connection %v", err)
	}
	return &DataBaseConnection{pool}
}

func (db *DataBaseConnection) GetUsers() ([]UserInfo, error) {
	users := make([]UserInfo, 0)
	rows, err := db.Pool.Query(context.Background(), "SELECT uid, name FROM users")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var man UserInfo
		if err := rows.Scan(&man.Uid, &man.Name); err != nil {
			return nil, err
		}
		users = append(users, man)
	}
	return users, nil
}

func (db *DataBaseConnection) GetUserById(uid string) (User, error) {
	var user User
	row := db.Pool.QueryRow(context.Background(), "SELECT name, rubles, pennies, rubles_res, pennies_res FROM users WHERE uid=$1", uid)
	if err := row.Scan(&user.Name, &user.Rub, &user.Pen, &user.RubRes, &user.PenRes); err != nil {
		return User{}, err
	}
	user.Uid = uid
	return user, nil
}

func (db *DataBaseConnection) CreateUser(user *User) (UserInfo, error) {
	user.Uid = uuid.New().String()
	_, err := db.Pool.Exec(context.Background(), "INSERT INTO users VALUES ($1, $2, $3, $4)", user.Uid, user.Name, user.Rub, user.Pen)
	if err != nil {
		return UserInfo{}, err
	}
	return UserInfo{
		Uid:  user.Uid,
		Name: user.Name,
	}, nil
}

func (db *DataBaseConnection) AddMoney(cash *Credition) error {
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return err
	}
	err = db.addMoney(tx, cash)
	if err != nil {
		tx.Rollback(context.Background())
		if err != nil {
			return err
		}
	} else {
		tx.Commit(context.Background())
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DataBaseConnection) ReserveMoneyFromBalance(cash *Credition) error {
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return err
	}
	err = db.reserveMoneyFromBalance(tx, cash)
	if err != nil {
		tx.Rollback(context.Background())
		if err != nil {
			return err
		}
	} else {
		tx.Commit(context.Background())
		if err != nil {
			return err
		}
	}
	return err
}

func (db *DataBaseConnection) TransferMoney(transfer *Transfer) error {
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return err
	}
	err = db.transferMoney(tx, transfer)
	if err != nil {
		tx.Rollback(context.Background())
		if err != nil {
			return err
		}
	} else {
		tx.Commit(context.Background())
		if err != nil {
			return err
		}
	}
	return err
}

func (db *DataBaseConnection) addMoney(tx pgx.Tx, cash *Credition) error {
	_, err := tx.Exec(context.Background(), "UPDATE users SET rubles=rubles+$1, pennies=pennies+$2 WHERE uid=$3", cash.Rubles, cash.Pennies, cash.Uid)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), "INSERT INTO operations VALUES ($1, $2, $3, $4)",
		time.Now().Format(time.DateTime), cash.Uid, cash.Uid, fmt.Sprintf("add %d rub %d pen", cash.Rubles, cash.Pennies))
	if err != nil {
		return err
	}
	return nil
}

func (db *DataBaseConnection) reserveMoneyFromBalance(tx pgx.Tx, cash *Credition) error {
	var avialable bool
	_, err := tx.Exec(context.Background(), "UPDATE users SET pennies=pennies+100, rubles=rubles-1 WHERE uid=$1 AND pennies<$2 AND rubles>=1", cash.Uid, cash.Pennies)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), "UPDATE users SET rubles=rubles-$1, pennies=pennies-$2, rubles_res=rubles_res+$1, pennies_res=pennies_res+$2 WHERE uid=$3 AND rubles>=$1 AND pennies>=$2",
		cash.Rubles, cash.Pennies, cash.Uid)
	if err != nil {
		return err
	}
	row := tx.QueryRow(context.Background(), "SELECT (rubles<0 OR pennies<0) FROM users WHERE uid=$1", cash.Uid)
	if err = row.Scan(&avialable); err != nil {
		return err
	}
	if avialable {
		return errors.New("error: there aren't enough money to make the transaction")
	}
	_, err = tx.Exec(context.Background(), "INSERT INTO operations VALUES ($1, $2, $3, $4)",
		time.Now().Format(time.DateTime), cash.Uid, cash.Uid, fmt.Sprintf("reserve %d rub %d pen", cash.Rubles, cash.Pennies))
	if err != nil {
		return err
	}
	return nil
}

func (db *DataBaseConnection) transferMoney(tx pgx.Tx, transfer *Transfer) error {
	var avialable bool
	_, err := tx.Exec(context.Background(), "UPDATE users SET pennies=pennies+100, rubles=rubles-1 WHERE uid=$1 AND pennies<$2 AND rubles>=1", transfer.UidFrom, transfer.Pennies)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), "UPDATE users SET rubles=rubles-$1, pennies=pennies-$2 WHERE uid=$3", transfer.Rubles, transfer.Pennies, transfer.UidFrom)
	if err != nil {
		return err
	}
	row := tx.QueryRow(context.Background(), "SELECT (rubles<0 OR pennies<0) FROM users WHERE uid=$1", transfer.UidFrom)
	if err = row.Scan(&avialable); err != nil {
		return err
	}
	if avialable {
		return errors.New("error: there aren't enough money to make the transaction")
	}
	_, err = tx.Exec(context.Background(), "UPDATE users SET rubles=rubles+$1, pennies=pennies+$2 WHERE uid=$3", transfer.Rubles, transfer.Pennies, transfer.UidTo)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), "INSERT INTO operations VALUES ($1, $2, $3, $4)",
		time.Now().Format(time.DateTime), transfer.UidFrom, transfer.UidTo, fmt.Sprintf("transfer %d rub %d pen", transfer.Rubles, transfer.Pennies))
	if err != nil {
		return err
	}
	return nil
}
