package services

import (
	"go.mongodb.org/mongo-driver/bson"
	"interview/models"
	"interview/router/request"
)

type Study struct {
	ServicesBase
}

func (sf *Study) BackendCourseList(filter bson.M, skipCount, limitCount int64) ([]models.RecommendCourse, int64, error) {
	papers := make([]models.RecommendCourse, 0)
	paperCount, err := sf.RecommendCourseModel().Where(filter).Count()
	if err != nil {
		return papers, paperCount, err
	}

	err = sf.RecommendCourseModel().Where(filter).Skip(skipCount).Limit(limitCount).Sort("-_id").Find(&papers)
	if err != nil {
		return papers, paperCount, err
	}
	return papers, paperCount, err
}

func (sf *Study) CourseList(filter bson.M, skipCount, limitCount int64) ([]models.RecommendCourse, int64, error) {
	filter["status"] = 1
	papers := make([]models.RecommendCourse, 0)
	paperCount, err := sf.RecommendCourseModel().Where(filter).Count()
	if err != nil {
		return papers, paperCount, err
	}

	err = sf.RecommendCourseModel().Where(filter).Skip(skipCount).Limit(limitCount).Sort("-sort", "-updated_time").Find(&papers)
	if err != nil {
		return papers, paperCount, err
	}
	return papers, paperCount, err
}

func (sf *Study) CourseEdit(param request.RecommendCourseEditRequest) (*models.RecommendCourse, error) {
	course := new(models.RecommendCourse)
	var err error
	if param.Id != "" {
		err = sf.RecommendCourseModel().Where(bson.M{"_id": sf.ObjectID(param.Id)}).Take(&course)
		if err != nil {
			return course, err
		}
	}

	course.Status = param.Status
	course.Title = param.Title
	course.CourseType = param.CourseType
	course.Province = param.Province
	course.ProvinceCode = param.ProvinceCode
	course.Year = param.Year
	course.ExamCategory = param.ExamCategory
	course.ExamChildCategory = param.ExamChildCategory
	course.OriginData = param.OriginData
	course.Sort = param.Sort
	course.JobTag = param.JobTag
	if param.Id != "" {
		_, err = sf.RecommendCourseModel().Where(bson.M{"_id": course.Id}).Update(course)
	} else {
		_, err = sf.RecommendCourseModel().Create(course)
	}

	return course, err
}

func (sf *Study) BackendDataPackList(filter bson.M, skipCount, limitCount int64) ([]models.RecommendDataPack, int64, error) {
	papers := make([]models.RecommendDataPack, 0)
	paperCount, err := sf.RecommendDataPackModel().Where(filter).Count()
	if err != nil {
		return papers, paperCount, err
	}

	err = sf.RecommendDataPackModel().Where(filter).Skip(skipCount).Limit(limitCount).Sort("-_id").Find(&papers)
	if err != nil {
		return papers, paperCount, err
	}
	return papers, paperCount, err
}

func (sf *Study) DataPackList(filter bson.M, skipCount, limitCount int64) ([]models.RecommendDataPack, int64, error) {
	filter["status"] = 1
	papers := make([]models.RecommendDataPack, 0)
	paperCount, err := sf.RecommendDataPackModel().Where(filter).Count()
	if err != nil {
		return papers, paperCount, err
	}

	err = sf.RecommendDataPackModel().Where(filter).Skip(skipCount).Limit(limitCount).Sort("-sort", "-updated_time").Find(&papers)
	if err != nil {
		return papers, paperCount, err
	}
	return papers, paperCount, err
}

func (sf *Study) DataPackEdit(param request.RecommendDataPackEditRequest) (*models.RecommendDataPack, error) {
	dataPack := new(models.RecommendDataPack)
	var err error
	if param.Id != "" {
		err = sf.RecommendDataPackModel().Where(bson.M{"_id": sf.ObjectID(param.Id)}).Take(&dataPack)
		if err != nil {
			return dataPack, err
		}
	}

	dataPack.Status = param.Status
	dataPack.Title = param.Title
	dataPack.CoverImg = param.CoverImg
	dataPack.ResUrl = param.ResUrl
	dataPack.ResId = param.ResId
	dataPack.ResTitle = param.ResTitle
	dataPack.JobTag = param.JobTag
	dataPack.ExamCategory = param.ExamCategory
	dataPack.ExamChildCategory = param.ExamChildCategory
	dataPack.Sort = param.Sort
	if param.Id != "" {
		_, err = sf.RecommendDataPackModel().Where(bson.M{"_id": dataPack.Id}).Update(dataPack)
	} else {
		_, err = sf.RecommendDataPackModel().Create(dataPack)
	}

	return dataPack, err
}
