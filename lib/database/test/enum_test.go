package test

import (
	"fmt"
	"sabot/lib/database"
)

func Test_enums() {
	var d database.DBType = database.TwoDB
	fmt.Println(d)             // Print : TwoDB
	fmt.Println(d.String())    // Print : TwoDB
	fmt.Println(d.EnumIndex()) // Print : 0

	var q database.QueryType = database.Kw
	fmt.Println(q)             // Print : Kw
	fmt.Println(q.String())    // Print : Kw
	fmt.Println(q.EnumIndex()) // Print : 1

	q = database.Idx
	fmt.Println(q)             // Print : Idx
	fmt.Println(q.String())    // Print : Idx
	fmt.Println(q.EnumIndex()) // Print : 0
}
