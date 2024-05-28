package rove

type Machine struct {
	Id      int64  `@:"id" @primary:"true" json:"-"`
	Address string `@:"address" @length:"255" json:"address"`
	KeyPath string `@:"key_path" @length:"1024" json:"-"`
	Name    string `@:"name" @length:"255" @unique:"true" json:"name"`
	Port    int64  `@:"port" json:"-"`
	User    string `@:"user" @length:"255" json:"-"`
}
