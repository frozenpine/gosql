package gosql

type filterAction string

const (
	FilterEqual filterAction = "="
)

type FilterDefine struct {
	columnName string
	action     filterAction
	value      interface{}
}

func (f *FilterDefine) String() string {
	return ""
}

func NewFilter(col string, act filterAction, v interface{}) *FilterDefine {
	return &FilterDefine{col, act, v}
}
