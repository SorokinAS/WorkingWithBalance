package db

import (
	"context"
	"errors"
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
	row := db.Pool.QueryRow(context.Background(), "SELECT name, rub, pen, rub_res, pen_res FROM users WHERE uid=$1", uid)
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

func (db *DataBaseConnection) BuyService(buyer *Buyer) error {
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return err
	}
	err = db.buyService(tx, buyer)
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
	_, err := tx.Exec(context.Background(), "UPDATE users SET rub=rub+$1, pen=pen+$2 WHERE uid=$3", cash.Rubles, cash.Pennies, cash.Uid)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), "INSERT INTO operations VALUES ($1, $2, $3, $4, $5, $6)",
		time.Now().Format(time.DateTime), cash.Uid, cash.Uid, "add", cash.Rubles, cash.Pennies)
	if err != nil {
		return err
	}
	return nil
}

func (db *DataBaseConnection) transferMoney(tx pgx.Tx, transfer *Transfer) error {
	var avialable bool
	_, err := tx.Exec(context.Background(), "UPDATE users SET pen=pen+100, rub=rub-1 WHERE uid=$1 AND pen<$2 AND rub>=1", transfer.UidFrom, transfer.Pennies)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), "UPDATE users SET rub=rub-$1, pen=pen-$2 WHERE uid=$3", transfer.Rubles, transfer.Pennies, transfer.UidFrom)
	if err != nil {
		return err
	}
	row := tx.QueryRow(context.Background(), "SELECT (rub<0 OR pen<0) FROM users WHERE uid=$1", transfer.UidFrom)
	if err = row.Scan(&avialable); err != nil {
		return err
	}
	if avialable {
		return errors.New("error: there aren't enough money to make the transaction")
	}
	_, err = tx.Exec(context.Background(), "UPDATE users SET rub=rub+$1, pen=pen+$2 WHERE uid=$3", transfer.Rubles, transfer.Pennies, transfer.UidTo)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), "INSERT INTO operations VALUES ($1, $2, $3, $4, $5, $6)",
		time.Now().Format(time.DateTime), transfer.UidFrom, transfer.UidTo, "transfer", transfer.Rubles, transfer.Pennies)
	if err != nil {
		return err
	}
	return nil
}

func (db *DataBaseConnection) buyService(tx pgx.Tx, buyer *Buyer) error {
	var rub int64
	var pen int

	row := db.Pool.QueryRow(context.Background(), "SELECT SUM(rub), SUM(pen) FROM services WHERE service_uid IN ($1)", buyer.ServicesUid)
	if err := row.Scan(&rub, &pen); err != nil {
		return err
	}
	_, err := tx.Exec(context.Background(), "UPDATE users SET pen=pen+100, rub=rub-1 WHERE uid=$1 AND pen<$2 AND rub>=1", buyer.BuyerUid, pen)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), "UPDATE users SET rub=rub-$1, pen=pen-$2, rub_res=rub_res+$1, pen_res=pen_res+$2 WHERE uid=$3 AND rub>=$1 AND pen>=$2",
		rub, pen, buyer.BuyerUid)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), "UPDATE users SET rub_res=rub_res-$1, pen_res=pen_res-$2 WHERE uid=$3 AND rub>=$1 AND pen>=$2",
		rub, pen, buyer.BuyerUid)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), "UPDATE organizations SET rub=rub+$1, pen=pen+$2 WHERE org_uid=(SELECT org_uid FROM services WHERE service_uid IN ($3) GROUP BY org_uid)",
		rub, pen, buyer.ServicesUid)
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(),
		"INSERT INTO operations VALUES ($1, $2, $3, $4, $5, $6), ($1, $2, (SELECT org_uid FROM services WHERE service_uid IN ($7) GROUP BY org_uid), $8, $5, $6)",
		time.Now().Format(time.DateTime), buyer.BuyerUid, buyer.BuyerUid, "reserve", rub, pen, buyer.ServicesUid, "buy")
	if err != nil {
		return err
	}
	return nil
}
