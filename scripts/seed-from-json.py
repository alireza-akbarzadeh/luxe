#!/usr/bin/env python3
"""
Import catalog + menu data from seed.json (local DB export) into PostgreSQL.

seed.json is a concatenation of JSON arrays (one per table), not a single valid JSON document.
Uses DATABASE_URL from the environment (same as make migrate-up).

Usage:
  DATABASE_URL=postgresql://... python3 scripts/seed-from-json.py
  make seed-from-json
"""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys
from collections import defaultdict
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[1]
DEFAULT_SEED = ROOT / "seed.json"

IMPORT_ORDER = [
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
]

SKIP_COLUMNS: dict[str, set[str]] = {
    "products": {"search_vector", "search_document"},
    "categories": {"search_vector", "search_document"},
    "stores": {"search_vector", "search_document"},
}

COLUMN_RENAME: dict[str, dict[str, str]] = {
    "nav_menus": {"order": '"order"'},
}

JSONB_COLUMNS: dict[str, set[str]] = {
    "nav_menus": {"view_all", "columns", "featured", "label_i18n", "badge_i18n"},
    "products": {"colors", "sizes"},
    "stores": {"settings"},
}

TEXT_ARRAY_COLUMNS: dict[str, set[str]] = {
    "products": {"images", "tags", "channels"},
    "product_attributes": {"values"},
}


def parse_chunks(text: str) -> list[list[dict[str, Any]]]:
    chunks: list[list[dict[str, Any]]] = []
    depth = 0
    start: int | None = None
    for index, char in enumerate(text):
        if char == "[":
            if depth == 0:
                start = index
            depth += 1
        elif char == "]":
            depth -= 1
            if depth == 0 and start is not None:
                chunk = text[start : index + 1]
                chunks.append(json.loads(chunk))
                start = None
    return chunks


def detect_table(row: dict[str, Any]) -> str | None:
    keys = set(row.keys())
    if {"from_state_id", "to_state_id", "event", "workflow_id"} <= keys:
        return "workflow_transitions"
    if {"workflow_id", "code", "color", "text_color"} <= keys:
        return "workflow_states"
    if {"entity_type", "key", "is_active"} <= keys and "workflow_id" not in keys:
        return "workflows"
    if {"coupon_id", "discount_amount", "order_id"} <= keys:
        return None
    if {"product_id", "recorded_at", "price"} <= keys and "name" not in keys:
        return None
    if {"group_id", "icon", "label"} <= keys:
        return "menu_items"
    if keys <= {"id", "name", "display_order", "created_at", "updated_at"}:
        return "menu_groups"
    if "type" in keys and "label" in keys and ("columns" in keys or "order" in keys):
        return "nav_menus"
    if {"product_id", "name", "values"} <= keys:
        return "product_attributes"
    if {"sku", "stock", "category_id", "slug"} <= keys:
        return "products"
    if {"follower_count", "banner_url", "is_verified"} <= keys:
        return "stores"
    if {"parent_id", "is_active", "path", "slug"} <= keys:
        return "categories"
    if {"logo_url", "slug", "status", "workflow_state_id"} <= keys and "sku" not in keys:
        return "brands"
    return None


def sql_quote(value: str) -> str:
    return "'" + value.replace("'", "''") + "'"


def to_sql_literal(table: str, column: str, value: Any) -> str:
    if value is None:
        return "NULL"
    if isinstance(value, bool):
        return "TRUE" if value else "FALSE"
    if isinstance(value, (int, float)):
        return str(value)
    if column in JSONB_COLUMNS.get(table, set()):
        return sql_quote(json.dumps(value)) + "::jsonb"
    if column in TEXT_ARRAY_COLUMNS.get(table, set()):
        if not isinstance(value, list):
            return "NULL"
        if not value:
            return "ARRAY[]::text[]"
        escaped = [sql_quote(str(item)) for item in value]
        return "ARRAY[" + ",".join(escaped) + "]::text[]"
    if isinstance(value, (list, dict)):
        return sql_quote(json.dumps(value)) + "::jsonb"
    return sql_quote(str(value))


