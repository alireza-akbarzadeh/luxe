// Package seedjson imports catalog and menu rows from a local DB JSON export (seed.json).
package seedjson

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

var importOrder = []string{
	"workflows",
	"workflow_states",
	"workflow_transitions",
	"brands",
	"categories",
	"stores",
	"products",
	"product_attributes",
	"nav_menus",
	"menu_groups",
	"menu_items",
}

var skipColumns = map[string]map[string]struct{}{
	"products":   {"search_vector": {}, "search_document": {}},
	"categories": {"search_vector": {}, "search_document": {}},
	"stores":     {"search_vector": {}, "search_document": {}},
}

var columnRename = map[string]map[string]string{
	"nav_menus": {"order": `"order"`},
}

var jsonbColumns = map[string]map[string]struct{}{
	"nav_menus":  {"view_all": {}, "columns": {}, "featured": {}, "label_i18n": {}, "badge_i18n": {}},
	"products":   {"colors": {}, "sizes": {}},
	"stores":     {"settings": {}},
}

var textArrayColumns = map[string]map[string]struct{}{
	"products":           {"images": {}, "tags": {}, "channels": {}},
	"product_attributes": {"values": {}},
}

// Import reads seedPath (concatenated JSON arrays) and upserts rows into PostgreSQL.
func Import(ctx context.Context, databaseURL, seedPath string) error {
	info, err := os.Stat(seedPath)
	if err != nil || info.Size() == 0 {
		return fmt.Errorf("%s is missing or empty — save your DB export first", seedPath)
	}

	text, err := os.ReadFile(seedPath)
	if err != nil {
		return err
	}

	chunks, err := parseChunks(string(text))
	if err != nil {
		return err
	}
	if len(chunks) == 0 {
		return fmt.Errorf("no JSON arrays found in %s", seedPath)
	}

	grouped := make(map[string][]map[string]any)
	skipped := 0
	for _, chunk := range chunks {
		for _, row := range chunk {
			table := detectTable(row)
			if table == "" {
				skipped++
				continue
			}
			if shouldSkipRow(table, row) {
				continue
			}
			grouped[table] = append(grouped[table], row)
		}
	}

	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	fmt.Printf("Importing into database (%d table types in file, skipped %d unrelated rows)...\n", len(grouped), skipped)

	imported := make([]string, 0, len(importOrder))
	for _, table := range importOrder {
		rows := grouped[table]
		if len(rows) == 0 {
			continue
		}
		imported = append(imported, table)
		fmt.Printf("  %s: %d rows\n", table, len(rows))
		for _, row := range rows {
			sql, args := buildUpsert(table, row)
			if sql == "" {
				continue
			}
			if _, err := tx.Exec(ctx, sql, args...); err != nil {
				return fmt.Errorf("%s id=%v: %w", table, row["id"], err)
			}
		}
	}

	if _, err := tx.Exec(ctx, shippingProvidersSQL); err != nil {
		return fmt.Errorf("shipping_providers: %w", err)
	}

	for _, table := range append(imported, "shipping_providers") {
		if _, err := tx.Exec(ctx, resetSequenceSQL(table)); err != nil {
			return fmt.Errorf("reset sequence %s: %w", table, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	fmt.Println("Seed import complete.")
	return nil
}

func parseChunks(text string) ([][]map[string]any, error) {
	var chunks [][]map[string]any
	depth := 0
	start := -1
	for i, ch := range text {
		switch ch {
		case '[':
			if depth == 0 {
				start = i
			}
			depth++
		case ']':
			depth--
			if depth == 0 && start >= 0 {
				var rows []map[string]any
				if err := json.Unmarshal([]byte(text[start:i+1]), &rows); err != nil {
					return nil, fmt.Errorf("parse JSON array at offset %d: %w", start, err)
				}
				chunks = append(chunks, rows)
				start = -1
			}
		}
	}
	return chunks, nil
}

func rowKeys(row map[string]any) map[string]struct{} {
	keys := make(map[string]struct{}, len(row))
	for k := range row {
		keys[k] = struct{}{}
	}
	return keys
}

func hasKeys(keys map[string]struct{}, required ...string) bool {
	for _, k := range required {
		if _, ok := keys[k]; !ok {
			return false
		}
	}
	return true
}

func detectTable(row map[string]any) string {
	keys := rowKeys(row)
	if hasKeys(keys, "from_state_id", "to_state_id", "event", "workflow_id") {
		return "workflow_transitions"
	}
	if hasKeys(keys, "workflow_id", "code", "color", "text_color") {
		return "workflow_states"
	}
	if hasKeys(keys, "entity_type", "key", "is_active") {
		if _, ok := keys["workflow_id"]; !ok {
			return "workflows"
		}
	}
	if hasKeys(keys, "coupon_id", "discount_amount", "order_id") {
		return ""
	}
	if hasKeys(keys, "product_id", "recorded_at", "price") {
		if _, ok := keys["name"]; !ok {
			return ""
		}
	}
	if hasKeys(keys, "group_id", "icon", "label") {
		return "menu_items"
	}
	if len(keys) == 5 {
		allowed := map[string]struct{}{
			"id": {}, "name": {}, "display_order": {}, "created_at": {}, "updated_at": {},
		}
		match := true
		for k := range keys {
			if _, ok := allowed[k]; !ok {
				match = false
				break
			}
		}
		if match {
			return "menu_groups"
		}
	}
	if _, ok := keys["type"]; ok {
		if _, ok := keys["label"]; ok {
			if _, ok := keys["columns"]; ok {
				return "nav_menus"
			}
			if _, ok := keys["order"]; ok {
				return "nav_menus"
			}
		}
	}
	if hasKeys(keys, "product_id", "name", "values") {
		return "product_attributes"
	}
	if hasKeys(keys, "sku", "stock", "category_id", "slug") {
		return "products"
	}
	if hasKeys(keys, "follower_count", "banner_url", "is_verified") {
		return "stores"
	}
	if hasKeys(keys, "parent_id", "is_active", "path", "slug") {
		return "categories"
	}
	if hasKeys(keys, "logo_url", "slug", "status", "workflow_state_id") {
		if _, ok := keys["sku"]; !ok {
			return "brands"
		}
	}
	return ""
}

func shouldSkipRow(table string, row map[string]any) bool {
	if table != "stores" {
		return false
	}
	slug, _ := row["slug"].(string)
	return strings.HasPrefix(slug, "integration-store-")
}

func sqlColumn(table, column string) string {
	if renamed, ok := columnRename[table][column]; ok {
		return renamed
	}
	return column
}

func buildUpsert(table string, row map[string]any) (string, []any) {
	skip := skipColumns[table]
	columns := make([]string, 0, len(row))
	for col := range row {
		if skip != nil {
			if _, omit := skip[col]; omit {
				continue
			}
		}
		columns = append(columns, col)
	}
	sort.Strings(columns)
	if len(columns) == 0 {
		return "", nil
	}

	sqlCols := make([]string, len(columns))
	placeholders := make([]string, len(columns))
	args := make([]any, len(columns))
	for i, col := range columns {
		sqlCols[i] = sqlColumn(table, col)
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = normalizeValue(table, col, row[col])
	}

	assignments := make([]string, 0, len(columns)-1)
	for _, col := range columns {
		if col == "id" {
			continue
		}
		name := sqlColumn(table, col)
		assignments = append(assignments, fmt.Sprintf("%s = EXCLUDED.%s", name, name))
	}

	sql := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (id) DO UPDATE SET %s",
		table,
		strings.Join(sqlCols, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(assignments, ", "),
	)
	return sql, args
}

func normalizeValue(table, column string, value any) any {
	if value == nil {
		return nil
	}
	if _, ok := jsonbColumns[table][column]; ok {
		return marshalJSON(value)
	}
	if _, ok := textArrayColumns[table][column]; ok {
		return toStringSlice(value)
	}
	switch v := value.(type) {
	case []any, map[string]any:
		return marshalJSON(v)
	default:
		return value
	}
}

func marshalJSON(value any) []byte {
	b, err := json.Marshal(value)
	if err != nil {
		return []byte("null")
	}
	return b
}

func toStringSlice(value any) []string {
	switch v := value.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, fmt.Sprint(item))
		}
		return out
	case []string:
		return v
	default:
		return nil
	}
}

const shippingProvidersSQL = `
INSERT INTO shipping_providers (id, name, description, price, is_active, created_at, updated_at)
VALUES
  (1, 'Standard Shipping', 'Delivery in 5–7 business days', 9.99, TRUE, NOW(), NOW()),
  (2, 'Express Shipping', 'Delivery in 2–3 business days', 19.99, TRUE, NOW(), NOW()),
  (3, 'Free Shipping', 'Complimentary shipping on eligible orders', 0.00, TRUE, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  price = EXCLUDED.price,
  is_active = EXCLUDED.is_active,
  updated_at = NOW()`

func resetSequenceSQL(table string) string {
	return fmt.Sprintf(
		`SELECT setval(pg_get_serial_sequence('%s', 'id'), (SELECT COALESCE(MAX(id), 1) FROM %s), true)`,
		table, table,
	)
}
