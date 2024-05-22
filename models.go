package rove

type Machine struct {
	Id      int64  `@:"id" @primary:"true"`
	Address string `@:"address" @length:"255"`
	KeyPath string `@:"key_path" @length:"1024"`
	Name    string `@:"name" @length:"255" @unique:"true"`
	Port    int64  `@:"port"`
	User    string `@:"user" @length:"255"`
}
