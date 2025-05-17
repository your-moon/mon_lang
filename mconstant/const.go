package mconstant

var IntZero = Int32{
	Value: 0,
}

var IntOne = Int32{
	Value: 1,
}

type Const interface {
	constant()
	GetValue() int64
}

type Int64 struct {
	Value int64
}

func (i Int64) constant()       {}
func (i Int64) GetValue() int64 { return i.Value }

type Int32 struct {
	Value int32
}

func (i Int32) constant()       {}
func (i Int32) GetValue() int64 { return int64(i.Value) }
