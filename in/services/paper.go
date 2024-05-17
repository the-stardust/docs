package services

import (
	"go.mongodb.org/mongo-driver/bson"
	"interview/models"
)

type Paper struct {
	ServicesBase
}

func (sf *Paper) PaperList(filter bson.M, skipCount, limitCount int64) ([]models.Paper, int64, error) {
	papers := make([]models.Paper, 0)
	paperCount, err := sf.DB().Collection("paper").Where(filter).Count()
	if err != nil {
		return papers, paperCount, err
	}

	err = sf.DB().Collection("paper").Where(filter).Skip(skipCount).Limit(limitCount).Sort("-updated_time").Find(&papers)
	if err != nil {
		return papers, paperCount, err
	}
	return papers, paperCount, err
}

func (sf *Paper) PaperInfo(filter bson.M) (models.Paper, error) {
	var paper models.Paper
	err := sf.DB().Collection("paper").Where(filter).Take(&paper)
	return paper, err
}

//func (sf *Paper)SavePaper(param struct)  {
//
//
//}
