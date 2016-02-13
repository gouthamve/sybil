package sybil_test

import sybil "../"
import "os"
import "fmt"

var TEST_TABLE_NAME = "__TEST0__"

func unload_test_table() {
	delete(sybil.LOADED_TABLES, TEST_TABLE_NAME)
}

func delete_test_db() {
	os.RemoveAll(fmt.Sprintf("db/%s", TEST_TABLE_NAME))
	unload_test_table()
}
