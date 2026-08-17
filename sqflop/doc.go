// Package sqflop applies validated flop requests to Squirrel select
// builders. It builds cursor seeks, offset pages, matching counts, ordering,
// and filters without executing queries or choosing a database driver.
//
// One predicate is dialect-specific: the has operator on a
// flop.TypeStringList field compiles to Postgres array membership,
// "? = ANY(column)", which MySQL and SQLite do not accept. Endpoints on those
// databases should leave TypeStringList out of their schema.
package sqflop
