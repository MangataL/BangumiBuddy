package types

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)


type Page struct {
	Num  int
	Size int
}

func (p *Page) Empty() bool {
	return p.Num == 0 && p.Size == 0
}

func (p *Page) Apply(query *gorm.DB) *gorm.DB {
	if p.Empty() {
		return query
	}
	return query.Limit(p.Size).Offset((p.Num - 1) * p.Size)
}

type Order struct {
	Field string
	Way   string
}

func (o *Order) Empty() bool {
	return o.Field == "" && o.Way == ""
}

func (o *Order) Apply(query *gorm.DB) *gorm.DB {
	if o.Empty() {
		return query
	}
	return query.Order(clause.OrderByColumn{
		Column: clause.Column{
			Name: o.Field,
		},
		Desc: o.Way == "desc",
	})
}