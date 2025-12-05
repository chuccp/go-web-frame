package util

import "testing"

func TestCreateMysqlConnection(t *testing.T) {

	db, err := CreateMysqlConnection("root", "Cooge_123", "124.220.164.58", 3306, "t_anti_lost_qrcode", "utf8")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(db)
	}

}
