package model

func (n Nothing) GetType() string {
	return "Nothing"
}

type Nothing struct {
	_ struct{} `plutusType:"DefList" plutusConstr:"1"`
}
