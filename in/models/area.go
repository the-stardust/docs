package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChildArea struct {
	Title      string      `json:"title" bson:"title"`
	Acronym    string      `json:"acronym" bson:"acronym"`
	ChildAreas []ChildArea `json:"childrenLabels" bson:"childrenLabels"`
}
type Areas struct {
	Id         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title      string             `json:"title" bson:"title"`
	Acronym    string             `json:"acronym" bson:"acronym"`
	ChildAreas []ChildArea        `json:"childrenLabels" bson:"childrenLabels"`
}

type Location struct {
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
}

type TreeNode struct {
	Title string     `json:"title"`
	Child []TreeNode `json:"child"`
}

func (sf *TreeNode) BuildTree(data []Location) TreeNode {
	// 创建根节点
	root := TreeNode{Title: "root", Child: []TreeNode{}}
	currentNode := &root

	for _, loc := range data {
		// 递归添加节点
		sf.AddNode(currentNode, loc)
	}

	// 返回根节点的子节点作为最终结果
	return root
}

func (sf *TreeNode) AddNode(node *TreeNode, loc Location) {
	if loc.Province == "" {
		return
	}

	// 查找是否已经存在相同的节点
	for i := range node.Child {
		if node.Child[i].Title == loc.Province {
			sf.AddCityAndDistrict(&node.Child[i], loc)
			return
		}
	}

	// 不存在相同的节点，创建一个新节点并添加
	newNode := TreeNode{Title: loc.Province, Child: []TreeNode{}}
	node.Child = append(node.Child, newNode)
	sf.AddCityAndDistrict(&node.Child[len(node.Child)-1], loc)
}

func (sf *TreeNode) AddCityAndDistrict(node *TreeNode, loc Location) {
	if loc.City == "" {
		return
	}

	// 查找是否已经存在相同的城市节点
	for i := range node.Child {
		if node.Child[i].Title == loc.City {
			sf.AddDistrict(&node.Child[i], loc)
			return
		}
	}

	// 不存在相同的城市节点，创建一个新城市节点并添加
	newNode := TreeNode{Title: loc.City, Child: []TreeNode{}}
	node.Child = append(node.Child, newNode)
	sf.AddDistrict(&node.Child[len(node.Child)-1], loc)
}

func (sf *TreeNode) AddDistrict(node *TreeNode, loc Location) {
	if loc.District == "" {
		return
	}

	// 查找是否已经存在相同的区域节点
	for i := range node.Child {
		if node.Child[i].Title == loc.District {
			return // 区域已存在，不再添加
		}
	}

	// 不存在相同的区域节点，创建一个新区域节点并添加
	districtNode := TreeNode{Title: loc.District, Child: []TreeNode{}}
	node.Child = append(node.Child, districtNode)
}
