-- Phase 1: users and shares. See docs/plan-from-scratch.md §5.

CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS shares (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  token TEXT UNIQUE NOT NULL,
  path TEXT NOT NULL,
  password_hash TEXT,
  expires_at DATETIME,
  name TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
