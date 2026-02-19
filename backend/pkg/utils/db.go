package utils

import "gorm.io/gorm"

// DropUnusedColumns 删除未使用的列
func DropUnusedColumns(db *gorm.DB, dst interface{}) {
    stmt := &gorm.Statement{DB: db}
    stmt.Parse(dst)
    fields := stmt.Schema.Fields
    columns, _ := db.Debug().Migrator().ColumnTypes(dst)

    for i := range columns {
        found := false
        for j := range fields {
            if columns[i].Name() == fields[j].DBName {
                found = true
                break
            }
        }
        if !found {
            db.Migrator().DropColumn(dst, columns[i].Name())
        }
    }
}