package models

import "github.com/speps/go-hashids"

func generateHash(ID int) string {
	hd := hashids.NewData()
	hd.Salt = "xOBtdmJZxRcz^jkkyHfkrkT1*02bJUn+YQts0*xCeka%cGHCN1fjaC*faFtY"
	hd.MinLength = 8
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode([]int{ID})
	return e
}