def should_skip_row(table: str, row: dict[str, Any]) -> bool:
    if table == "stores":
        slug = str(row.get("slug") or "")
        if slug.startswith("integration-store-"):
            return True
    return False


def build_upsert(table: str, row: dict[str, Any]) -> str:
    skip = SKIP_COLUMNS.get(table, set())
    columns = [key for key in row.keys() if key not in skip]
    if not columns:
        return ""

    sql_columns = [COLUMN_RENAME.get(table, {}).get(col, col) for col in columns]
    values = [to_sql_literal(table, col, row[col]) for col in columns]
    assignments = [
        f"{COLUMN_RENAME.get(table, {}).get(col, col)} = EXCLUDED.{COLUMN_RENAME.get(table, {}).get(col, col)}"
        for col in columns
        if col != "id"
    ]

    return (
        f"INSERT INTO {table} ({', '.join(sql_columns)}) "
        f"VALUES ({', '.join(values)}) "
        f"ON CONFLICT (id) DO UPDATE SET {', '.join(assignments)};"
    )


def shipping_provider_seed() -> str:
    return """
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
  updated_at = NOW();
SELECT setval(pg_get_serial_sequence('shipping_providers', 'id'), (SELECT COALESCE(MAX(id), 1) FROM shipping_providers));
"""


def reset_sequences(tables: list[str]) -> str:
    lines = []
    for table in tables:
        lines.append(
            f"SELECT setval(pg_get_serial_sequence('{table}', 'id'), "
            f"(SELECT COALESCE(MAX(id), 1) FROM {table}));"
        )
    return "\n".join(lines)


def main() -> int:
    seed_path = Path(os.environ.get("SEED_JSON", DEFAULT_SEED))
    database_url = os.environ.get("DATABASE_URL", "").strip()
    if not database_url:
        print("ERROR: DATABASE_URL is not set. Use .env or export DATABASE_URL.", file=sys.stderr)
        return 1

    if not seed_path.is_file() or seed_path.stat().st_size == 0:
        print(
            f"ERROR: {seed_path} is missing or empty.\n"
            "Save your exported JSON to seed.json in the repo root, then re-run.",
            file=sys.stderr,
        )
        return 1

    text = seed_path.read_text(encoding="utf-8")
    chunks = parse_chunks(text)
    if not chunks:
        print(f"ERROR: No JSON arrays found in {seed_path}", file=sys.stderr)
        return 1

    grouped: dict[str, list[dict[str, Any]]] = defaultdict(list)
    unknown = 0
    for chunk in chunks:
        for row in chunk:
            table = detect_table(row)
            if table is None:
                unknown += 1
                continue
            if should_skip_row(table, row):
                continue
            grouped[table].append(row)

    statements = ["BEGIN;"]
    imported: list[str] = []

    for table in IMPORT_ORDER:
        rows = grouped.get(table, [])
        if not rows:
            continue
        imported.append(table)
        print(f"  {table}: {len(rows)} rows")
        for row in rows:
            sql = build_upsert(table, row)
            if sql:
                statements.append(sql)

    statements.append(shipping_provider_seed())
    statements.append(reset_sequences(imported + ["shipping_providers"]))
    statements.append("COMMIT;")

    sql_body = "\n".join(statements)
    print(f"Importing into remote database ({len(imported)} tables, skipped {unknown} unrelated rows)...")

    result = subprocess.run(
        ["psql", database_url, "-v", "ON_ERROR_STOP=1"],
        input=sql_body,
        text=True,
        capture_output=True,
    )
    if result.returncode != 0:
        print(result.stdout)
        print(result.stderr, file=sys.stderr)
        return result.returncode

    print("Seed import complete.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
