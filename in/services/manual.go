package services

import (
	"encoding/json"
	"fmt"
	"interview/common/rediskey"
	"interview/helper"
	"interview/models"
	"strings"
)

type Manual struct {
	ServicesBase
}

func NewManualService() *Manual {
	return &Manual{}
}

func (sf *Manual) GetCategoryUsersPermissions(uid string, redisKey string) map[string]map[string][]string {
	userPermissionSli := make(map[string]map[string][]string, 0)
	userPermissionStr, _ := helper.BankRedisHGet(redisKey, uid)
	if userPermissionStr != "" {
		_ = json.Unmarshal([]byte(userPermissionStr), &userPermissionSli)
	}
	return userPermissionSli
}

func (sf *Manual) GetUserCategoryPermissions(uid string, examCategory []models.ExamCategory) []models.ExamCategory {
	resp := make([]models.ExamCategory, 0)

	permissionMap := sf.GetCategoryUsersPermissions(uid, string(rediskey.CategoryPermissionInterview))
	fmt.Println(permissionMap)
	if len(permissionMap) != 0 {
		resp = make([]models.ExamCategory, 0)
		for _, category := range examCategory {
			for s, _ := range permissionMap {
				cateArr := strings.Split(s, "/")
				if cateArr[0] == category.Title {
					// 说明有child category
					if len(cateArr) > 1 {
						newCate := models.ExamCategory{
							Title:         category.Title,
							Id:            category.Id,
							ChildCategory: make([]models.ExamCategoryItem, 0),
						}
						index := -1
						for i, i2 := range resp {
							if i2.Title == category.Title && len(i2.ChildCategory) > 0 {
								newCate = i2
								index = i
								break
							}
						}

						for _, item := range category.ChildCategory {
							if item.Title == cateArr[1] {
								newCate.ChildCategory = append(newCate.ChildCategory, item)
								break
							}
						}
						if index >= 0 {
							resp[index] = newCate
						} else {
							resp = append(resp, category)
						}
					} else {
						resp = append(resp, category)
					}
					break
				}
			}
		}
	} else { // 走到else 说明没限制权限
		resp = examCategory
	}
	return resp
}

func (sf *Manual) GetUserKeypointsPermissions(uid, examCategory, examChildCategory string, oldResp []models.QuestionCategoryItem) []models.QuestionCategoryItem {
	key := examCategory
	if examChildCategory != "" {
		key += "/" + examChildCategory
	}
	resp := make([]models.QuestionCategoryItem, 0)

	permissionMap := sf.GetCategoryUsersPermissions(uid, string(rediskey.CategoryPermissionInterview))
	if len(permissionMap) != 0 {
		if categoryPermission, ok := permissionMap[key]; ok {
			resp = make([]models.QuestionCategoryItem, 0)
			for keypoints := range categoryPermission {
				for _, s := range oldResp {
					if s.Title == keypoints {
						resp = append(resp, s)
						break
					}
				}
			}
		} else {
			resp = make([]models.QuestionCategoryItem, 0)
		}
	} else { // 走到else 说明没限制权限
		resp = oldResp
	}
	return resp
}
