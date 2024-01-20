package db

import (
	"context"
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
		return nil, fmt.Errorf("get users failure: %w", err)
	}
	for rows.Next() {
		var man UserInfo
		if err := rows.Scan(&man.Uid, &man.Name); err != nil {
			return nil, fmt.Errorf("scan user from list of users failure: %w", err)
		}
		users = append(users, man)
	}
	return users, nil
}

func (db *DataBaseConnection) GetUserById(uid string) (User, error) {
	var user User
	row := db.Pool.QueryRow(context.Background(), "SELECT name, rub, pen, rub_res, pen_res FROM users WHERE uid=$1", uid)
	if err := row.Scan(&user.Name, &user.Rub, &user.Pen, &user.RubRes, &user.PenRes); err != nil {
		return User{}, fmt.Errorf("get user failure: %w", err)
	}
	user.Uid = uid
	return user, nil
}

func (db *DataBaseConnection) CreateUser(user *User) (UserInfo, error) {
	user.Uid = uuid.New().String()
	_, err := db.Pool.Exec(context.Background(), "INSERT INTO users VALUES ($1, $2, $3, $4)", user.Uid, user.Name, user.Rub, user.Pen)
	if err != nil {
		return UserInfo{}, fmt.Errorf("insert user failure: %w", err)
	}
	return UserInfo{
		Uid:  user.Uid,
		Name: user.Name,
	}, nil
}

func (db *DataBaseConnection) AddMoney(cash *Credition) error {
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("begin transaction failure: %w", err)
	}
	err = db.addMoney(tx, cash)
	if err != nil {
		tx.Rollback(context.Background())
		if err != nil {
			return fmt.Errorf("rollback operation failure: %w", err)
		}
	} else {
		tx.Commit(context.Background())
		if err != nil {
			return fmt.Errorf("commiting operation failure: %w", err)
		}
	}
	return nil
}

func (db *DataBaseConnection) TransferMoney(transfer *Transfer) error {
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("begin transaction failure: %w", err)
	}
	err = db.transferMoney(tx, transfer)
	if err != nil {
		tx.Rollback(context.Background())
		if err != nil {
			return fmt.Errorf("rollback operation failure: %w", err)
		}
	} else {
		tx.Commit(context.Background())
		if err != nil {
			return fmt.Errorf("commiting operation failure: %w", err)
		}
	}
	return err
}

func (db *DataBaseConnection) BuyService(buyer *Buyer) error {
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("begin transaction failure: %w", err)
	}
	err = db.buyService(tx, buyer)
	if err != nil {
		tx.Rollback(context.Background())
		if err != nil {
			return fmt.Errorf("rollback operation failure: %w", err)
		}
	} else {
		tx.Commit(context.Background())
		if err != nil {
			return fmt.Errorf("commiting operation failure: %w", err)
		}
	}
	return err
}

func (db *DataBaseConnection) addMoney(tx pgx.Tx, cash *Credition) error {
	_, err := tx.Exec(context.Background(), "UPDATE users SET rub=rub+$1, pen=pen+$2 WHERE uid=$3", cash.Rubles, cash.Pennies, cash.Uid)
	if err != nil {
		return fmt.Errorf("update adding balance failure: %w", err)
	}
	_, err = tx.Exec(context.Background(), "INSERT INTO operations VALUES ($1, $2, $3, $4, $5, $6)",
		time.Now().Format(time.DateTime), cash.Uid, cash.Uid, "add", cash.Rubles, cash.Pennies)
	if err != nil {
		return fmt.Errorf("insert operations for adding balance failure: %w", err)
	}
	return nil
}

