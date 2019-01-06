package set

import ()

type HashSet struct {
	m map[interface{}]bool
}

func NewHashSet() *HashSet {
	return &HashSet{m: make(map[interface{}]bool)}
}

//添加元素值
func (set *HashSet) Add(e interface{}) bool {
	if !set.m[e] {
		set.m[e] = true
		return true
	}
	return false
}

//删除元素值
func (set *HashSet) Remove(e interface{}) {
	delete(set.m, e)
}

//清除所有元素
func (set *HashSet) Clear() {
	set.m = make(map[interface{}]bool)
}

//判断是否包含某个元素值
func (set *HashSet) Contains(e interface{}) bool {
	return set.m[e]
}

//获取元素值的数量
func (set *HashSet) Len() int {
	return len(set.m)
}

//判断与其他HashSet类型值是否相同
func (set *HashSet) Same(other *HashSet) bool {
	if other == nil {
		return false
	}
	if set.Len() != other.Len() {
		return false
	}
	for key := range set.m {
		if !other.Contains(key) {
			return false
		}
	}
	return true
}
