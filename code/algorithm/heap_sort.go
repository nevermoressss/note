package algorithm

type heap []int

// 交换
func (h heap) swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// 小堆
func (h heap) less(i, j int) bool {
	return h[i] < h[j]
}

func (h heap) len() int {
	return len(h)
}

// 用于插入新节点时候，新节点到合适的位置
func (h heap) up(i int) {
	for {
		f := (i - 1) / 2 // 父亲结点，和父节点进行比较，看是否需要交换
		if i == f || h.less(f, i) {
			break
		}
		h.swap(f, i)
		i = f
	}
}

func (h *heap) Push(x int) {
	*h = append(*h, x)
	h.up(len(*h) - 1)
}

// 下降节点
// 每次下层和子节点中较小（大）的做比较
func (h heap) down(i int) {
	for {
		l := 2*i + 1 // 左孩子
		if l >= len(h) {
			break // i已经是叶子结点了
		}
		j := l
		if r := l + 1; r < len(h) && h.less(r, l) {
			j = r // 右孩子
		}
		if h.less(i, j) {
			break // 如果父结点比孩子结点小，则不交换
		}
		h.swap(i, j) // 交换父结点和子结点
		i = j        //继续向下比较
	}
}

// 删除堆中位置为i的元素
// 返回被删元素的值
func (h *heap) Remove(i int) (int, bool) {
	if i < 0 || i > len(*h)-1 {
		return 0, false
	}
	n := len(*h) - 1
	h.swap(i, n) // 用最后的元素值替换被删除元素
	// 删除最后的元素
	x := (*h)[n]
	*h = (*h)[0:n]
	// 如果当前元素大于父结点，向下筛选
	if (*h)[i] > (*h)[(i-1)/2] {
		h.down(i)
	} else { // 当前元素小于父结点，向上筛选
		h.up(i)
	}
	return x, true
}

// 弹出堆顶的元素，并返回其值
func (h *heap) Pop() int {
	n := len(*h) - 1
	h.swap(0, n)
	x := (*h)[n]
	*h = (*h)[0:n]
	h.down(0)
	return x
}

func (h heap) Init() {
	n := len(h)
	// i > n/2-1 的结点为叶子结点本身已经是堆了
	for i := n/2 - 1; i >= 0; i-- {
		h.down(i)
	}
}
