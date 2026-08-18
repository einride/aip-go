package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ysomad/flop"
	"github.com/ysomad/flop/expr"
	"github.com/ysomad/flop/sqflop"
)

type payment struct {
	ID         string    `json:"id"`
	Amount     int64     `json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
	CapturedAt time.Time `json:"captured_at"`
	Provider   string    `json:"provider"`
	UserID     string    `json:"user_id"`
}

type listRequest struct {
	PageSize int32  `json:"page_size"`
	OrderBy  string `json:"order_by"`
	Filter   string `json:"filter"`
}

type cursorRequest struct {
	listRequest
	Cursor string `json:"cursor"`
}

type offsetRequest struct {
	listRequest
	Page int32 `json:"page"`
}

// OrderBy in both responses is the order the paginator resolved, including the
// key columns it appends, not the order the client asked for.
type cursorResponse struct {
	Items      []payment `json:"items"`
	OrderBy    string    `json:"order_by"`
	NextCursor string    `json:"next_cursor,omitempty"`
}

type offsetResponse struct {
	Items      []payment `json:"items"`
	OrderBy    string    `json:"order_by"`
	Page       int32     `json:"page"`
	TotalPages int64     `json:"total_pages"`
	TotalItems int64     `json:"total_items"`
}

var (
	psql           = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	paymentFilters = mustFilterSchema()
	filterCompiler = mustFilterCompiler()
	paginator      = mustPaginator()
)

func mustFilterSchema() *flop.FilterSchema {
	schema, err := flop.NewFilterSchema([]flop.FilterField{
		flop.NewFilterField("id", flop.TypeString, true),
		flop.NewFilterField("amount", flop.TypeInt, false),
		flop.NewFilterField("created_at", flop.TypeTime, false),
		flop.NewFilterField("captured_at", flop.TypeTime, false),
		flop.NewFilterField("provider", flop.TypeString, true),
		flop.NewFilterField("user_id", flop.TypeString, true),
	}, flop.FilterLimits{MaxBytes: 4096, MaxDepth: 16, MaxNodes: 128})
	if err != nil {
		panic(err)
	}
	return schema
}

func mustPaginator() *flop.Paginator[payment] {
	pag, err := flop.NewPaginator(
		[]flop.Column[payment]{
			flop.Key("id", func(p payment) expr.Value { return expr.StringValue(p.ID) }),
			flop.Sortable("amount", func(p payment) expr.Value { return expr.Int64Value(p.Amount) }),
			flop.Sortable("provider", func(p payment) expr.Value { return expr.StringValue(p.Provider) }),
			flop.Sortable("user_id", func(p payment) expr.Value { return expr.StringValue(p.UserID) }),
			flop.Key("captured_at", func(p payment) expr.Value { return expr.TimeValue(p.CapturedAt) }),
		},
		paymentFilters,
		[]flop.OrderField{flop.Desc("captured_at")},
		25,
		100,
	)
	if err != nil {
		panic(err)
	}
	return pag
}

// The demo selects from one table, so every filter field names its own column
// and needs no mapping. A joined query maps only the ambiguous fields:
//
//	func mustFilterCompiler() *sqflop.FilterCompiler {
//		c, err := sqflop.NewFilterCompiler(paymentFilters, map[string]string{
//			"amount":      "p.amount",
//			"captured_at": "p.captured_at",
//			"user_id":     "u.id",
//		})
//		if err != nil {
//			panic(err)
//		}
//		return c
//	}
func mustFilterCompiler() *sqflop.FilterCompiler {
	c, err := sqflop.NewFilterCompiler(paymentFilters, nil)
	if err != nil {
		panic(err)
	}
	return c
}

var db *pgxpool.Pool

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var err error

	db, err = pgxpool.New(ctx, env("DATABASE_URL", "postgres://shop@localhost:5432/shop?sslmode=disable"))
	if err != nil {
		slog.Error("configure database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	migrationCtx, cancelMigration := context.WithTimeout(ctx, 2*time.Minute)
	if err := migrate(migrationCtx, db); err != nil {
		cancelMigration()
		slog.Error("migrate database", "error", err)
		os.Exit(1)
	}
	cancelMigration()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /cursor-payments", handleCursorPayments)
	mux.HandleFunc("POST /offset-payments", handleOffsetPayments)

	server := &http.Server{
		Addr:    env("ADDR", ":8080"),
		Handler: mux,
	}

	slog.Info("server start", "address", server.Addr)

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.ListenAndServe() }()

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("serve", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		slog.Info("server stop")

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown", "error", err)
			os.Exit(1)
		}
	}
}

func handleCursorPayments(w http.ResponseWriter, r *http.Request) {
	var req cursorRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	orderBy, err := flop.ParseOrderBy(req.OrderBy)
	if err != nil {
		writeListError(w, err)
		return
	}

	f, err := flop.ParseFilter(req.Filter)
	if err != nil {
		writeListError(w, err)
		return
	}

	params, err := paginator.CursorParams(req.Cursor, req.PageSize, orderBy, f)
	if err != nil {
		writeListError(w, err)
		return
	}

	payments, err := listCursorPayments(r.Context(), params)
	if err != nil {
		writeListError(w, err)
		return
	}

	page, err := paginator.CursorPage(payments, params)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, cursorResponse{
		Items:      page.Items,
		OrderBy:    params.OrderBy().String(),
		NextCursor: page.NextCursor,
	})
}

func handleOffsetPayments(w http.ResponseWriter, r *http.Request) {
	var req offsetRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	orderBy, err := flop.ParseOrderBy(req.OrderBy)
	if err != nil {
		writeListError(w, err)
		return
	}

	f, err := flop.ParseFilter(req.Filter)
	if err != nil {
		writeListError(w, err)
		return
	}

	params, err := paginator.OffsetParams(req.Page, req.PageSize, orderBy, f)
	if err != nil {
		writeListError(w, err)
		return
	}

	items, total, err := listOffsetPayments(r.Context(), params)
	if err != nil {
		writeListError(w, err)
		return
	}

	page, err := flop.NewOffsetPage(items, params, total)
	if err != nil {
		writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, offsetResponse{
		Items:      page.Items,
		OrderBy:    params.OrderBy().String(),
		Page:       page.Number,
		TotalPages: page.TotalPages,
		TotalItems: page.TotalItems,
	})
}

func baseQuery() sq.SelectBuilder {
	return psql.
		Select("id", "amount", "created_at", "captured_at", "provider", "user_id").
		From("payments")
}

func listCursorPayments(ctx context.Context, params flop.CursorParams) ([]payment, error) {
	base := filterCompiler.Where(baseQuery(), params.Filter())

	sql, args, err := sqflop.CursorQuery(base, params).ToSql()
	if err != nil {
		return nil, fmt.Errorf("cursor: %w", err)
	}

	slog.Info("[cursor]", "sql", sql, "args", args)

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return scanPayments(rows)
}

func listOffsetPayments(ctx context.Context, params flop.OffsetParams) ([]payment, int64, error) {
	// Filter once before Page and Count so rows and totals share predicates.
	base := filterCompiler.Where(baseQuery(), params.Filter())

	sql, args, err := sqflop.OffsetQuery(base, params).ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("offset: %w", err)
	}

	slog.Info("[offset] list payments", "sql", sql, "args", args)

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}

	payments, err := scanPayments(rows)
	if err != nil {
		return nil, 0, err
	}

	sql, args, err = sqflop.CountQuery(base).ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("offset count: %w", err)
	}

	slog.Info("[offset] count", "sql", sql, "args", args)

	var total int64

	if err := db.QueryRow(ctx, sql, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

func scanPayments(rows pgx.Rows) ([]payment, error) {
	defer rows.Close()
	pp := make([]payment, 0)
	for rows.Next() {
		var p payment
		if err := rows.Scan(&p.ID, &p.Amount, &p.CreatedAt, &p.CapturedAt, &p.Provider, &p.UserID); err != nil {
			return nil, err
		}
		pp = append(pp, p)
	}
	return pp, rows.Err()
}

// writeListError maps a pagination failure onto a status.
//
// flop names what it found wrong and takes no view on what that means, so
// deciding which causes are the client's fault is the endpoint's job. Every
// sentinel below reports values that reached flop unusable; anything else is a
// misdeclared endpoint or a broken query, which the client cannot fix.
func writeListError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, flop.ErrInvalidFilter),
		errors.Is(err, flop.ErrInvalidOrder),
		errors.Is(err, flop.ErrInvalidCursor),
		errors.Is(err, flop.ErrCursorMismatch),
		errors.Is(err, flop.ErrInvalidPageSize),
		errors.Is(err, flop.ErrInvalidPage):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeInternalError(w, err)
	}
}

func writeInternalError(w http.ResponseWriter, err error) {
	slog.Error("request failed", "error", err)
	writeError(w, http.StatusInternalServerError, errors.New("internal server error"))
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
