## LRU算法(最近最少使用)
- 使用hashMap+双向链表，链表自己实现
- 双向链表使用头结点和尾结点，这样获取头尾节点的时间复杂度是O(1)

## 定义lru
```go
// 链表节点的结构
type MyListNode struct {
	key,val int
	pre,next *MyListNode
}
// 缓存结构
type LRUCace struct{
    cap int // 缓存大小
    size int // 真实大小
    head,tail *MyListNode // 头尾节点
    cache map[int]*MyListNode // map存缓存
}
```
## 实现
需要实现的方法
```go
// 初始化节点的方法
// 初始化节点的方法
func initNode(key,val int) *MyListNode{
	return &MyListNode{
		key:key,
		val:val,
	}
}
// 初始化缓存结构
func Constructor(capacity int) LRUCache {
	l := LRUCache{
		cap:capacity,
		head:initNode(0,0),
		tail:initNode(0,0),
		cache:map[int]*MyListNode{},
	}
	l.head.next = l.tail
	l.tail.next = l.head
	return l
}


// 获取节点
// 节点不在，返回-1，节点存在，返回节点值并移动节点到头部
func (this *LRUCache) Get(key int) int {
	if _,ok := this.cache[key];!ok {
		return -1
	}else{
		node := this.cache[key]
		this.moveToHead(node)
		return node.val
	}
}

// 插入节点 节点存在，修改值 移动到头部 不存在，新建一个节点插入头部，长度超了 移除最后一个元素，
// 记得最后一个元素是尾结点的pre
func (this *LRUCache) Put(key int, value int)  {
	if _,ok := this.cache[key];!ok {
		node := initNode(key,value)
		this.cache[key] = node
		this.addToHead(node)
		this.size++
		if this.size > this.cap {
			rem := this.removeTail()
			delete(this.cache,rem.key)
			this.size--
		}
	}else{
		node := this.cache[key]
		node.val = value
		this.moveToHead(node)
	}
}
```
接下来是四个辅助方法
```go
// 把链表中一个节点移动到头部
func (this *LRUCache) moveToHead (node *MyListNode){
    // 删除当前节点
	this.removeNode(node)
    // 添加到头部
	this.addToHead(node)
}
// 删除节点
func (this *LRUCache) removeNode (node *MyListNode){
    // 修改前后节点的指针
	node.pre.next = node.next
	node.next.pre  = node.pre
}
// 添加一个节点到头部
func (this *LRUCache) addToHead (node *MyListNode){
    // 改变当前节点的前后指针
	node.pre = this.head
	node.next = this.head.next
    // 改变前后节点的pre\next指针
	this.head.next.pre = node
	this.head.next = node
}
// 删除最后一个节点。链表最后一个节点是尾结点，是无效节点
// 最后一个有效节点是tail.pre
func (this *LRUCache) removeTail() *MyListNode {
	node := this.tail.pre
	this.removeNode(node)
	return node
}
```