func (db *DataBaseConnection) transferMoney(tx pgx.Tx, transfer *Transfer) error {
	var avialable bool
	_, err := tx.Exec(context.Background(), "UPDATE users SET pen=pen+100, rub=rub-1 WHERE uid=$1 AND pen<$2 AND rub>=1", transfer.UidFrom, transfer.Pennies)
	if err != nil {
		return fmt.Errorf("pen update failure: %w", err)
	}
	_, err = tx.Exec(context.Background(), "UPDATE users SET rub=rub-$1, pen=pen-$2 WHERE uid=$3", transfer.Rubles, transfer.Pennies, transfer.UidFrom)
	if err != nil {
		return fmt.Errorf("withdrawing money from balance failure: %w", err)
	}
	row := tx.QueryRow(context.Background(), "SELECT (rub<0 OR pen<0) FROM users WHERE uid=$1", transfer.UidFrom)
	if err = row.Scan(&avialable); err != nil {
		return fmt.Errorf("select a negative balance failure: %w", err)
	}
	if avialable {
		return fmt.Errorf("there aren't enough money for transfer: %w", err)
	}
	_, err = tx.Exec(context.Background(), "UPDATE users SET rub=rub+$1, pen=pen+$2 WHERE uid=$3", transfer.Rubles, transfer.Pennies, transfer.UidTo)
	if err != nil {
		return fmt.Errorf("update reciever's balance after transfer failure: %w", err)
	}
	_, err = tx.Exec(context.Background(), "INSERT INTO operations VALUES ($1, $2, $3, $4, $5, $6)",
		time.Now().Format(time.DateTime), transfer.UidFrom, transfer.UidTo, "transfer", transfer.Rubles, transfer.Pennies)
	if err != nil {
		return fmt.Errorf("insert operations for transfer failure: %w", err)
	}
	return nil
}

func (db *DataBaseConnection) buyService(tx pgx.Tx, buyer *Buyer) error {
	var rub int64
	var pen int
	row := db.Pool.QueryRow(context.Background(), "SELECT SUM(rub), SUM(pen) FROM services WHERE service_uid = ANY(@uids)", pgx.NamedArgs{"uids": buyer.ServicesUid})
	if err := row.Scan(&rub, &pen); err != nil {
		return fmt.Errorf("sum select failure: %w", err)
	}
	_, err := tx.Exec(context.Background(), "UPDATE users SET pen=pen+100, rub=rub-1 WHERE uid=$1 AND pen<$2 AND rub>=1", buyer.BuyerUid, pen)
	if err != nil {
		return fmt.Errorf("failure update pen in user: %w", err)
	}
	_, err = tx.Exec(context.Background(), "UPDATE users SET rub=rub-$1, pen=pen-$2, rub_res=rub_res+$1, pen_res=pen_res+$2 WHERE uid=$3 AND rub>=$1 AND pen>=$2",
		rub, pen, buyer.BuyerUid)
	if err != nil {
		return fmt.Errorf("reserve money failure: %w", err)
	}
	_, err = tx.Exec(context.Background(), "UPDATE users SET rub_res=rub_res-$1, pen_res=pen_res-$2 WHERE uid=$3 AND rub>=$1 AND pen>=$2",
		rub, pen, buyer.BuyerUid)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	_, err = tx.Exec(context.Background(), "UPDATE organizations SET rub=rub+@rub, pen=pen+@pen WHERE org_uid=(SELECT org_uid FROM services WHERE service_uid = ANY(@uids) GROUP BY org_uid)",
		pgx.NamedArgs{"rub": rub, "pen": pen, "uids": buyer.ServicesUid})
	if err != nil {
		return fmt.Errorf("update organizations failure: %w", err)
	}
	_, err = tx.Exec(context.Background(),
		"INSERT INTO operations VALUES (@time, @buyerUid, @buyerUid, 'reserve', @rub, @pen), (@time, @buyerUid, (SELECT org_uid FROM services WHERE service_uid = ANY(@uids) GROUP BY org_uid), 'buy', @rub, @pen)",
		pgx.NamedArgs{"time": time.Now().Format(time.DateTime), "buyerUid": buyer.BuyerUid, "rub": rub, "pen": pen, "uids": buyer.ServicesUid})
	if err != nil {
		return fmt.Errorf("insert operations for buying failure: %w", err)
	}
	return nil
}
