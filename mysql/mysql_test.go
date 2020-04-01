// +build debug
// go test -v -tags debug

package mysql

import (
	"testing"
)

func TestInitMySQL(t *testing.T) {
	type args struct {
		dsn string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{{"", args{"root:@tcp(localhost)/test"}, false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := InitMySQL(tt.args.dsn); (err != nil) != tt.wantErr {
				t.Errorf("InitMySQL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
