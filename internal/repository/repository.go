package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/linkiog/lo/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveOrder(order *models.Order) error {
	fmt.Printf("Saving order to DB: %+v\n", order)

	parsedDate, err := time.Parse(time.RFC3339, order.DateCreated)
	if err != nil {
		return fmt.Errorf("invalid date_created format (%s): %w", order.DateCreated, err)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.Exec(`
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (order_uid) DO UPDATE SET
			track_number      = EXCLUDED.track_number,
			entry             = EXCLUDED.entry,
			locale            = EXCLUDED.locale,
			internal_signature= EXCLUDED.internal_signature,
			customer_id       = EXCLUDED.customer_id,
			delivery_service  = EXCLUDED.delivery_service,
			shardkey          = EXCLUDED.shardkey,
			sm_id             = EXCLUDED.sm_id,
			date_created      = EXCLUDED.date_created,
			oof_shard         = EXCLUDED.oof_shard
	`,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.ShardKey,
		order.SmID,
		parsedDate,
		order.OofShard,
	)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO delivery (
			order_uid, name, phone, zip, city, address, region, email
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (order_uid) DO UPDATE SET
			name    = EXCLUDED.name,
			phone   = EXCLUDED.phone,
			zip     = EXCLUDED.zip,
			city    = EXCLUDED.city,
			address = EXCLUDED.address,
			region  = EXCLUDED.region,
			email   = EXCLUDED.email
	`,
		order.OrderUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO payment (
			order_uid, transaction, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (order_uid) DO UPDATE SET
			transaction   = EXCLUDED.transaction,
			request_id    = EXCLUDED.request_id,
			currency      = EXCLUDED.currency,
			provider      = EXCLUDED.provider,
			amount        = EXCLUDED.amount,
			payment_dt    = EXCLUDED.payment_dt,
			bank          = EXCLUDED.bank,
			delivery_cost = EXCLUDED.delivery_cost,
			goods_total   = EXCLUDED.goods_total,
			custom_fee    = EXCLUDED.custom_fee
	`,
		order.OrderUID,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDT,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM items WHERE order_uid = $1", order.OrderUID)
	if err != nil {
		return err
	}

	for _, it := range order.Items {
		_, err := tx.Exec(`
			INSERT INTO items (
				order_uid, chrt_id, track_number, price, rid, name,
				sale, size, total_price, nm_id, brand, status
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`,
			order.OrderUID,
			it.ChrtID,
			it.TrackNumber,
			it.Price,
			it.Rid,
			it.Name,
			it.Sale,
			it.Size,
			it.TotalPrice,
			it.NmID,
			it.Brand,
			it.Status,
		)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetOrder(orderUID string) (*models.Order, error) {
	order := &models.Order{}

	row := r.db.QueryRow(`
		SELECT order_uid, track_number, entry, locale, internal_signature,
		       customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders
		WHERE order_uid = $1
	`, orderUID)

	err := row.Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.ShardKey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	row = r.db.QueryRow(`
		SELECT name, phone, zip, city, address, region, email
		FROM delivery
		WHERE order_uid = $1
	`, orderUID)

	err = row.Scan(
		&order.Delivery.Name,
		&order.Delivery.Phone,
		&order.Delivery.Zip,
		&order.Delivery.City,
		&order.Delivery.Address,
		&order.Delivery.Region,
		&order.Delivery.Email,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	row = r.db.QueryRow(`
		SELECT transaction, request_id, currency, provider, amount,
		       payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payment
		WHERE order_uid = $1
	`, orderUID)

	err = row.Scan(
		&order.Payment.Transaction,
		&order.Payment.RequestID,
		&order.Payment.Currency,
		&order.Payment.Provider,
		&order.Payment.Amount,
		&order.Payment.PaymentDT,
		&order.Payment.Bank,
		&order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	rows, err := r.db.Query(`
		SELECT chrt_id, track_number, price, rid, name, sale,
		       size, total_price, nm_id, brand, status
		FROM items
		WHERE order_uid = $1
	`, orderUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var it models.Item
		if err := rows.Scan(
			&it.ChrtID,
			&it.TrackNumber,
			&it.Price,
			&it.Rid,
			&it.Name,
			&it.Sale,
			&it.Size,
			&it.TotalPrice,
			&it.NmID,
			&it.Brand,
			&it.Status,
		); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, it)
	}

	return order, nil
}

func (r *Repository) GetAllOrders() ([]*models.Order, error) {

	rows, err := r.db.Query("SELECT order_uid FROM orders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.Order
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		order, err := r.GetOrder(uid)
		if err != nil {
			return nil, err
		}
		if order != nil {
			result = append(result, order)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
